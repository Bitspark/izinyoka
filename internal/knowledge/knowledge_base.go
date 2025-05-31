package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// KnowledgeBaseConfig contains configuration for the knowledge base
type KnowledgeBaseConfig struct {
	StoragePath   string
	BackupEnabled bool
	BackupPath    string
	MaxItems      int
}

// KnowledgeBase is the central repository for domain knowledge
type KnowledgeBase struct {
	config KnowledgeBaseConfig
	items  map[string]KnowledgeItem
	mutex  sync.RWMutex
	logger *logrus.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// NewKnowledgeBase creates a new knowledge base
func NewKnowledgeBase(ctx context.Context, config KnowledgeBaseConfig) (*KnowledgeBase, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create knowledge storage directory: %w", err)
	}

	// Create backup directory if enabled
	if config.BackupEnabled {
		if err := os.MkdirAll(config.BackupPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create knowledge backup directory: %w", err)
		}
	}

	// Create context with cancellation
	kbCtx, cancelFunc := context.WithCancel(ctx)

	kb := &KnowledgeBase{
		config: config,
		items:  make(map[string]KnowledgeItem),
		logger: logrus.New(),
		ctx:    kbCtx,
		cancel: cancelFunc,
	}

	// Load existing items
	if err := kb.loadItems(); err != nil {
		kb.logger.WithError(err).Warn("Failed to load existing knowledge items")
	}

	// Start background tasks if enabled
	if config.BackupEnabled {
		go kb.periodicBackup()
	}

	return kb, nil
}

// Store adds or updates a knowledge item
func (kb *KnowledgeBase) Store(item KnowledgeItem) error {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	// Generate ID if not provided
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	// Set timestamps if not provided
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.LastModified = now

	// Check if we're at capacity
	if len(kb.items) >= kb.config.MaxItems && kb.items[item.ID] == (KnowledgeItem{}) {
		return errors.New("knowledge base at maximum capacity")
	}

	// Store item
	kb.items[item.ID] = item

	// Save to disk
	if err := kb.saveItem(item); err != nil {
		kb.logger.WithError(err).Error("Failed to save knowledge item")
		return fmt.Errorf("failed to save knowledge item: %w", err)
	}

	return nil
}

// Retrieve gets a knowledge item by ID
func (kb *KnowledgeBase) Retrieve(id string) (KnowledgeItem, error) {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	item, exists := kb.items[id]
	if !exists {
		return KnowledgeItem{}, fmt.Errorf("knowledge item with ID %s not found", id)
	}

	return item, nil
}

// Delete removes a knowledge item
func (kb *KnowledgeBase) Delete(id string) error {
	kb.mutex.Lock()
	defer kb.mutex.Unlock()

	// Check if item exists
	if _, exists := kb.items[id]; !exists {
		return fmt.Errorf("knowledge item with ID %s not found", id)
	}

	// Remove from memory
	delete(kb.items, id)

	// Remove from disk
	filePath := filepath.Join(kb.config.StoragePath, id+".json")
	if err := os.Remove(filePath); err != nil {
		if !os.IsNotExist(err) {
			kb.logger.WithError(err).Error("Failed to delete knowledge item file")
			return fmt.Errorf("failed to delete knowledge item file: %w", err)
		}
	}

	return nil
}

// Query searches for knowledge items matching the criteria
func (kb *KnowledgeBase) Query(query KnowledgeQuery) ([]KnowledgeItem, error) {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	results := make([]KnowledgeItem, 0)

	// If ID is specified, return that specific item
	if query.ID != "" {
		item, exists := kb.items[query.ID]
		if exists {
			results = append(results, item)
		}
		return results, nil
	}

	// Filter items by criteria
	for _, item := range kb.items {
		if matchesQuery(item, query) {
			results = append(results, item)
		}
	}

	// Apply limit if specified
	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

// Count returns the number of items in the knowledge base
func (kb *KnowledgeBase) Count() int {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	return len(kb.items)
}

// GetStats returns statistics about the knowledge base
func (kb *KnowledgeBase) GetStats() map[string]interface{} {
	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	// Count by type
	typeCount := make(map[string]int)
	totalConfidence := 0.0
	oldestTime := time.Now()
	newestTime := time.Time{}

	for _, item := range kb.items {
		typeCount[item.Type]++
		totalConfidence += item.Confidence

		if item.CreatedAt.Before(oldestTime) {
			oldestTime = item.CreatedAt
		}
		if item.LastModified.After(newestTime) {
			newestTime = item.LastModified
		}
	}

	avgConfidence := 0.0
	if len(kb.items) > 0 {
		avgConfidence = totalConfidence / float64(len(kb.items))
	}

	return map[string]interface{}{
		"totalItems":    len(kb.items),
		"typeCount":     typeCount,
		"avgConfidence": avgConfidence,
		"oldest":        oldestTime,
		"newest":        newestTime,
		"maxItems":      kb.config.MaxItems,
		"utilization":   float64(len(kb.items)) / float64(kb.config.MaxItems),
	}
}

// CreateBackup creates a backup of the knowledge base
func (kb *KnowledgeBase) CreateBackup() error {
	if !kb.config.BackupEnabled {
		return errors.New("backups are not enabled")
	}

	kb.mutex.RLock()
	defer kb.mutex.RUnlock()

	// Create backup directory with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(kb.config.BackupPath, fmt.Sprintf("backup-%s", timestamp))
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Write metadata file
	metadataFile := filepath.Join(backupDir, "metadata.json")
	metadata := map[string]interface{}{
		"timestamp":  timestamp,
		"itemCount":  len(kb.items),
		"maxItems":   kb.config.MaxItems,
		"backupPath": backupDir,
	}
	metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize backup metadata: %w", err)
	}
	if err := os.WriteFile(metadataFile, metadataBytes, 0644); err != nil {
		return fmt.Errorf("failed to write backup metadata: %w", err)
	}

	// Copy all items
	for id, item := range kb.items {
		itemPath := filepath.Join(backupDir, id+".json")
		itemBytes, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			kb.logger.WithError(err).WithField("id", id).Error("Failed to serialize item for backup")
			continue
		}
		if err := os.WriteFile(itemPath, itemBytes, 0644); err != nil {
			kb.logger.WithError(err).WithField("id", id).Error("Failed to write item for backup")
			continue
		}
	}

	kb.logger.WithFields(logrus.Fields{
		"backupPath": backupDir,
		"itemCount":  len(kb.items),
	}).Info("Knowledge base backup created")

	return nil
}

// Shutdown cleanly shuts down the knowledge base
func (kb *KnowledgeBase) Shutdown() error {
	kb.logger.Info("Shutting down knowledge base")

	// Cancel context to stop background tasks
	kb.cancel()

	// Create final backup if enabled
	if kb.config.BackupEnabled {
		if err := kb.CreateBackup(); err != nil {
			kb.logger.WithError(err).Error("Failed to create final backup during shutdown")
		}
	}

	return nil
}

// loadItems loads knowledge items from disk
func (kb *KnowledgeBase) loadItems() error {
	// Get all JSON files in the storage directory
	pattern := filepath.Join(kb.config.StoragePath, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list knowledge item files: %w", err)
	}

	for _, file := range files {
		// Read file
		data, err := os.ReadFile(file)
		if err != nil {
			kb.logger.WithError(err).WithField("file", file).Error("Failed to read knowledge item file")
			continue
		}

		// Parse JSON
		var item KnowledgeItem
		if err := json.Unmarshal(data, &item); err != nil {
			kb.logger.WithError(err).WithField("file", file).Error("Failed to parse knowledge item file")
			continue
		}

		// Add to in-memory store
		kb.items[item.ID] = item
	}

	kb.logger.WithField("count", len(kb.items)).Info("Loaded knowledge items from disk")
	return nil
}

// saveItem saves a knowledge item to disk
func (kb *KnowledgeBase) saveItem(item KnowledgeItem) error {
	// Serialize item
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize knowledge item: %w", err)
	}

	// Write to file
	filePath := filepath.Join(kb.config.StoragePath, item.ID+".json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write knowledge item file: %w", err)
	}

	return nil
}

// periodicBackup performs regular backups of the knowledge base
func (kb *KnowledgeBase) periodicBackup() {
	// Backup every 6 hours
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-kb.ctx.Done():
			return
		case <-ticker.C:
			if err := kb.CreateBackup(); err != nil {
				kb.logger.WithError(err).Error("Failed to create periodic backup")
			}
		}
	}
}

// matchesQuery checks if a knowledge item matches a query
func matchesQuery(item KnowledgeItem, query KnowledgeQuery) bool {
	// Type filter
	if query.Type != "" && item.Type != query.Type {
		return false
	}

	// Tag filter
	if len(query.Tags) > 0 {
		found := false
		for _, queryTag := range query.Tags {
			for _, itemTag := range item.Tags {
				if itemTag == queryTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Confidence filter
	if query.MinConfidence > 0 && item.Confidence < query.MinConfidence {
		return false
	}

	// Source filter
	if query.Source != "" && !strings.Contains(item.Source, query.Source) {
		return false
	}

	// Date filter
	if !query.StartDate.IsZero() && item.CreatedAt.Before(query.StartDate) {
		return false
	}
	if !query.EndDate.IsZero() && item.CreatedAt.After(query.EndDate) {
		return false
	}

	return true
}
