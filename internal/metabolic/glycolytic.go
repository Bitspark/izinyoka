package metabolic

import (
	"context"
	"sync"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/knowledge"
	"github.com/fullscreen-triangle/izinyoka/internal/stream"
)

// GlycolicCycle manages task partitioning and resource allocation
type GlycolicCycle struct {
	knowledge     *knowledge.KnowledgeBase
	taskBatchSize int
	resourceLimit int
	resourcePool  *ResourcePool
	active        bool
	mu            sync.Mutex
}

// ResourcePool manages computational resources
type ResourcePool struct {
	total     int
	available int
	mu        sync.Mutex
}

// Task represents a computational unit
type Task struct {
	ID          string
	Priority    float64
	Complexity  float64
	Deadline    time.Time
	Data        stream.StreamData
	Status      TaskStatus
	Allocations map[string]float64
}

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskProcessing TaskStatus = "processing"
	TaskComplete   TaskStatus = "complete"
	TaskIncomplete TaskStatus = "incomplete"
)

// GlycolicConfig holds configuration for the glycolytic cycle
type GlycolicConfig struct {
	TaskBatchSize  int
	ResourceLimit  int
	TimeoutMs      int
	PriorityLevels int
}

// NewGlycolicCycle creates a new glycolytic cycle
func NewGlycolicCycle(kb *knowledge.KnowledgeBase, config GlycolicConfig) *GlycolicCycle {
	resourceLimit := config.ResourceLimit
	if resourceLimit <= 0 {
		resourceLimit = 100 // Default value
	}

	return &GlycolicCycle{
		knowledge:     kb,
		taskBatchSize: config.TaskBatchSize,
		resourceLimit: resourceLimit,
		resourcePool: &ResourcePool{
			total:     resourceLimit,
			available: resourceLimit,
		},
		active: true,
	}
}

// AllocateTasks breaks down input data into manageable tasks and allocates resources
func (gc *GlycolicCycle) AllocateTasks(ctx context.Context, input <-chan stream.StreamData) <-chan stream.StreamData {
	output := make(chan stream.StreamData)

	go func() {
		defer close(output)

		taskBatch := make([]stream.StreamData, 0, gc.taskBatchSize)
		timer := time.NewTimer(100 * time.Millisecond)

		for {
			select {
			case <-ctx.Done():
				return

			case data, ok := <-input:
				if !ok {
					// Process any remaining tasks
					if len(taskBatch) > 0 {
						gc.processBatch(ctx, taskBatch, output)
					}
					return
				}

				taskBatch = append(taskBatch, data)

				// Process batch when it reaches the desired size
				if len(taskBatch) >= gc.taskBatchSize {
					gc.processBatch(ctx, taskBatch, output)
					taskBatch = make([]stream.StreamData, 0, gc.taskBatchSize)

					// Reset timer
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(100 * time.Millisecond)
				}

			case <-timer.C:
				// Process batch on timeout if not empty
				if len(taskBatch) > 0 {
					gc.processBatch(ctx, taskBatch, output)
					taskBatch = make([]stream.StreamData, 0, gc.taskBatchSize)
				}
				timer.Reset(100 * time.Millisecond)
			}
		}
	}()

	return output
}

// processBatch handles a batch of tasks
func (gc *GlycolicCycle) processBatch(
	ctx context.Context,
	batch []stream.StreamData,
	output chan<- stream.StreamData,
) {
	// Calculate complexity of each task
	tasks := make([]Task, len(batch))
	for i, data := range batch {
		complexity := calculateComplexity(data)
		priority := calculatePriority(data)

		tasks[i] = Task{
			ID:         "task-" + data.ID,
			Data:       data,
			Complexity: complexity,
			Priority:   priority,
			Deadline:   time.Now().Add(5 * time.Second),
			Status:     TaskPending,
		}
	}

	// Sort tasks by priority (higher priority first)
	sortTasksByPriority(tasks)

	// Allocate resources to tasks
	allocatedTasks := gc.allocateResources(tasks)

	// Process each allocated task
	for _, task := range allocatedTasks {
		select {
		case <-ctx.Done():
			return
		case output <- task.Data:
			// Task dispatched for processing
		}
	}
}

// allocateResources distributes available resources to tasks based on priority
func (gc *GlycolicCycle) allocateResources(tasks []Task) []Task {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if !gc.active {
		return nil
	}

	allocatedTasks := make([]Task, 0, len(tasks))

	gc.resourcePool.mu.Lock()
	availableResources := gc.resourcePool.available
	gc.resourcePool.mu.Unlock()

	// Calculate total complexity-adjusted priority
	var totalPriority float64
	for _, task := range tasks {
		totalPriority += task.Priority * task.Complexity
	}

	// Allocate resources based on priority and complexity
	remainingResources := availableResources
	for i := range tasks {
		if remainingResources <= 0 {
			break
		}

		// Calculate fair share of resources
		share := int((tasks[i].Priority * tasks[i].Complexity / totalPriority) * float64(availableResources))
		if share > remainingResources {
			share = remainingResources
		}

		// Don't allocate if resources are below minimum threshold
		minResources := int(tasks[i].Complexity * 10)
		if share < minResources {
			continue
		}

		// Update task with resource allocation
		tasks[i].Status = TaskProcessing
		tasks[i].Allocations = map[string]float64{
			"cpu":    float64(share) * 0.8,
			"memory": float64(share) * 0.2,
		}

		allocatedTasks = append(allocatedTasks, tasks[i])
		remainingResources -= share
	}

	// Update resource pool
	gc.resourcePool.mu.Lock()
	gc.resourcePool.available = remainingResources
	gc.resourcePool.mu.Unlock()

	return allocatedTasks
}

// Shutdown stops the glycolytic cycle and releases resources
func (gc *GlycolicCycle) Shutdown(ctx context.Context) {
	gc.mu.Lock()
	gc.active = false
	gc.mu.Unlock()

	// Wait for resources to be released
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			gc.resourcePool.mu.Lock()
			available := gc.resourcePool.available
			total := gc.resourcePool.total
			gc.resourcePool.mu.Unlock()

			if available == total {
				return
			}
			timer.Reset(100 * time.Millisecond)
		}
	}
}

// calculateComplexity estimates the computational complexity of a task
func calculateComplexity(data stream.StreamData) float64 {
	// This is a simplistic model - real implementation would be more sophisticated
	complexity := 1.0

	// If content is large, increase complexity
	switch v := data.Content.(type) {
	case string:
		complexity += float64(len(v)) / 1000.0
	case []byte:
		complexity += float64(len(v)) / 1000.0
	case map[string]interface{}:
		complexity += float64(len(v)) / 10.0
	}

	// Check metadata for complexity hints
	if c, ok := data.Metadata["complexity"].(float64); ok {
		complexity = c
	}

	return complexity
}

// calculatePriority determines the processing priority of a task
func calculatePriority(data stream.StreamData) float64 {
	// Default priority
	priority := 1.0

	// Check if priority is specified in metadata
	if p, ok := data.Metadata["priority"].(float64); ok {
		priority = p
	}

	// Boost priority for time-sensitive data
	if deadline, ok := data.Metadata["deadline"].(time.Time); ok {
		timeRemaining := time.Until(deadline).Seconds()
		if timeRemaining < 10 {
			// Urgency boost for imminent deadlines
			priority += (10 - timeRemaining) / 2
		}
	}

	return priority
}

// sortTasksByPriority sorts tasks by priority (higher first)
func sortTasksByPriority(tasks []Task) {
	// Simple insertion sort for small batches
	for i := 1; i < len(tasks); i++ {
		key := tasks[i]
		j := i - 1

		for j >= 0 && tasks[j].Priority < key.Priority {
			tasks[j+1] = tasks[j]
			j--
		}

		tasks[j+1] = key
	}
}
