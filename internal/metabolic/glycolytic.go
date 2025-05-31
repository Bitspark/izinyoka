package metabolic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/stream"
	"github.com/sirupsen/logrus"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskProcessing TaskStatus = "processing"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
	TaskCancelled  TaskStatus = "cancelled"
)

// Task represents a unit of work in the glycolytic cycle
type Task struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Priority     int                    `json:"priority"`
	Status       TaskStatus             `json:"status"`
	Data         map[string]interface{} `json:"data"`
	Result       map[string]interface{} `json:"result,omitempty"`
	Error        string                 `json:"error,omitempty"`
	CreatedAt    time.Time              `json:"createdAt"`
	StartedAt    time.Time              `json:"startedAt,omitempty"`
	CompletedAt  time.Time              `json:"completedAt,omitempty"`
	MaxDuration  time.Duration          `json:"maxDuration,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

// GlycolyticConfig contains configuration for the glycolytic cycle
type GlycolyticConfig struct {
	MaxConcurrentTasks  int           `json:"maxConcurrentTasks"`
	DefaultTaskPriority int           `json:"defaultTaskPriority"`
	DefaultMaxDuration  time.Duration `json:"defaultMaxDuration"`
	IdleCheckInterval   time.Duration `json:"idleCheckInterval"`
	EnableAutoScale     bool          `json:"enableAutoScale"`
	MinWorkers          int           `json:"minWorkers"`
	MaxWorkers          int           `json:"maxWorkers"`
}

// GlycolyticCycle implements task management inspired by cellular glycolysis
type GlycolyticCycle struct {
	config         GlycolyticConfig
	tasks          map[string]*Task
	priorityQueues map[int][]*Task
	inProgress     map[string]bool
	dependencies   map[string][]string
	dependents     map[string][]string
	workers        []*taskWorker
	ctx            context.Context
	cancelFunc     context.CancelFunc
	mutex          sync.RWMutex
	taskChan       chan *Task
	completionChan chan *Task
	logger         *logrus.Logger
}

// taskWorker represents a worker goroutine that processes tasks
type taskWorker struct {
	id        int
	cycle     *GlycolyticCycle
	ctx       context.Context
	cancel    context.CancelFunc
	taskChan  chan *Task
	idleSince time.Time
}

// NewGlycolyticCycle creates a new instance of the glycolytic cycle
func NewGlycolyticCycle(ctx context.Context, config GlycolyticConfig) *GlycolyticCycle {
	if config.MaxConcurrentTasks <= 0 {
		config.MaxConcurrentTasks = 10
	}
	if config.DefaultTaskPriority <= 0 {
		config.DefaultTaskPriority = 5
	}
	if config.DefaultMaxDuration <= 0 {
		config.DefaultMaxDuration = 5 * time.Minute
	}
	if config.IdleCheckInterval <= 0 {
		config.IdleCheckInterval = 30 * time.Second
	}
	if config.MinWorkers <= 0 {
		config.MinWorkers = 2
	}
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 20
	}

	cycleCtx, cancelFunc := context.WithCancel(ctx)

	cycle := &GlycolyticCycle{
		config:         config,
		tasks:          make(map[string]*Task),
		priorityQueues: make(map[int][]*Task),
		inProgress:     make(map[string]bool),
		dependencies:   make(map[string][]string),
		dependents:     make(map[string][]string),
		ctx:            cycleCtx,
		cancelFunc:     cancelFunc,
		taskChan:       make(chan *Task, config.MaxConcurrentTasks),
		completionChan: make(chan *Task, config.MaxConcurrentTasks),
		logger:         logrus.New(),
	}

	// Start the manager and initial workers
	go cycle.manager()
	cycle.scaleWorkers(config.MinWorkers)

	return cycle
}

// SubmitTask adds a new task to the glycolytic cycle
func (gc *GlycolyticCycle) SubmitTask(task *Task) (string, error) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// Generate task ID if not provided
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	// Check if task already exists
	if _, exists := gc.tasks[task.ID]; exists {
		return "", fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Set defaults
	if task.Priority <= 0 {
		task.Priority = gc.config.DefaultTaskPriority
	}
	if task.Status == "" {
		task.Status = TaskPending
	}
	if task.MaxDuration <= 0 {
		task.MaxDuration = gc.config.DefaultMaxDuration
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// Register the task
	gc.tasks[task.ID] = task

	// Add to priority queue
	priority := task.Priority
	if _, exists := gc.priorityQueues[priority]; !exists {
		gc.priorityQueues[priority] = make([]*Task, 0)
	}
	gc.priorityQueues[priority] = append(gc.priorityQueues[priority], task)

	// Record dependencies
	for _, depID := range task.Dependencies {
		gc.dependencies[task.ID] = append(gc.dependencies[task.ID], depID)
		gc.dependents[depID] = append(gc.dependents[depID], task.ID)
	}

	gc.logger.WithFields(logrus.Fields{
		"taskID":   task.ID,
		"taskType": task.Type,
		"priority": task.Priority,
	}).Info("Task submitted")

	return task.ID, nil
}

// GetTask retrieves information about a specific task
func (gc *GlycolyticCycle) GetTask(taskID string) (*Task, error) {
	gc.mutex.RLock()
	defer gc.mutex.RUnlock()

	task, exists := gc.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

// CancelTask cancels a pending or in-progress task
func (gc *GlycolyticCycle) CancelTask(taskID string) error {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	task, exists := gc.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if task.Status == TaskCompleted || task.Status == TaskFailed || task.Status == TaskCancelled {
		return fmt.Errorf("cannot cancel task %s with status %s", taskID, task.Status)
	}

	// Update the task status
	task.Status = TaskCancelled
	task.CompletedAt = time.Now()

	// Remove from priority queue if pending
	if task.Status == TaskPending {
		priority := task.Priority
		queue := gc.priorityQueues[priority]
		for i, t := range queue {
			if t.ID == taskID {
				gc.priorityQueues[priority] = append(queue[:i], queue[i+1:]...)
				break
			}
		}
	}

	// Remove from in-progress if necessary
	delete(gc.inProgress, taskID)

	gc.logger.WithFields(logrus.Fields{
		"taskID": taskID,
	}).Info("Task cancelled")

	return nil
}

// TaskFromJob converts a stream job to a glycolytic task
func TaskFromJob(job *stream.Job) *Task {
	return &Task{
		ID:          job.ID,
		Type:        job.ProcessType,
		Priority:    job.Priority,
		Status:      TaskPending,
		Data:        job.InputData,
		CreatedAt:   time.Now(),
		MaxDuration: time.Duration(job.MaxDuration) * time.Second,
	}
}

// JobFromTask converts a glycolytic task to a stream job
func JobFromTask(task *Task) *stream.Job {
	return &stream.Job{
		ID:          task.ID,
		InputData:   task.Data,
		ProcessType: task.Type,
		Priority:    task.Priority,
		MaxDuration: int(task.MaxDuration.Seconds()),
		Results:     task.Result,
	}
}

// manager processes task completion and schedules new tasks
func (gc *GlycolyticCycle) manager() {
	ticker := time.NewTicker(100 * time.Millisecond)
	idleCheckTicker := time.NewTicker(gc.config.IdleCheckInterval)
	defer ticker.Stop()
	defer idleCheckTicker.Stop()

	for {
		select {
		case <-gc.ctx.Done():
			return

		case completedTask := <-gc.completionChan:
			gc.handleCompletedTask(completedTask)

		case <-ticker.C:
			gc.scheduleNextBatch()

		case <-idleCheckTicker.C:
			if gc.config.EnableAutoScale {
				gc.adjustWorkerCount()
			}
		}
	}
}

// handleCompletedTask processes a task that has finished execution
func (gc *GlycolyticCycle) handleCompletedTask(task *Task) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// Mark as no longer in progress
	delete(gc.inProgress, task.ID)

	// Update task in the registry
	if existingTask, ok := gc.tasks[task.ID]; ok {
		existingTask.Status = task.Status
		existingTask.Result = task.Result
		existingTask.Error = task.Error
		existingTask.CompletedAt = task.CompletedAt
	}

	// Process dependent tasks
	dependents := gc.dependents[task.ID]
	for _, depID := range dependents {
		depTask, exists := gc.tasks[depID]
		if exists && depTask.Status == TaskPending {
			// Check if all dependencies are satisfied
			allDependenciesMet := true
			for _, reqID := range gc.dependencies[depID] {
				reqTask, reqExists := gc.tasks[reqID]
				if !reqExists || reqTask.Status != TaskCompleted {
					allDependenciesMet = false
					break
				}
			}

			// If dependencies are met, task can be scheduled
			if allDependenciesMet {
				priority := depTask.Priority
				if _, exists := gc.priorityQueues[priority]; !exists {
					gc.priorityQueues[priority] = []*Task{}
				}
				gc.priorityQueues[priority] = append(gc.priorityQueues[priority], depTask)
			}
		}
	}

	gc.logger.WithFields(logrus.Fields{
		"taskID": task.ID,
		"status": task.Status,
	}).Info("Task completed")
}

// scheduleNextBatch schedules available tasks for execution
func (gc *GlycolyticCycle) scheduleNextBatch() {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	// Find the highest priority with available tasks
	highestPriority := 0
	for priority := range gc.priorityQueues {
		if len(gc.priorityQueues[priority]) > 0 && priority > highestPriority {
			highestPriority = priority
		}
	}

	// If no tasks are available, return
	if highestPriority == 0 {
		return
	}

	// Get available worker capacity
	availableCapacity := cap(gc.taskChan) - len(gc.taskChan)
	if availableCapacity <= 0 {
		return
	}

	// Schedule tasks from highest priority queue
	queue := gc.priorityQueues[highestPriority]
	for i := 0; i < availableCapacity && i < len(queue); i++ {
		task := queue[i]

		// Skip tasks that are already in progress or have dependencies
		if gc.inProgress[task.ID] || !gc.areDependenciesMet(task.ID) {
			continue
		}

		// Mark as in progress and dispatch
		gc.inProgress[task.ID] = true
		task.Status = TaskProcessing
		task.StartedAt = time.Now()

		// Remove from queue
		if i == len(queue)-1 {
			gc.priorityQueues[highestPriority] = queue[:i]
		} else {
			gc.priorityQueues[highestPriority] = append(queue[:i], queue[i+1:]...)
		}

		// Send to worker
		select {
		case gc.taskChan <- task:
			gc.logger.WithFields(logrus.Fields{
				"taskID":   task.ID,
				"priority": task.Priority,
			}).Debug("Task dispatched to worker")
		default:
			// Revert if channel is full
			gc.inProgress[task.ID] = false
			task.Status = TaskPending
			task.StartedAt = time.Time{}
			gc.priorityQueues[highestPriority] = append(gc.priorityQueues[highestPriority], task)
		}
	}
}

// areDependenciesMet checks if all dependencies of a task are satisfied
func (gc *GlycolyticCycle) areDependenciesMet(taskID string) bool {
	deps, exists := gc.dependencies[taskID]
	if !exists || len(deps) == 0 {
		return true
	}

	for _, depID := range deps {
		depTask, exists := gc.tasks[depID]
		if !exists || depTask.Status != TaskCompleted {
			return false
		}
	}

	return true
}

// scaleWorkers adjusts the number of worker goroutines
func (gc *GlycolyticCycle) scaleWorkers(targetCount int) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	currentCount := len(gc.workers)

	// Scale down if needed
	if currentCount > targetCount {
		for i := targetCount; i < currentCount; i++ {
			gc.workers[i].cancel()
		}
		gc.workers = gc.workers[:targetCount]
		return
	}

	// Scale up if needed
	for i := currentCount; i < targetCount; i++ {
		workerCtx, cancel := context.WithCancel(gc.ctx)
		worker := &taskWorker{
			id:        i,
			cycle:     gc,
			ctx:       workerCtx,
			cancel:    cancel,
			taskChan:  gc.taskChan,
			idleSince: time.Now(),
		}
		gc.workers = append(gc.workers, worker)
		go worker.run()
	}

	gc.logger.WithFields(logrus.Fields{
		"oldCount": currentCount,
		"newCount": targetCount,
	}).Info("Worker count adjusted")
}

// adjustWorkerCount automatically scales workers based on workload
func (gc *GlycolyticCycle) adjustWorkerCount() {
	gc.mutex.Lock()
	totalPending := 0
	for _, queue := range gc.priorityQueues {
		totalPending += len(queue)
	}
	inProgressCount := len(gc.inProgress)
	currentWorkers := len(gc.workers)
	gc.mutex.Unlock()

	// Calculate target worker count based on workload
	targetCount := currentWorkers
	if totalPending > currentWorkers*2 && currentWorkers < gc.config.MaxWorkers {
		// Scale up if there are many pending tasks
		targetCount = min(currentWorkers+2, gc.config.MaxWorkers)
	} else if inProgressCount < currentWorkers/2 && currentWorkers > gc.config.MinWorkers {
		// Scale down if workers are underutilized
		targetCount = max(currentWorkers-1, gc.config.MinWorkers)
	}

	if targetCount != currentWorkers {
		gc.scaleWorkers(targetCount)
	}
}

// run is the main loop for a worker goroutine
func (tw *taskWorker) run() {
	for {
		select {
		case <-tw.ctx.Done():
			return

		case task, ok := <-tw.taskChan:
			if !ok {
				return
			}

			tw.idleSince = time.Time{}
			tw.processTask(task)
			tw.idleSince = time.Now()
		}
	}
}

// processTask handles the execution of a single task
func (tw *taskWorker) processTask(task *Task) {
	// Create a timeout context
	ctx, cancel := context.WithTimeout(tw.ctx, task.MaxDuration)
	defer cancel()

	// Execute the task
	logger := tw.cycle.logger.WithFields(logrus.Fields{
		"taskID":   task.ID,
		"worker":   tw.id,
		"taskType": task.Type,
	})

	logger.Info("Processing task")

	// In a real implementation, we would have a registry of task handlers
	// This is a simplified example where we simulate processing
	result := make(map[string]interface{})
	var taskErr error

	// Simulate task processing with timeout
	done := make(chan bool)
	go func() {
		// Simple simulation - in a real implementation this would call appropriate
		// handler based on task.Type
		time.Sleep(100 * time.Millisecond)

		// Simulate result based on task data
		result["processed"] = true
		result["timestamp"] = time.Now().Unix()
		if len(task.Data) > 0 {
			result["input_size"] = len(task.Data)
		}

		done <- true
	}()

	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			task.Status = TaskFailed
			task.Error = "task execution exceeded maximum duration"
			logger.Warn("Task execution timed out")
		} else {
			task.Status = TaskCancelled
			task.Error = "task was cancelled"
			logger.Info("Task was cancelled")
		}
	case <-done:
		if taskErr != nil {
			task.Status = TaskFailed
			task.Error = taskErr.Error()
			logger.WithError(taskErr).Error("Task execution failed")
		} else {
			task.Status = TaskCompleted
			task.Result = result
			logger.Info("Task executed successfully")
		}
	}

	// Record completion time
	task.CompletedAt = time.Now()

	// Send to completion channel
	select {
	case tw.cycle.completionChan <- task:
	default:
		logger.Error("Failed to send task to completion channel")
	}
}

// helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
