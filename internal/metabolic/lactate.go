package metabolic

import (
	"context"
	"sync"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/stream"
)

// LactateCycle manages incomplete computations
type LactateCycle struct {
	knowledge          *knowledge.KnowledgeBase
	incompleteTasks    map[string]Task
	maxIncompleteTasks int
	recycleThreshold   float64
	mu                 sync.RWMutex
}

// LactateConfig holds configuration for the lactate cycle
type LactateConfig struct {
	MaxIncompleteTasks int
	RecycleThreshold   float64
	PersistPath        string
}

// NewLactateCycle creates a new lactate cycle
func NewLactateCycle(kb *knowledge.KnowledgeBase, config LactateConfig) *LactateCycle {
	maxIncompleteTasks := config.MaxIncompleteTasks
	if maxIncompleteTasks <= 0 {
		maxIncompleteTasks = 100 // Default value
	}

	recycleThreshold := config.RecycleThreshold
	if recycleThreshold <= 0 {
		recycleThreshold = 0.4 // Default value
	}

	return &LactateCycle{
		knowledge:          kb,
		incompleteTasks:    make(map[string]Task),
		maxIncompleteTasks: maxIncompleteTasks,
		recycleThreshold:   recycleThreshold,
	}
}

// StoreIncompleteTask stores a task that couldn't complete
func (lc *LactateCycle) StoreIncompleteTask(data stream.StreamData) {
	// Extract completion percentage from metadata or estimate it
	completionPercentage := 0.0
	if cp, ok := data.Metadata["completion_percentage"].(float64); ok {
		completionPercentage = cp
	} else {
		completionPercentage = data.Confidence
	}

	// Only store tasks that have made substantial progress
	if completionPercentage < 0.1 {
		return
	}

	task := Task{
		ID:         "incomplete-" + data.ID,
		Data:       data,
		Complexity: calculateComplexity(data),
		Priority:   calculatePriority(data),
		Deadline:   time.Now().Add(time.Hour), // Long deadline for recycling
		Status:     TaskIncomplete,
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	// If we've reached capacity, remove lowest priority task
	if len(lc.incompleteTasks) >= lc.maxIncompleteTasks {
		lowestPriority := task.Priority
		lowestID := task.ID

		for id, t := range lc.incompleteTasks {
			if t.Priority < lowestPriority {
				lowestPriority = t.Priority
				lowestID = id
			}
		}

		// Only store if this task has higher priority than the lowest
		if lowestPriority < task.Priority {
			delete(lc.incompleteTasks, lowestID)
			lc.incompleteTasks[task.ID] = task
		}
	} else {
		lc.incompleteTasks[task.ID] = task
	}
}

// GetRecyclableTasks returns tasks that can be resumed
func (lc *LactateCycle) GetRecyclableTasks(availableResources int) []Task {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var recyclable []Task
	remainingResources := availableResources

	// Convert map to slice for sorting
	tasks := make([]Task, 0, len(lc.incompleteTasks))
	for _, task := range lc.incompleteTasks {
		tasks = append(tasks, task)
	}

	// Sort by priority
	sortTasksByPriority(tasks)

	// Select tasks until resources are exhausted
	for _, task := range tasks {
		resourceNeeded := int(task.Complexity * 10)
		if resourceNeeded > remainingResources {
			continue
		}

		recyclable = append(recyclable, task)
		remainingResources -= resourceNeeded

		if remainingResources <= 0 {
			break
		}
	}

	return recyclable
}

// RecycleTasks returns a stream of recyclable tasks
func (lc *LactateCycle) RecycleTasks(ctx context.Context, resourceAvailability <-chan int) <-chan stream.StreamData {
	out := make(chan stream.StreamData)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return

			case resources, ok := <-resourceAvailability:
				if !ok {
					return
				}

				if resources <= 0 {
					continue
				}

				// Get recyclable tasks
				tasks := lc.GetRecyclableTasks(resources)
				if len(tasks) == 0 {
					continue
				}

				// Remove these tasks from incomplete list
				lc.mu.Lock()
				for _, task := range tasks {
					delete(lc.incompleteTasks, task.ID)
				}
				lc.mu.Unlock()

				// Send tasks for processing
				for _, task := range tasks {
					select {
					case <-ctx.Done():
						return
					case out <- task.Data:
						// Mark as recycled in metadata
						task.Data.Metadata["recycled"] = true
						task.Data.Metadata["recycled_time"] = time.Now()
					}
				}
			}
		}
	}()

	return out
}

// PersistIncompleteTasks saves incomplete tasks to the knowledge base
func (lc *LactateCycle) PersistIncompleteTasks() error {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	// Store each incomplete task in the knowledge base for long-term storage
	for _, task := range lc.incompleteTasks {
		item := knowledge.KnowledgeItem{
			ID:        task.ID,
			Type:      "incomplete_task",
			Content:   task.Data.Content,
			Metadata:  task.Data.Metadata,
			Timestamp: time.Now(),
			Tags:      []string{"incomplete", "lactate_cycle"},
		}

		if err := lc.knowledge.Store(item); err != nil {
			return err
		}
	}

	return nil
}

// LoadIncompleteTasks loads persisted incomplete tasks from the knowledge base
func (lc *LactateCycle) LoadIncompleteTasks() error {
	query := &knowledge.KnowledgeQuery{
		Type: "incomplete_task",
		Tags: []string{"incomplete", "lactate_cycle"},
	}

	items, err := lc.knowledge.Query(query)
	if err != nil {
		return err
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()

	for _, item := range items {
		// Convert knowledge item back to task
		task := Task{
			ID:     item.ID,
			Status: TaskIncomplete,
			Data: stream.StreamData{
				Content:    item.Content,
				Metadata:   item.Metadata,
				Timestamp:  item.Timestamp,
				ID:         item.ID,
				Confidence: 0.0, // Will be recalculated
			},
		}

		// Calculate complexity and priority
		task.Complexity = calculateComplexity(task.Data)
		task.Priority = calculatePriority(task.Data)

		// Only load if we have space
		if len(lc.incompleteTasks) < lc.maxIncompleteTasks {
			lc.incompleteTasks[task.ID] = task
		}
	}

	return nil
}
