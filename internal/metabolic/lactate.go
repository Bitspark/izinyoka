package metabolic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// PartialComputation represents an incomplete computation that is stored
// for potential reuse or completion
type PartialComputation struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Data          map[string]interface{} `json:"data"`
	Progress      float64                `json:"progress"`
	Intermediate  map[string]interface{} `json:"intermediate"`
	CreatedAt     time.Time              `json:"createdAt"`
	LastUpdatedAt time.Time              `json:"lastUpdatedAt"`
	ExpiresAt     time.Time              `json:"expiresAt"`
	Origin        string                 `json:"origin,omitempty"`
	Priority      int                    `json:"priority"`
	Recoverable   bool                   `json:"recoverable"`
}

// LactateConfig contains configuration for the lactate cycle
type LactateConfig struct {
	MaxPartialComputations int           `json:"maxPartialComputations"`
	DefaultTTL             time.Duration `json:"defaultTTL"`
	CleanupInterval        time.Duration `json:"cleanupInterval"`
	PrioritizeRecoverable  bool          `json:"prioritizeRecoverable"`
	MaxStorageBytes        int64         `json:"maxStorageBytes"`
}

// LactateCycle implements storage and reuse of partial computations
type LactateCycle struct {
	config      LactateConfig
	partials    map[string]*PartialComputation
	byType      map[string]map[string]struct{}
	byPriority  map[int]map[string]struct{}
	mutex       sync.RWMutex
	ctx         context.Context
	cancelFunc  context.CancelFunc
	storageSize int64
	logger      *logrus.Logger
}

// NewLactateCycle creates a new instance of the lactate cycle
func NewLactateCycle(ctx context.Context, config LactateConfig) *LactateCycle {
	if config.MaxPartialComputations <= 0 {
		config.MaxPartialComputations = 1000
	}
	if config.DefaultTTL <= 0 {
		config.DefaultTTL = 24 * time.Hour
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 10 * time.Minute
	}
	if config.MaxStorageBytes <= 0 {
		config.MaxStorageBytes = 100 * 1024 * 1024 // 100 MB
	}

	cycleCtx, cancelFunc := context.WithCancel(ctx)

	lc := &LactateCycle{
		config:      config,
		partials:    make(map[string]*PartialComputation),
		byType:      make(map[string]map[string]struct{}),
		byPriority:  make(map[int]map[string]struct{}),
		ctx:         cycleCtx,
		cancelFunc:  cancelFunc,
		storageSize: 0,
		logger:      logrus.New(),
	}

	// Start cleanup goroutine
	go lc.runCleanup()

	return lc
}

// StorePartial adds or updates a partial computation
func (lc *LactateCycle) StorePartial(partial *PartialComputation) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	// Generate ID if not provided
	if partial.ID == "" {
		partial.ID = fmt.Sprintf("pc-%d", time.Now().UnixNano())
	}

	// Check if we're updating an existing computation
	existingSize := int64(0)
	if existing, exists := lc.partials[partial.ID]; exists {
		// Calculate existing size for storage tracking
		existingSize = approximateSize(existing)

		// Remove from indices
		lc.removeFromIndices(existing)
	}

	// Ensure required fields are set
	now := time.Now()
	if partial.CreatedAt.IsZero() {
		partial.CreatedAt = now
	}
	partial.LastUpdatedAt = now

	// Set expiration if not provided
	if partial.ExpiresAt.IsZero() {
		partial.ExpiresAt = now.Add(lc.config.DefaultTTL)
	}

	// Initialize intermediate data if nil
	if partial.Intermediate == nil {
		partial.Intermediate = make(map[string]interface{})
	}

	// Track computation size
	newSize := approximateSize(partial)
	lc.storageSize = lc.storageSize - existingSize + newSize

	// Check if we're over capacity
	if lc.storageSize > lc.config.MaxStorageBytes || len(lc.partials) >= lc.config.MaxPartialComputations {
		// Try to free up space
		lc.evictPartials(1 + len(lc.partials) - lc.config.MaxPartialComputations)

		// Check if we're still over capacity
		if lc.storageSize > lc.config.MaxStorageBytes || len(lc.partials) >= lc.config.MaxPartialComputations {
			return fmt.Errorf("storage capacity exceeded, partial computation rejected")
		}
	}

	// Store partial computation
	lc.partials[partial.ID] = partial

	// Update indices
	lc.addToIndices(partial)

	lc.logger.WithFields(logrus.Fields{
		"id":       partial.ID,
		"type":     partial.Type,
		"progress": partial.Progress,
	}).Debug("Stored partial computation")

	return nil
}

// FindPartial retrieves a specific partial computation by ID
func (lc *LactateCycle) FindPartial(id string) (*PartialComputation, error) {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()

	partial, exists := lc.partials[id]
	if !exists {
		return nil, fmt.Errorf("partial computation with ID %s not found", id)
	}

	// Check if it has expired
	if time.Now().After(partial.ExpiresAt) {
		return nil, fmt.Errorf("partial computation with ID %s has expired", id)
	}

	return partial, nil
}

// QueryPartials finds partial computations matching the specified criteria
func (lc *LactateCycle) QueryPartials(queryType string, data map[string]interface{}, minProgress float64) ([]*PartialComputation, error) {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()

	results := make([]*PartialComputation, 0)

	// Start with computations of the requested type
	candidates := make(map[string]*PartialComputation)
	if queryType != "" {
		if typeIndex, exists := lc.byType[queryType]; exists {
			for id := range typeIndex {
				if partial, ok := lc.partials[id]; ok {
					candidates[id] = partial
				}
			}
		}
	} else {
		// Use all partials if no type specified
		for id, partial := range lc.partials {
			candidates[id] = partial
		}
	}

	// Filter by provided data fields
	if len(data) > 0 {
		for id, partial := range candidates {
			if !matchesData(partial.Data, data) {
				delete(candidates, id)
			}
		}
	}

	// Filter by minimum progress
	if minProgress > 0 {
		for id, partial := range candidates {
			if partial.Progress < minProgress {
				delete(candidates, id)
			}
		}
	}

	// Filter out expired computations
	now := time.Now()
	for id, partial := range candidates {
		if now.After(partial.ExpiresAt) {
			delete(candidates, id)
		}
	}

	// Convert to slice
	for _, partial := range candidates {
		results = append(results, partial)
	}

	return results, nil
}

// UpdateProgress updates the progress value of a partial computation
func (lc *LactateCycle) UpdateProgress(id string, progress float64, intermediate map[string]interface{}) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	partial, exists := lc.partials[id]
	if !exists {
		return fmt.Errorf("partial computation with ID %s not found", id)
	}

	// Check if it has expired
	if time.Now().After(partial.ExpiresAt) {
		return fmt.Errorf("partial computation with ID %s has expired", id)
	}

	// Calculate existing size for storage tracking
	existingSize := approximateSize(partial)

	// Update progress
	partial.Progress = progress
	partial.LastUpdatedAt = time.Now()

	// Update intermediate data if provided
	if intermediate != nil {
		for k, v := range intermediate {
			partial.Intermediate[k] = v
		}
	}

	// Update storage tracking
	newSize := approximateSize(partial)
	lc.storageSize = lc.storageSize - existingSize + newSize

	lc.logger.WithFields(logrus.Fields{
		"id":       partial.ID,
		"progress": partial.Progress,
	}).Debug("Updated partial computation progress")

	return nil
}

// DeletePartial removes a partial computation
func (lc *LactateCycle) DeletePartial(id string) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	partial, exists := lc.partials[id]
	if !exists {
		return fmt.Errorf("partial computation with ID %s not found", id)
	}

	// Remove from indices
	lc.removeFromIndices(partial)

	// Update storage size
	lc.storageSize -= approximateSize(partial)

	// Remove from storage
	delete(lc.partials, id)

	lc.logger.WithFields(logrus.Fields{
		"id": partial.ID,
	}).Debug("Deleted partial computation")

	return nil
}

// ExtendTTL extends the time-to-live for a partial computation
func (lc *LactateCycle) ExtendTTL(id string, extension time.Duration) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	partial, exists := lc.partials[id]
	if !exists {
		return fmt.Errorf("partial computation with ID %s not found", id)
	}

	// Check if it has already expired
	if time.Now().After(partial.ExpiresAt) {
		return fmt.Errorf("partial computation with ID %s has already expired", id)
	}

	// Extend the expiration time
	partial.ExpiresAt = partial.ExpiresAt.Add(extension)

	// Update last modified time
	partial.LastUpdatedAt = time.Now()

	lc.logger.WithFields(logrus.Fields{
		"id":        partial.ID,
		"extension": extension.String(),
		"expiresAt": partial.ExpiresAt,
	}).Debug("Extended partial computation TTL")

	return nil
}

// GetStorageStats returns information about the lactate cycle's storage utilization
func (lc *LactateCycle) GetStorageStats() map[string]interface{} {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()

	// Count active (non-expired) partials
	now := time.Now()
	activeCount := 0
	typeDistribution := make(map[string]int)

	for _, partial := range lc.partials {
		if !now.After(partial.ExpiresAt) {
			activeCount++
			typeDistribution[partial.Type] = typeDistribution[partial.Type] + 1
		}
	}

	return map[string]interface{}{
		"totalPartials":      len(lc.partials),
		"activePartials":     activeCount,
		"storageBytes":       lc.storageSize,
		"maxStorageBytes":    lc.config.MaxStorageBytes,
		"storageUtilization": float64(lc.storageSize) / float64(lc.config.MaxStorageBytes),
		"typeDistribution":   typeDistribution,
	}
}

// Shutdown cleanly stops the lactate cycle
func (lc *LactateCycle) Shutdown() {
	lc.cancelFunc()
}

// Helper methods

// addToIndices adds a partial computation to all indices
func (lc *LactateCycle) addToIndices(partial *PartialComputation) {
	// Add to type index
	if _, exists := lc.byType[partial.Type]; !exists {
		lc.byType[partial.Type] = make(map[string]struct{})
	}
	lc.byType[partial.Type][partial.ID] = struct{}{}

	// Add to priority index
	if _, exists := lc.byPriority[partial.Priority]; !exists {
		lc.byPriority[partial.Priority] = make(map[string]struct{})
	}
	lc.byPriority[partial.Priority][partial.ID] = struct{}{}
}

// removeFromIndices removes a partial computation from all indices
func (lc *LactateCycle) removeFromIndices(partial *PartialComputation) {
	// Remove from type index
	if typeIndex, exists := lc.byType[partial.Type]; exists {
		delete(typeIndex, partial.ID)
		if len(typeIndex) == 0 {
			delete(lc.byType, partial.Type)
		}
	}

	// Remove from priority index
	if priorityIndex, exists := lc.byPriority[partial.Priority]; exists {
		delete(priorityIndex, partial.ID)
		if len(priorityIndex) == 0 {
			delete(lc.byPriority, partial.Priority)
		}
	}
}

// runCleanup periodically removes expired partials
func (lc *LactateCycle) runCleanup() {
	ticker := time.NewTicker(lc.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lc.ctx.Done():
			return
		case <-ticker.C:
			lc.cleanupExpired()
		}
	}
}

// cleanupExpired removes all expired partial computations
func (lc *LactateCycle) cleanupExpired() {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	now := time.Now()
	expiredCount := 0
	freedBytes := int64(0)

	// Find expired computations
	for id, partial := range lc.partials {
		if now.After(partial.ExpiresAt) {
			// Track stats
			expiredCount++
			freedBytes += approximateSize(partial)

			// Remove from indices
			lc.removeFromIndices(partial)

			// Remove from storage
			delete(lc.partials, id)
		}
	}

	// Update storage size
	lc.storageSize -= freedBytes

	if expiredCount > 0 {
		lc.logger.WithFields(logrus.Fields{
			"count":      expiredCount,
			"freedBytes": freedBytes,
		}).Info("Cleaned up expired partial computations")
	}
}

// evictPartials removes the lowest priority partials to free up space
func (lc *LactateCycle) evictPartials(countToEvict int) {
	if countToEvict <= 0 {
		return
	}

	// Find priorities in ascending order
	priorities := make([]int, 0, len(lc.byPriority))
	for priority := range lc.byPriority {
		priorities = append(priorities, priority)
	}
	sortAscending(priorities)

	// Evict partials starting from lowest priority
	evictedCount := 0
	freedBytes := int64(0)

	for _, priority := range priorities {
		if evictedCount >= countToEvict {
			break
		}

		priorityIndex := lc.byPriority[priority]

		// For recoverable vs non-recoverable, prioritize keeping recoverable if configured
		if lc.config.PrioritizeRecoverable {
			// First try to evict non-recoverable partials
			for id := range priorityIndex {
				if evictedCount >= countToEvict {
					break
				}

				partial := lc.partials[id]
				if !partial.Recoverable {
					// Track stats
					evictedCount++
					freedBytes += approximateSize(partial)

					// Remove from indices
					lc.removeFromIndices(partial)

					// Remove from storage
					delete(lc.partials, id)
					delete(priorityIndex, id)
				}
			}
		}

		// If we still need to evict more, then evict the rest
		for id := range priorityIndex {
			if evictedCount >= countToEvict {
				break
			}

			partial := lc.partials[id]

			// Track stats
			evictedCount++
			freedBytes += approximateSize(partial)

			// Remove from indices
			lc.removeFromIndices(partial)

			// Remove from storage
			delete(lc.partials, id)
			delete(priorityIndex, id)
		}

		// If priority index is now empty, remove it
		if len(priorityIndex) == 0 {
			delete(lc.byPriority, priority)
		}
	}

	// Update storage size
	lc.storageSize -= freedBytes

	if evictedCount > 0 {
		lc.logger.WithFields(logrus.Fields{
			"count":      evictedCount,
			"freedBytes": freedBytes,
		}).Info("Evicted partial computations due to storage constraints")
	}
}

// Helper functions

// matchesData checks if a data map matches a query data subset
func matchesData(data, query map[string]interface{}) bool {
	for k, v := range query {
		dataValue, exists := data[k]
		if !exists || dataValue != v {
			return false
		}
	}
	return true
}

// approximateSize estimates the memory size of a partial computation
func approximateSize(partial *PartialComputation) int64 {
	// This is a simplified approximation that could be made more accurate
	// Start with a base size for the struct
	size := int64(200)

	// Add size for data
	size += int64(len(partial.ID) * 2)
	size += int64(len(partial.Type) * 2)
	size += int64(len(partial.Origin) * 2)

	// Add size for maps
	if partial.Data != nil {
		size += estimateMapSize(partial.Data)
	}

	if partial.Intermediate != nil {
		size += estimateMapSize(partial.Intermediate)
	}

	return size
}

// estimateMapSize approximates the size of a map
func estimateMapSize(m map[string]interface{}) int64 {
	// Base size for map structure
	size := int64(48)

	// Add size for each key and estimate value size
	for k, v := range m {
		size += int64(len(k) * 2) // UTF-16 chars
		size += estimateValueSize(v)
	}

	return size
}

// estimateValueSize approximates the size of a value
func estimateValueSize(v interface{}) int64 {
	if v == nil {
		return 0
	}

	switch val := v.(type) {
	case string:
		return int64(len(val) * 2) // UTF-16 chars
	case int, int32, float32, bool:
		return 4
	case int64, float64:
		return 8
	case []interface{}:
		size := int64(24) // Base array size
		for _, item := range val {
			size += estimateValueSize(item)
		}
		return size
	case map[string]interface{}:
		return estimateMapSize(val)
	default:
		return 16 // Default approximation
	}
}

// sortAscending sorts an int slice in ascending order
func sortAscending(slice []int) {
	// Simple insertion sort for small slices
	for i := 1; i < len(slice); i++ {
		j := i
		for j > 0 && slice[j-1] > slice[j] {
			slice[j], slice[j-1] = slice[j-1], slice[j]
			j--
		}
	}
}
