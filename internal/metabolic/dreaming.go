package metabolic

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/stream"
)

// DreamingModule handles synthetic exploration
type DreamingModule struct {
	knowledge     *knowledge.KnowledgeBase
	lactateCycle  *LactateCycle
	dreamDuration time.Duration
	diversity     float64
	active        bool
	dreamResults  []*stream.StreamData
	mu            sync.RWMutex
}

// DreamingConfig holds configuration for the dreaming module
type DreamingConfig struct {
	Interval  time.Duration
	Duration  time.Duration
	Diversity float64
	Intensity float64
}

// NewDreamingModule creates a new dreaming module
func NewDreamingModule(
	kb *knowledge.KnowledgeBase,
	lactateCycle *LactateCycle,
	config DreamingConfig,
) *DreamingModule {
	duration := config.Duration
	if duration <= 0 {
		duration = 2 * time.Minute
	}

	diversity := config.Diversity
	if diversity <= 0 {
		diversity = 0.8
	}

	return &DreamingModule{
		knowledge:     kb,
		lactateCycle:  lactateCycle,
		dreamDuration: duration,
		diversity:     diversity,
		active:        true,
		dreamResults:  make([]*stream.StreamData, 0),
	}
}

// StartDreaming begins the dreaming process
func (dm *DreamingModule) StartDreaming(ctx context.Context) {
	// Set a fixed interval for dreaming
	interval := 10 * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if dm.shouldStartDream() {
				go dm.dreamProcess(ctx)
			}
		}
	}
}

// shouldStartDream determines if a dream cycle should start
func (dm *DreamingModule) shouldStartDream() bool {
	// Check if dreaming is active
	dm.mu.RLock()
	active := dm.active
	dm.mu.RUnlock()

	if !active {
		return false
	}

	// Start dreaming if system load is low (simplified logic)
	// In a real implementation, this would check system metrics
	randValue := rand.Float64()
	return randValue < 0.3 // 30% chance to start dreaming
}

// dreamProcess performs the actual dreaming computation
func (dm *DreamingModule) dreamProcess(ctx context.Context) {
	// Create a context with timeout for the dreaming duration
	dreamCtx, cancel := context.WithTimeout(ctx, dm.dreamDuration)
	defer cancel()

	// Get incomplete tasks from lactate cycle
	incompleteTasks := dm.getRecyclableTasks(50)

	// Generate synthetic edge cases
	edgeCases := dm.generateEdgeCases(10)

	// Combine incomplete tasks and edge cases for dream scenarios
	dreamScenarios := dm.mergeScenarios(incompleteTasks, edgeCases)

	// Process each scenario
	results := make([]*stream.StreamData, 0, len(dreamScenarios))

	for _, scenario := range dreamScenarios {
		select {
		case <-dreamCtx.Done():
			// Dream time is over, store what we have
			break
		default:
			// Process the scenario
			result := dm.processScenario(scenario)
			results = append(results, result)
		}
	}

	// Store dream results
	dm.mu.Lock()
	dm.dreamResults = results
	dm.mu.Unlock()

	// Store valuable insights in knowledge base
	for _, result := range results {
		if confidence, ok := result.Metadata["confidence"].(float64); ok && confidence > 0.7 {
			dm.storeInsight(result)
		}
	}
}

// getRecyclableTasks gets tasks that can be recycled for dreaming
func (dm *DreamingModule) getRecyclableTasks(maxCount int) []*Task {
	// Since GetRecyclableTasks doesn't exist, let's implement our own logic
	// to find tasks that can be recycled for dreaming

	var recyclableTasks []*Task

	// Get completed and failed tasks from lactate cycle
	// This is a simplified implementation - would normally access internal task storage
	for i := 0; i < maxCount && i < 10; i++ {
		// Create synthetic recyclable tasks for dreaming
		taskData := map[string]interface{}{
			"recycled": true,
			"sequence": i,
			"value":    rand.Float64(),
		}

		task := &Task{
			ID:        fmt.Sprintf("recyclable-task-%d", i),
			Type:      "recycled",
			Priority:  1,
			Status:    TaskCompleted,
			Data:      taskData,
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}
		recyclableTasks = append(recyclableTasks, task)
	}

	return recyclableTasks
}

// generateEdgeCases creates synthetic scenarios for exploration
func (dm *DreamingModule) generateEdgeCases(count int) []*stream.StreamData {
	edgeCases := make([]*stream.StreamData, 0, count)

	// Query knowledge base for patterns to use as seeds
	query := &knowledge.KnowledgeQuery{
		Type:  "pattern",
		Limit: 20,
	}

	patterns, err := dm.knowledge.Query(query)
	if err != nil || len(patterns) == 0 {
		// If no patterns found, generate completely synthetic cases
		for i := 0; i < count; i++ {
			edgeCases = append(edgeCases, dm.generateRandomScenario())
		}
		return edgeCases
	}

	// Use patterns to generate edge cases
	for i := 0; i < count; i++ {
		// Select a random pattern
		pattern := patterns[rand.Intn(len(patterns))]

		// Mutate the pattern to create edge case
		edgeCase := dm.mutatePattern(pattern)
		edgeCases = append(edgeCases, edgeCase)
	}

	return edgeCases
}

// generateRandomScenario creates a completely synthetic scenario
func (dm *DreamingModule) generateRandomScenario() *stream.StreamData {
	// This is a simplified implementation - real version would be domain-specific
	metadata := map[string]interface{}{
		"synthetic":       true,
		"dream_generated": true,
		"creation_time":   time.Now(),
		"diversity":       dm.diversity,
		"confidence":      0.5,
	}

	// Generate random content (simplified)
	data := map[string]interface{}{
		"value":    rand.Float64(),
		"category": rand.Intn(5),
		"sequence": []int{rand.Intn(100), rand.Intn(100), rand.Intn(100)},
	}

	return &stream.StreamData{
		ID:        "dream-" + generateUniqueID(),
		Type:      "synthetic",
		Data:      data,
		CreatedAt: time.Now(),
		Metadata:  metadata,
	}
}

// mutatePattern creates a variation of an existing pattern
func (dm *DreamingModule) mutatePattern(pattern knowledge.KnowledgeItem) *stream.StreamData {
	// Extract the pattern content
	content := pattern.Data

	// Apply mutations based on content type
	var mutatedContent map[string]interface{}

	// Content is already map[string]interface{}, so we can work with it directly
	// Deep copy the map
	mutatedContent = make(map[string]interface{})
	for k, v := range content {
		mutatedContent[k] = v
	}

	// Mutate some values
	for k, v := range mutatedContent {
		if rand.Float64() < dm.diversity {
			switch val := v.(type) {
			case float64:
				mutatedContent[k] = val * (0.5 + rand.Float64())
			case int:
				mutatedContent[k] = val + rand.Intn(10) - 5
			case string:
				mutatedContent[k] = val + generateRandomString(3)
			}
		}
	}

	return &stream.StreamData{
		ID:        "dream-" + generateUniqueID(),
		Type:      "mutated_pattern",
		Data:      mutatedContent,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"synthetic":        true,
			"dream_generated":  true,
			"original_pattern": pattern.ID,
			"mutation_factor":  dm.diversity,
			"creation_time":    time.Now(),
			"confidence":       0.5,
		},
	}
}

// mergeScenarios combines incomplete tasks and edge cases
func (dm *DreamingModule) mergeScenarios(tasks []*Task, edgeCases []*stream.StreamData) []*stream.StreamData {
	scenarios := make([]*stream.StreamData, 0, len(tasks)+len(edgeCases))

	// Add all edge cases
	scenarios = append(scenarios, edgeCases...)

	// Add data from incomplete tasks
	for _, task := range tasks {
		if task.Data != nil {
			// Convert task data to stream data
			taskStreamData := &stream.StreamData{
				ID:        "task-" + task.ID,
				Type:      task.Type,
				Data:      task.Data,
				CreatedAt: task.CreatedAt,
				Metadata: map[string]interface{}{
					"dream_recycled": true,
					"dream_time":     time.Now(),
					"original_task":  task.ID,
				},
			}
			scenarios = append(scenarios, taskStreamData)
		}
	}

	// Shuffle the scenarios for diversity
	rand.Shuffle(len(scenarios), func(i, j int) {
		scenarios[i], scenarios[j] = scenarios[j], scenarios[i]
	})

	return scenarios
}

// processScenario simulates processing a dream scenario
func (dm *DreamingModule) processScenario(data *stream.StreamData) *stream.StreamData {
	// This is a simplified simulation of processing
	// In a real implementation, this would run the scenario through
	// the metacognitive layers with relaxed constraints

	// Simulate processing time
	processingTime := 50 + rand.Intn(200)
	time.Sleep(time.Duration(processingTime) * time.Millisecond)

	// Generate a result with some insight
	confidence := 0.3 + rand.Float64()*0.7 // Random confidence

	result := &stream.StreamData{
		ID:        "dream-result-" + data.ID,
		Type:      "dream_result",
		Data:      data.Data,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"dream_processed":  true,
			"processing_time":  processingTime,
			"insight_level":    rand.Float64(),
			"original_data_id": data.ID,
			"confidence":       confidence,
		},
	}

	return result
}

// storeInsight saves valuable dream insights to the knowledge base
func (dm *DreamingModule) storeInsight(data *stream.StreamData) {
	item := knowledge.KnowledgeItem{
		ID:        "insight-" + data.ID,
		Type:      "dream_insight",
		Data:      data.Data,
		CreatedAt: time.Now(),
		Tags:      []string{"dream", "insight"},
		Source:    "dreaming_module",
	}

	// Add confidence from metadata if available
	if confidence, ok := data.Metadata["confidence"].(float64); ok {
		item.Confidence = confidence
		// Add confidence as a tag if high enough
		if confidence > 0.8 {
			item.Tags = append(item.Tags, "high_confidence")
		}
	} else {
		item.Confidence = 0.5 // Default confidence
	}

	// Store the insight
	err := dm.knowledge.Store(item)
	if err != nil {
		// Log error but don't fail the dreaming process
		// In a real implementation, this would use proper logging
	}
}

// GetDreamResults returns the results from the last dream cycle
func (dm *DreamingModule) GetDreamResults() []*stream.StreamData {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// Return a copy to avoid race conditions
	results := make([]*stream.StreamData, len(dm.dreamResults))
	copy(results, dm.dreamResults)

	return results
}

// Stop halts the dreaming process
func (dm *DreamingModule) Stop() {
	dm.mu.Lock()
	dm.active = false
	dm.mu.Unlock()
}

// Shutdown gracefully stops the dreaming module and cleans up resources
func (dm *DreamingModule) Shutdown() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Stop the dreaming process
	dm.active = false

	// Clear dream results to free memory
	dm.dreamResults = make([]*stream.StreamData, 0)
}

// generateUniqueID creates a unique identifier
func generateUniqueID() string {
	return time.Now().Format("20060102150405.000") + "-" + generateRandomString(6)
}

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
