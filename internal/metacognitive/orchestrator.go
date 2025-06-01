package metacognitive

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fullscreen-triangle/izinyoka/internal/metabolic"
	"github.com/fullscreen-triangle/izinyoka/internal/stream"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// OrchestratorConfig contains orchestrator configuration
type OrchestratorConfig struct {
	BalanceThreshold      float64 `yaml:"balanceThreshold" json:"balanceThreshold"`
	AdaptiveWeights       bool    `yaml:"adaptiveWeights" json:"adaptiveWeights"`
	PrioritizeConsistency bool    `yaml:"prioritizeConsistency" json:"prioritizeConsistency"`
}

// Orchestrator coordinates the three metacognitive layers and metabolic components
type Orchestrator struct {
	contextLayer     *ContextLayer
	reasoningLayer   *ReasoningLayer
	intuitionLayer   *IntuitionLayer
	metabolicManager *metabolic.Manager
	config           OrchestratorConfig
	streamRouter     *streamRouter
	jobRegistry      map[string]*jobInfo
	jobStatusInfo    map[string]*stream.JobStatusInfo
	domain           map[string]DomainAdapter
	mutex            sync.RWMutex
	ctx              context.Context
	cancelFunc       context.CancelFunc
	logger           *logrus.Logger
}

// jobInfo contains internal job processing information
type jobInfo struct {
	job           *stream.Job
	streamData    *stream.StreamData
	routingConfig *routingConfig
	errors        []string
	startTime     time.Time
}

// routingConfig defines how data flows through the metacognitive layers
type routingConfig struct {
	startLayer        string   // "context", "reasoning", or "intuition"
	processingFlow    []string // Ordered processing layers, e.g. ["context", "reasoning", "intuition"]
	skipLayers        []string // Layers to skip, if any
	contextWeight     float64  // How much influence the context layer has (0-1)
	reasoningWeight   float64  // How much influence the reasoning layer has (0-1)
	intuitionWeight   float64  // How much influence the intuition layer has (0-1)
	feedbackEnabled   bool     // Whether to provide feedback to knowledge base
	adaptiveRouting   bool     // Whether to adapt routing based on confidence
	confidenceBooster float64  // Additional confidence factor for results (0-1)
}

// streamRouter handles the flow of data between layers
type streamRouter struct {
	orchestrator        *Orchestrator
	inputChan           chan *stream.StreamData
	outputChan          chan *stream.StreamData
	contextInputChan    chan *stream.StreamData
	contextOutputChan   chan *stream.StreamData
	reasoningInputChan  chan *stream.StreamData
	reasoningOutputChan chan *stream.StreamData
	intuitionInputChan  chan *stream.StreamData
	intuitionOutputChan chan *stream.StreamData
}

// MetacognitiveOrchestrator coordinates the three-layer metacognitive architecture
type MetacognitiveOrchestrator struct {
	contextLayer   *ContextLayer
	reasoningLayer *ReasoningLayer
	intuitionLayer *IntuitionLayer
	
	// Configuration
	config OrchestratorConfig
	
	// State management
	state         *OrchestratorState
	metrics       *OrchestratorMetrics
	mu            sync.RWMutex
	
	// Communication channels
	contextChan   chan *ContextResult
	reasoningChan chan *ReasoningResult
	intuitionChan chan *IntuitionResult
	errorChan     chan error
	
	// Lifecycle
	running bool
	done    chan struct{}
}

// OrchestratorState tracks the current state of the orchestrator
type OrchestratorState struct {
	CurrentData     *stream.StreamData
	LastProcessed   time.Time
	ProcessingCount int64
	ErrorCount      int64
	
	// Layer states
	ContextActive   bool
	ReasoningActive bool
	IntuitionActive bool
	
	// Performance tracking
	AverageProcessingTime time.Duration
	TotalProcessingTime   time.Duration
}

// OrchestratorMetrics tracks performance metrics
type OrchestratorMetrics struct {
	ProcessingTimes    []time.Duration
	ConfidenceScores   []float64
	ThroughputMetrics  map[string]float64
	ErrorRates         map[string]float64
	LayerPerformance   map[string]LayerMetrics
	mu                 sync.RWMutex
}

// LayerMetrics tracks metrics for individual layers
type LayerMetrics struct {
	ProcessingTime time.Duration
	Confidence     float64
	ErrorRate      float64
	Throughput     float64
}

// ProcessingResult represents the final result from the orchestrator
type ProcessingResult struct {
	// Individual layer results
	ContextResult   *ContextResult
	ReasoningResult *ReasoningResult
	IntuitionResult *IntuitionResult
	
	// Aggregated results
	FinalDecision    interface{}
	Confidence       float64
	Reasoning        string
	Patterns         []Pattern
	Predictions      []Prediction
	
	// Metadata
	ProcessingTime   time.Duration
	LayerTiming      map[string]time.Duration
	Timestamp        time.Time
	InputHash        string
}

// NewOrchestrator creates a new metacognitive orchestrator
func NewOrchestrator(
	contextLayer *ContextLayer,
	reasoningLayer *ReasoningLayer,
	intuitionLayer *IntuitionLayer,
	metabolicManager *metabolic.Manager,
	config OrchestratorConfig,
) *Orchestrator {
	// Create cancellable context
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Create orchestrator instance
	orch := &Orchestrator{
		contextLayer:     contextLayer,
		reasoningLayer:   reasoningLayer,
		intuitionLayer:   intuitionLayer,
		metabolicManager: metabolicManager,
		config:           config,
		jobRegistry:      make(map[string]*jobInfo),
		jobStatusInfo:    make(map[string]*stream.JobStatusInfo),
		domain:           make(map[string]DomainAdapter),
		ctx:              ctx,
		cancelFunc:       cancelFunc,
		logger:           logrus.New(),
	}

	// Create and configure the stream router
	orch.streamRouter = newStreamRouter(orch)

	return orch
}

// Start initializes and starts the orchestrator
func (o *Orchestrator) Start(ctx context.Context) error {
	o.logger.Info("Starting metacognitive orchestrator")

	// Start the stream router
	go o.streamRouter.start()

	// Start progress monitoring
	go o.monitorJobProgress()

	return nil
}

// Stop gracefully shuts down the orchestrator
func (o *Orchestrator) Stop(ctx context.Context) error {
	o.logger.Info("Stopping metacognitive orchestrator")

	// Signal cancellation to all components
	o.cancelFunc()

	return nil
}

// RegisterDomain registers a domain-specific adapter
func (o *Orchestrator) RegisterDomain(name string, adapter DomainAdapter) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if _, exists := o.domain[name]; exists {
		return fmt.Errorf("domain %s is already registered", name)
	}

	o.domain[name] = adapter
	o.logger.WithField("domain", name).Info("Domain registered with orchestrator")

	return nil
}

// SubmitJob submits a new job for processing
func (o *Orchestrator) SubmitJob(ctx context.Context, job *stream.Job) (string, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	// Generate job ID if not provided
	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%s", uuid.New().String())
	}

	// Check if job already exists
	if _, exists := o.jobRegistry[job.ID]; exists {
		return "", fmt.Errorf("job with ID %s already exists", job.ID)
	}

	// Set up job timestamps
	now := time.Now()
	job.CreatedAt = now
	job.Status = stream.JobPending

	// Create job info
	jobInfo := &jobInfo{
		job:       job,
		startTime: now,
		errors:    make([]string, 0),
		routingConfig: &routingConfig{
			startLayer:      "context",
			processingFlow:  []string{"context", "reasoning", "intuition"},
			contextWeight:   0.4,
			reasoningWeight: 0.4,
			intuitionWeight: 0.2,
			feedbackEnabled: true,
			adaptiveRouting: o.config.AdaptiveWeights,
		},
	}

	// Create job status info
	statusInfo := stream.NewJobStatusInfo(job.ID, stream.JobPending)
	o.jobStatusInfo[job.ID] = statusInfo

	// Register the job
	o.jobRegistry[job.ID] = jobInfo

	// Create a stream data object from the job
	streamData := &stream.StreamData{
		ID:        fmt.Sprintf("data-%s", uuid.New().String()),
		Type:      job.ProcessType,
		Data:      job.InputData,
		CreatedAt: now,
		Metadata: map[string]interface{}{
			"jobId":       job.ID,
			"processType": job.ProcessType,
			"priority":    job.Priority,
		},
	}

	// Store the stream data in the job info
	jobInfo.streamData = streamData

	// Submit to the metabolic glycolytic cycle as a task
	task := &metabolic.Task{
		ID:          job.ID,
		Type:        job.ProcessType,
		Priority:   job.Priority,
		Status:      metabolic.TaskPending,
		Data:        job.InputData,
		CreatedAt:   now,
		MaxDuration: time.Duration(job.MaxDuration) * time.Second,
	}

	// Use the glycolytic cycle to prioritize and manage the task
	glycolytic := o.metabolicManager.GetGlycolyticCycle()
	_, err := glycolytic.SubmitTask(task)
	if err != nil {
		return "", fmt.Errorf("failed to submit task to glycolytic cycle: %w", err)
	}

	// Update job status
	job.Status = stream.JobProcessing
	job.StartedAt = time.Now()
	statusInfo.Status = stream.JobProcessing
	statusInfo.StartedAt = job.StartedAt
	statusInfo.CurrentStep = "preparing"

	// Submit the stream data to the router
	go func() {
		// Send the data to the router
		o.streamRouter.inputChan <- streamData
	}()

	o.logger.WithFields(logrus.Fields{
		"jobId":       job.ID,
		"processType": job.ProcessType,
	}).Info("Job submitted")

	return job.ID, nil
}

// SubmitDomainJob submits a job to a specific domain for processing
func (o *Orchestrator) SubmitDomainJob(ctx context.Context, domainName, processType string, inputData interface{}) (string, error) {
	o.mutex.RLock()
	domainAdapter, exists := o.domain[domainName]
	o.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("domain %s not registered", domainName)
	}

	// Convert domain-specific data to generic map
	dataMap, err := domainAdapter.ConvertToGeneric(inputData)
	if err != nil {
		return "", fmt.Errorf("failed to convert domain data: %w", err)
	}

	// Create a job with the domain and process type
	job := &stream.Job{
		ProcessType: fmt.Sprintf("%s.%s", domainName, processType),
		InputData:   dataMap,
		Priority:    5,   // Medium priority by default
		MaxDuration: 300, // 5 minutes default
	}

	// Submit the generic job
	return o.SubmitJob(ctx, job)
}

// GetJobStatus retrieves the status of a job
func (o *Orchestrator) GetJobStatus(ctx context.Context, jobID string) (*stream.JobStatusInfo, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	statusInfo, exists := o.jobStatusInfo[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	return statusInfo, nil
}

// GetJobResults retrieves the results of a completed job
func (o *Orchestrator) GetJobResults(ctx context.Context, jobID string) (map[string]interface{}, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	jobInfo, exists := o.jobRegistry[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	if jobInfo.job.Status != stream.JobCompleted {
		return nil, fmt.Errorf("job %s is not completed (status: %s)", jobID, jobInfo.job.Status)
	}

	return jobInfo.job.Results, nil
}

// CancelJob cancels a running job
func (o *Orchestrator) CancelJob(ctx context.Context, jobID string) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	jobInfo, exists := o.jobRegistry[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	if jobInfo.job.Status == stream.JobCompleted ||
		jobInfo.job.Status == stream.JobFailed ||
		jobInfo.job.Status == stream.JobCancelled {
		return fmt.Errorf("job %s is already in final state: %s", jobID, jobInfo.job.Status)
	}

	// Update job status
	jobInfo.job.Status = stream.JobCancelled
	jobInfo.job.CompletedAt = time.Now()

	// Update status info
	if statusInfo, exists := o.jobStatusInfo[jobID]; exists {
		statusInfo.Status = stream.JobCancelled
		statusInfo.UpdatedAt = time.Now()
		statusInfo.CurrentStep = "cancelled"
	}

	// Try to cancel the task in the glycolytic cycle
	glycolytic := o.metabolicManager.GetGlycolyticCycle()
	_ = glycolytic.CancelTask(jobID) // Ignore errors, as task might have already completed

	o.logger.WithField("jobId", jobID).Info("Job cancelled")

	return nil
}

// QueryDomainResource queries a domain-specific resource
func (o *Orchestrator) QueryDomainResource(ctx context.Context, domainName, resourceType, resourceID string) (interface{}, error) {
	o.mutex.RLock()
	domainAdapter, exists := o.domain[domainName]
	o.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("domain %s not registered", domainName)
	}

	return domainAdapter.QueryResource(resourceType, resourceID)
}

// monitorJobProgress periodically updates job progress information
func (o *Orchestrator) monitorJobProgress() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.updateAllJobProgress()
		}
	}
}

// updateAllJobProgress updates progress information for all active jobs
func (o *Orchestrator) updateAllJobProgress() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	now := time.Now()

	// Iterate through job registry
	for jobID, jobInfo := range o.jobRegistry {
		// Skip completed jobs
		if jobInfo.job.Status == stream.JobCompleted ||
			jobInfo.job.Status == stream.JobFailed ||
			jobInfo.job.Status == stream.JobCancelled {
			continue
		}

		// Get status info
		statusInfo, exists := o.jobStatusInfo[jobID]
		if !exists {
			continue
		}

		// Update elapsed time
		elapsed := now.Sub(jobInfo.startTime)
		statusInfo.AddMetric("elapsedSeconds", elapsed.Seconds())

		// Check for timeout
		if jobInfo.job.MaxDuration > 0 {
			maxDuration := time.Duration(jobInfo.job.MaxDuration) * time.Second
			if elapsed > maxDuration {
				jobInfo.job.Status = stream.JobFailed
				jobInfo.job.Error = "job exceeded maximum duration"
				jobInfo.job.CompletedAt = now
				statusInfo.Status = stream.JobFailed
				statusInfo.CurrentStep = "timed out"
				o.logger.WithField("jobId", jobID).Warn("Job timed out")
				continue
			}

			// Estimate time left
			if statusInfo.Progress > 0 {
				estimatedTotal := elapsed.Seconds() / statusInfo.Progress
				timeLeft := time.Duration((estimatedTotal - elapsed.Seconds()) * float64(time.Second))
				if timeLeft > 0 {
					statusInfo.SetEstimatedTimeLeft(timeLeft)
				}
			}
		}

		// Update metrics based on processing log
		if jobInfo.streamData != nil && len(jobInfo.streamData.ProcessingLog) > 0 {
			lastEntry := jobInfo.streamData.ProcessingLog[len(jobInfo.streamData.ProcessingLog)-1]
			statusInfo.CurrentStep = fmt.Sprintf("%s:%s", lastEntry.Processor, lastEntry.Operation)
		}
	}
}

// processResult handles the final result of a job
func (o *Orchestrator) processResult(jobID string, result map[string]interface{}, err error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	jobInfo, exists := o.jobRegistry[jobID]
	if !exists {
		o.logger.WithField("jobId", jobID).Error("Job not found when processing result")
		return
	}

	// Update job with results or error
	if err != nil {
		jobInfo.job.Status = stream.JobFailed
		jobInfo.job.Error = err.Error()
	} else {
		jobInfo.job.Status = stream.JobCompleted
		jobInfo.job.Results = result
	}
	jobInfo.job.CompletedAt = time.Now()

	// Update status info
	if statusInfo, exists := o.jobStatusInfo[jobID]; exists {
		if err != nil {
			statusInfo.Status = stream.JobFailed
			statusInfo.CurrentStep = "failed"
		} else {
			statusInfo.Status = stream.JobCompleted
			statusInfo.Progress = 1.0
			statusInfo.CurrentStep = "completed"
		}
		statusInfo.UpdatedAt = time.Now()
	}

	// Log completion
	if err != nil {
		o.logger.WithFields(logrus.Fields{
			"jobId": jobID,
			"error": err.Error(),
		}).Error("Job failed")
	} else {
		o.logger.WithField("jobId", jobID).Info("Job completed successfully")
	}

	// Execute callback if specified
	if jobInfo.job.CallbackURL != "" {
		// In a real implementation, this would send an HTTP callback
		o.logger.WithFields(logrus.Fields{
			"jobId":       jobID,
			"callbackUrl": jobInfo.job.CallbackURL,
		}).Info("Would send callback notification")
	}
}

// createRoutingConfig creates a routing configuration based on job type
func (o *Orchestrator) createRoutingConfig(jobType string) *routingConfig {
	// Default config favors balanced approach
	config := &routingConfig{
		startLayer:      "context",
		processingFlow:  []string{"context", "reasoning", "intuition"},
		contextWeight:   0.4,
		reasoningWeight: 0.4,
		intuitionWeight: 0.2,
		feedbackEnabled: true,
		adaptiveRouting: o.config.AdaptiveWeights,
	}

	// Customize based on job type
	if strings.HasPrefix(jobType, "pattern_") {
		// Pattern recognition tasks favor intuition
		config.startLayer = "intuition"
		config.processingFlow = []string{"intuition", "context", "reasoning"}
		config.intuitionWeight = 0.6
		config.contextWeight = 0.3
		config.reasoningWeight = 0.1
	} else if strings.HasPrefix(jobType, "logic_") {
		// Logical tasks favor reasoning
		config.startLayer = "reasoning"
		config.processingFlow = []string{"reasoning", "context", "intuition"}
		config.reasoningWeight = 0.7
		config.contextWeight = 0.2
		config.intuitionWeight = 0.1
	} else if strings.HasPrefix(jobType, "knowledge_") {
		// Knowledge-intensive tasks favor context
		config.startLayer = "context"
		config.processingFlow = []string{"context", "reasoning", "intuition"}
		config.contextWeight = 0.7
		config.reasoningWeight = 0.2
		config.intuitionWeight = 0.1
	}

	return config
}

// newStreamRouter creates a new stream router for the orchestrator
func newStreamRouter(orch *Orchestrator) *streamRouter {
	bufferSize := 100 // Buffer size for channels

	router := &streamRouter{
		orchestrator:        orch,
		inputChan:           make(chan *stream.StreamData, bufferSize),
		outputChan:          make(chan *stream.StreamData, bufferSize),
		contextInputChan:    make(chan *stream.StreamData, bufferSize),
		contextOutputChan:   make(chan *stream.StreamData, bufferSize),
		reasoningInputChan:  make(chan *stream.StreamData, bufferSize),
		reasoningOutputChan: make(chan *stream.StreamData, bufferSize),
		intuitionInputChan:  make(chan *stream.StreamData, bufferSize),
		intuitionOutputChan: make(chan *stream.StreamData, bufferSize),
	}

	return router
}

// start begins the stream router operation
func (r *streamRouter) start() {
	// Start layer processors
	go r.processContextLayer()
	go r.processReasoningLayer()
	go r.processIntuitionLayer()

	// Start input router
	go r.routeInput()

	// Start output collector
	go r.collectOutput()
}

// routeInput routes incoming stream data to the appropriate layer
func (r *streamRouter) routeInput() {
	for {
		select {
		case <-r.orchestrator.ctx.Done():
			return

		case data := <-r.inputChan:
			// Get job info
			jobID, _ := data.Metadata["jobId"].(string)
			r.orchestrator.mutex.RLock()
			jobInfo, exists := r.orchestrator.jobRegistry[jobID]
			r.orchestrator.mutex.RUnlock()

			if !exists {
				r.orchestrator.logger.WithField("jobId", jobID).Error("Job not found when routing input")
				continue
			}

			// Route based on the routing config
			startLayer := jobInfo.routingConfig.startLayer
			switch startLayer {
			case "context":
				r.contextInputChan <- data
			case "reasoning":
				r.reasoningInputChan <- data
			case "intuition":
				r.intuitionInputChan <- data
			default:
				r.contextInputChan <- data
			}
		}
	}
}

// collectOutput collects and processes results from all layers
func (r *streamRouter) collectOutput() {
	// Map to track which layers have processed each job
	processedLayers := make(map[string]map[string]bool)
	layerResults := make(map[string]map[string]*stream.StreamData)

	for {
		select {
		case <-r.orchestrator.ctx.Done():
			return

		case data := <-r.contextOutputChan:
			r.handleLayerOutput(data, "context", processedLayers, layerResults)

		case data := <-r.reasoningOutputChan:
			r.handleLayerOutput(data, "reasoning", processedLayers, layerResults)

		case data := <-r.intuitionOutputChan:
			r.handleLayerOutput(data, "intuition", processedLayers, layerResults)
		}
	}
}

// handleLayerOutput processes output from a metacognitive layer
func (r *streamRouter) handleLayerOutput(
	data *stream.StreamData,
	layer string,
	processedLayers map[string]map[string]bool,
	layerResults map[string]map[string]*stream.StreamData,
) {
	// Get job info
	jobID, _ := data.Metadata["jobId"].(string)
	if jobID == "" {
		r.orchestrator.logger.Error("Stream data missing job ID")
		return
	}

	// Initialize tracking maps if needed
	if processedLayers[jobID] == nil {
		processedLayers[jobID] = make(map[string]bool)
	}
	if layerResults[jobID] == nil {
		layerResults[jobID] = make(map[string]*stream.StreamData)
	}

	// Mark this layer as processed and store result
	processedLayers[jobID][layer] = true
	layerResults[jobID][layer] = data

	// Get job routing config
	r.orchestrator.mutex.RLock()
	jobInfo, exists := r.orchestrator.jobRegistry[jobID]
	statusInfo, statusExists := r.orchestrator.jobStatusInfo[jobID]
	r.orchestrator.mutex.RUnlock()

	if !exists {
		r.orchestrator.logger.WithField("jobId", jobID).Error("Job not found when handling layer output")
		return
	}

	// Update job progress
	if statusExists {
		// Count processed layers
		processedCount := len(processedLayers[jobID])
		totalLayers := len(jobInfo.routingConfig.processingFlow)
		if totalLayers > 0 {
			progress := float64(processedCount) / float64(totalLayers)
			statusInfo.UpdateProgress(progress, fmt.Sprintf("processed_by_%s", layer))
		}
	}

	// Determine next layer to process
	routingConfig := jobInfo.routingConfig

	// Check if we should route to another layer
	nextLayer := r.determineNextLayer(jobID, layer, routingConfig, processedLayers[jobID])
	if nextLayer != "" {
		// Route to next layer
		switch nextLayer {
		case "context":
			r.contextInputChan <- data
		case "reasoning":
			r.reasoningInputChan <- data
		case "intuition":
			r.intuitionInputChan <- data
		}
		return
	}

	// If all layers are processed, merge results and finalize
	if r.allLayersProcessed(jobID, routingConfig, processedLayers[jobID]) {
		// Merge results from all layers
		finalResult := r.mergeLayerResults(jobID, layerResults[jobID], routingConfig)

		// Update job with results
		result := finalResult.Data
		if result == nil {
			result = make(map[string]interface{})
		}

		// Add confidence as a result property
		confidence, _ := finalResult.Metadata["confidence"].(float64)
		result["confidence"] = confidence

		// Process the final result
		r.orchestrator.processResult(jobID, result, nil)

		// Clean up tracking maps
		delete(processedLayers, jobID)
		delete(layerResults, jobID)
	}
}

// determineNextLayer determines the next layer to process
func (r *streamRouter) determineNextLayer(
	jobID string,
	currentLayer string,
	routingConfig *routingConfig,
	processed map[string]bool,
) string {
	// Get the processing flow
	flow := routingConfig.processingFlow

	// Find current layer index
	currentIndex := -1
	for i, layer := range flow {
		if layer == currentLayer {
			currentIndex = i
			break
		}
	}

	// Find next unprocessed layer in the flow
	if currentIndex >= 0 && currentIndex < len(flow)-1 {
		for i := currentIndex + 1; i < len(flow); i++ {
			nextLayer := flow[i]

			// Skip if already processed or in skip list
			if processed[nextLayer] || contains(routingConfig.skipLayers, nextLayer) {
				continue
			}

			return nextLayer
		}
	}

	// If adaptive routing is enabled, check if we should skip any layer based on confidence
	if routingConfig.adaptiveRouting {
		// This would implement adaptive routing based on confidence levels
		// Not fully implemented in this example
	}

	// No next layer found
	return ""
}

// allLayersProcessed checks if all required layers have processed the data
func (r *streamRouter) allLayersProcessed(
	jobID string,
	routingConfig *routingConfig,
	processed map[string]bool,
) bool {
	for _, layer := range routingConfig.processingFlow {
		// Skip layers in the skip list
		if contains(routingConfig.skipLayers, layer) {
			continue
		}

		// If a required layer hasn't processed, return false
		if !processed[layer] {
			return false
		}
	}

	return true
}

// mergeLayerResults combines results from all layers based on weights
func (r *streamRouter) mergeLayerResults(
	jobID string,
	results map[string]*stream.StreamData,
	routingConfig *routingConfig,
) *stream.StreamData {
	// Start with the context layer result as the base
	var base *stream.StreamData
	if results["context"] != nil {
		base = results["context"].Clone()
	} else if results["reasoning"] != nil {
		base = results["reasoning"].Clone()
	} else if results["intuition"] != nil {
		base = results["intuition"].Clone()
	} else {
		// Should never happen if we've validated that all layers processed
		r.orchestrator.logger.WithField("jobId", jobID).Error("No layer results found when merging")
		return &stream.StreamData{
			ID:        fmt.Sprintf("merged-%s", jobID),
			Type:      "merged_result",
			Data:      make(map[string]interface{}),
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
		}
	}

	// Normalize weights
	totalWeight := routingConfig.contextWeight + routingConfig.reasoningWeight + routingConfig.intuitionWeight
	if totalWeight <= 0 {
		totalWeight = 1.0
	}

	contextWeight := routingConfig.contextWeight / totalWeight
	reasoningWeight := routingConfig.reasoningWeight / totalWeight
	intuitionWeight := routingConfig.intuitionWeight / totalWeight

	// Calculate confidence as weighted average
	confidence := 0.0
	if results["context"] != nil {
		if conf, ok := results["context"].Metadata["confidence"].(float64); ok {
			confidence += conf * contextWeight
		}
	}

	if results["reasoning"] != nil {
		if conf, ok := results["reasoning"].Metadata["confidence"].(float64); ok {
			confidence += conf * reasoningWeight
		}
	}

	if results["intuition"] != nil {
		if conf, ok := results["intuition"].Metadata["confidence"].(float64); ok {
			confidence += conf * intuitionWeight
		}
	}

	// Apply confidence booster if configured
	confidence += routingConfig.confidenceBooster * (1.0 - confidence)

	// Set final confidence
	base.Metadata["confidence"] = confidence

	// Add trace data
	base.Metadata["merged_result"] = true
	base.Metadata["merge_weights"] = map[string]float64{
		"context":   contextWeight,
		"reasoning": reasoningWeight,
		"intuition": intuitionWeight,
	}

	// Log the merge
	r.orchestrator.logger.WithFields(logrus.Fields{
		"jobId":      jobID,
		"confidence": confidence,
	}).Debug("Merged layer results")

	return base
}

// processContextLayer handles processing in the context layer
func (r *streamRouter) processContextLayer() {
	for {
		select {
		case <-r.orchestrator.ctx.Done():
			return

		case data := <-r.contextInputChan:
			// Start processing timer
			startTime := time.Now()

			// Process through context layer
			jobID, _ := data.Metadata["jobId"].(string)
			r.orchestrator.logger.WithField("jobId", jobID).Debug("Processing in context layer")

			// Clone the data for this layer
			processedData := data.Clone()

			// Record the processing in the log
			processedData.LogProcessing("context", "analyze", time.Since(startTime))

			// Add a random confidence for simulation
			confidence := 0.7 + (float64(time.Now().UnixNano()%30) / 100.0) // 0.7-1.0
			processedData.Metadata["confidence"] = confidence

			// Send to output channel
			r.contextOutputChan <- processedData
		}
	}
}

// processReasoningLayer handles processing in the reasoning layer
func (r *streamRouter) processReasoningLayer() {
	for {
		select {
		case <-r.orchestrator.ctx.Done():
			return

		case data := <-r.reasoningInputChan:
			// Start processing timer
			startTime := time.Now()

			// Process through reasoning layer
			jobID, _ := data.Metadata["jobId"].(string)
			r.orchestrator.logger.WithField("jobId", jobID).Debug("Processing in reasoning layer")

			// Clone the data for this layer
			processedData := data.Clone()

			// Record the processing in the log
			processedData.LogProcessing("reasoning", "deduce", time.Since(startTime))

			// Add a random confidence for simulation
			confidence := 0.8 + (float64(time.Now().UnixNano()%20) / 100.0) // 0.8-1.0
			processedData.Metadata["confidence"] = confidence

			// Send to output channel
			r.reasoningOutputChan <- processedData
		}
	}
}

// processIntuitionLayer handles processing in the intuition layer
func (r *streamRouter) processIntuitionLayer() {
	for {
		select {
		case <-r.orchestrator.ctx.Done():
			return

		case data := <-r.intuitionInputChan:
			// Start processing timer
			startTime := time.Now()

			// Process through intuition layer
			jobID, _ := data.Metadata["jobId"].(string)
			r.orchestrator.logger.WithField("jobId", jobID).Debug("Processing in intuition layer")

			// Clone the data for this layer
			processedData := data.Clone()

			// Record the processing in the log
			processedData.LogProcessing("intuition", "recognize", time.Since(startTime))

			// Add a random confidence for simulation
			confidence := 0.6 + (float64(time.Now().UnixNano()%40) / 100.0) // 0.6-1.0
			processedData.Metadata["confidence"] = confidence

			// Send to output channel
			r.intuitionOutputChan <- processedData
		}
	}
}

// Helper functions

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// NewMetacognitiveOrchestrator creates a new orchestrator
func NewMetacognitiveOrchestrator(config OrchestratorConfig) *MetacognitiveOrchestrator {
	orchestrator := &MetacognitiveOrchestrator{
		contextLayer:   NewContextLayer(config.Context),
		reasoningLayer: NewReasoningLayer(config.Reasoning),
		intuitionLayer: NewIntuitionLayer(config.Intuition),
		config:         config,
		state:          &OrchestratorState{},
		metrics:        &OrchestratorMetrics{
			ThroughputMetrics: make(map[string]float64),
			ErrorRates:        make(map[string]float64),
			LayerPerformance:  make(map[string]LayerMetrics),
		},
		contextChan:   make(chan *ContextResult, 100),
		reasoningChan: make(chan *ReasoningResult, 100),
		intuitionChan: make(chan *IntuitionResult, 100),
		errorChan:     make(chan error, 100),
		done:          make(chan struct{}),
	}
	
	return orchestrator
}

// Start starts the orchestrator
func (mo *MetacognitiveOrchestrator) Start(ctx context.Context) error {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	
	if mo.running {
		return fmt.Errorf("orchestrator is already running")
	}
	
	mo.running = true
	mo.state.ContextActive = true
	mo.state.ReasoningActive = true
	mo.state.IntuitionActive = true
	
	// Start background workers if concurrent processing is enabled
	if mo.config.ConcurrentProcessing {
		go mo.errorHandler(ctx)
		go mo.metricsCollector(ctx)
	}
	
	return nil
}

// Stop stops the orchestrator
func (mo *MetacognitiveOrchestrator) Stop() error {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	
	if !mo.running {
		return fmt.Errorf("orchestrator is not running")
	}
	
	mo.running = false
	close(mo.done)
	
	return nil
}

// Process processes data through the complete metacognitive pipeline
func (mo *MetacognitiveOrchestrator) Process(ctx context.Context, data *stream.StreamData) (*ProcessingResult, error) {
	startTime := time.Now()
	
	// Check if orchestrator is running
	mo.mu.RLock()
	if !mo.running {
		mo.mu.RUnlock()
		return nil, fmt.Errorf("orchestrator is not running")
	}
	mo.mu.RUnlock()
	
	// Create context with timeout
	processCtx, cancel := context.WithTimeout(ctx, mo.config.ProcessingTimeout)
	defer cancel()
	
	// Update state
	mo.updateState(data)
	
	var result *ProcessingResult
	var err error
	
	if mo.config.ConcurrentProcessing {
		result, err = mo.processConcurrent(processCtx, data)
	} else {
		result, err = mo.processSequential(processCtx, data)
	}
	
	if err != nil {
		mo.recordError(err)
		return nil, err
	}
	
	// Record metrics
	processingTime := time.Since(startTime)
	mo.recordMetrics(processingTime, result.Confidence)
	
	result.ProcessingTime = processingTime
	result.Timestamp = startTime
	
	return result, nil
}

// processSequential processes data through layers sequentially
func (mo *MetacognitiveOrchestrator) processSequential(ctx context.Context, data *stream.StreamData) (*ProcessingResult, error) {
	layerTiming := make(map[string]time.Duration)
	
	// 1. Context Layer Processing
	contextStart := time.Now()
	contextResult, err := mo.contextLayer.Process(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("context layer processing failed: %w", err)
	}
	layerTiming["context"] = time.Since(contextStart)
	
	// Check context confidence threshold
	if contextResult.Confidence < mo.config.MinContextConfidence {
		return nil, fmt.Errorf("context confidence %.2f below threshold %.2f", 
			contextResult.Confidence, mo.config.MinContextConfidence)
	}
	
	// 2. Reasoning Layer Processing
	reasoningStart := time.Now()
	reasoningResult, err := mo.reasoningLayer.Process(ctx, contextResult.Interpretation)
	if err != nil {
		return nil, fmt.Errorf("reasoning layer processing failed: %w", err)
	}
	layerTiming["reasoning"] = time.Since(reasoningStart)
	
	// Check reasoning confidence threshold
	if reasoningResult.Confidence < mo.config.MinReasoningConfidence {
		return nil, fmt.Errorf("reasoning confidence %.2f below threshold %.2f",
			reasoningResult.Confidence, mo.config.MinReasoningConfidence)
	}
	
	// 3. Intuition Layer Processing
	intuitionStart := time.Now()
	intuitionResult, err := mo.intuitionLayer.Process(ctx, contextResult.Interpretation, reasoningResult)
	if err != nil {
		return nil, fmt.Errorf("intuition layer processing failed: %w", err)
	}
	layerTiming["intuition"] = time.Since(intuitionStart)
	
	// Check intuition confidence threshold
	if intuitionResult.Confidence < mo.config.MinIntuitionConfidence {
		return nil, fmt.Errorf("intuition confidence %.2f below threshold %.2f",
			intuitionResult.Confidence, mo.config.MinIntuitionConfidence)
	}
	
	// 4. Aggregate Results
	finalResult := mo.aggregateResults(contextResult, reasoningResult, intuitionResult)
	finalResult.LayerTiming = layerTiming
	
	return finalResult, nil
}

// processConcurrent processes data through layers concurrently
func (mo *MetacognitiveOrchestrator) processConcurrent(ctx context.Context, data *stream.StreamData) (*ProcessingResult, error) {
	var wg sync.WaitGroup
	layerTiming := make(map[string]time.Duration)
	
	// Channel for collecting results
	results := make(chan interface{}, 3)
	errors := make(chan error, 3)
	
	// Start context layer (must complete before others can start)
	wg.Add(1)
	go func() {
		defer wg.Done()
		contextStart := time.Now()
		contextResult, err := mo.contextLayer.Process(ctx, data)
		layerTiming["context"] = time.Since(contextStart)
		
		if err != nil {
			errors <- fmt.Errorf("context layer: %w", err)
			return
		}
		
		if contextResult.Confidence < mo.config.MinContextConfidence {
			errors <- fmt.Errorf("context confidence %.2f below threshold", contextResult.Confidence)
			return
		}
		
		results <- contextResult
		
		// After context completes, start reasoning and intuition layers
		if mo.config.LayerSynchronization {
			wg.Add(2)
			
			// Reasoning layer
			go func() {
				defer wg.Done()
				reasoningStart := time.Now()
				reasoningResult, err := mo.reasoningLayer.Process(ctx, contextResult.Interpretation)
				layerTiming["reasoning"] = time.Since(reasoningStart)
				
				if err != nil {
					errors <- fmt.Errorf("reasoning layer: %w", err)
					return
				}
				
				if reasoningResult.Confidence < mo.config.MinReasoningConfidence {
					errors <- fmt.Errorf("reasoning confidence %.2f below threshold", reasoningResult.Confidence)
					return
				}
				
				results <- reasoningResult
			}()
			
			// Intuition layer (waits for reasoning to complete)
			go func() {
				defer wg.Done()
				
				// Wait for reasoning result
				var reasoningResult *ReasoningResult
				for i := 0; i < 2; i++ {
					select {
					case res := <-results:
						if rr, ok := res.(*ReasoningResult); ok {
							reasoningResult = rr
							break
						}
					case <-ctx.Done():
						errors <- ctx.Err()
						return
					}
				}
				
				intuitionStart := time.Now()
				intuitionResult, err := mo.intuitionLayer.Process(ctx, contextResult.Interpretation, reasoningResult)
				layerTiming["intuition"] = time.Since(intuitionStart)
				
				if err != nil {
					errors <- fmt.Errorf("intuition layer: %w", err)
					return
				}
				
				if intuitionResult.Confidence < mo.config.MinIntuitionConfidence {
					errors <- fmt.Errorf("intuition confidence %.2f below threshold", intuitionResult.Confidence)
					return
				}
				
				results <- intuitionResult
			}()
		}
	}()
	
	// Wait for all layers to complete
	wg.Wait()
	close(results)
	close(errors)
	
	// Check for errors
	select {
	case err := <-errors:
		return nil, err
	default:
	}
	
	// Collect results
	var contextResult *ContextResult
	var reasoningResult *ReasoningResult
	var intuitionResult *IntuitionResult
	
	for result := range results {
		switch r := result.(type) {
		case *ContextResult:
			contextResult = r
		case *ReasoningResult:
			reasoningResult = r
		case *IntuitionResult:
			intuitionResult = r
		}
	}
	
	// Ensure all results are present
	if contextResult == nil || reasoningResult == nil || intuitionResult == nil {
		return nil, fmt.Errorf("incomplete processing results")
	}
	
	// Aggregate results
	finalResult := mo.aggregateResults(contextResult, reasoningResult, intuitionResult)
	finalResult.LayerTiming = layerTiming
	
	return finalResult, nil
}

// aggregateResults aggregates results from all three layers
func (mo *MetacognitiveOrchestrator) aggregateResults(
	contextResult *ContextResult,
	reasoningResult *ReasoningResult,
	intuitionResult *IntuitionResult,
) *ProcessingResult {
	
	// Calculate weighted confidence
	weightSum := mo.config.ContextWeight + mo.config.ReasoningWeight + mo.config.IntuitionWeight
	if weightSum == 0 {
		weightSum = 3.0 // Default equal weights
		mo.config.ContextWeight = 1.0
		mo.config.ReasoningWeight = 1.0
		mo.config.IntuitionWeight = 1.0
	}
	
	aggregatedConfidence := (
		contextResult.Confidence*mo.config.ContextWeight +
		reasoningResult.Confidence*mo.config.ReasoningWeight +
		intuitionResult.Confidence*mo.config.IntuitionWeight,
	) / weightSum
	
	// Create final decision based on highest confidence layer
	var finalDecision interface{}
	var reasoning string
	
	if contextResult.Confidence >= reasoningResult.Confidence && contextResult.Confidence >= intuitionResult.Confidence {
		finalDecision = contextResult.Interpretation
		reasoning = "Context-driven decision based on knowledge integration"
	} else if reasoningResult.Confidence >= intuitionResult.Confidence {
		if len(reasoningResult.Conclusions) > 0 {
			finalDecision = reasoningResult.Conclusions[0].Value
		}
		reasoning = "Logic-driven decision based on rule application"
	} else {
		if len(intuitionResult.Predictions) > 0 {
			finalDecision = intuitionResult.Predictions[0].Value
		}
		reasoning = "Intuition-driven decision based on pattern recognition"
	}
	
	// Combine patterns and predictions
	patterns := intuitionResult.RecognizedPatterns
	predictions := intuitionResult.Predictions
	
	return &ProcessingResult{
		ContextResult:   contextResult,
		ReasoningResult: reasoningResult,
		IntuitionResult: intuitionResult,
		FinalDecision:   finalDecision,
		Confidence:      aggregatedConfidence,
		Reasoning:       reasoning,
		Patterns:        patterns,
		Predictions:     predictions,
	}
}

// updateState updates the orchestrator state
func (mo *MetacognitiveOrchestrator) updateState(data *stream.StreamData) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	
	mo.state.CurrentData = data
	mo.state.LastProcessed = time.Now()
	mo.state.ProcessingCount++
}

// recordError records an error
func (mo *MetacognitiveOrchestrator) recordError(err error) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	
	mo.state.ErrorCount++
	
	// Update error rates
	mo.metrics.mu.Lock()
	defer mo.metrics.mu.Unlock()
	
	errorType := "general"
	if err != nil {
		errorType = fmt.Sprintf("%T", err)
	}
	
	mo.metrics.ErrorRates[errorType]++
}

// recordMetrics records performance metrics
func (mo *MetacognitiveOrchestrator) recordMetrics(processingTime time.Duration, confidence float64) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	
	// Update state metrics
	mo.state.TotalProcessingTime += processingTime
	mo.state.AverageProcessingTime = mo.state.TotalProcessingTime / time.Duration(mo.state.ProcessingCount)
	
	// Update detailed metrics
	mo.metrics.mu.Lock()
	defer mo.metrics.mu.Unlock()
	
	mo.metrics.ProcessingTimes = append(mo.metrics.ProcessingTimes, processingTime)
	mo.metrics.ConfidenceScores = append(mo.metrics.ConfidenceScores, confidence)
	
	// Keep only last 1000 measurements to prevent memory growth
	if len(mo.metrics.ProcessingTimes) > 1000 {
		mo.metrics.ProcessingTimes = mo.metrics.ProcessingTimes[100:]
		mo.metrics.ConfidenceScores = mo.metrics.ConfidenceScores[100:]
	}
	
	// Calculate throughput (requests per second)
	if len(mo.metrics.ProcessingTimes) > 1 {
		recentTimes := mo.metrics.ProcessingTimes[len(mo.metrics.ProcessingTimes)-10:]
		avgTime := time.Duration(0)
		for _, t := range recentTimes {
			avgTime += t
		}
		avgTime /= time.Duration(len(recentTimes))
		
		if avgTime > 0 {
			mo.metrics.ThroughputMetrics["requests_per_second"] = float64(time.Second) / float64(avgTime)
		}
	}
}

// errorHandler handles errors in background
func (mo *MetacognitiveOrchestrator) errorHandler(ctx context.Context) {
	for {
		select {
		case err := <-mo.errorChan:
			mo.recordError(err)
		case <-ctx.Done():
			return
		case <-mo.done:
			return
		}
	}
}

// metricsCollector collects metrics in background
func (mo *MetacognitiveOrchestrator) metricsCollector(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mo.collectSystemMetrics()
		case <-ctx.Done():
			return
		case <-mo.done:
			return
		}
	}
}

// collectSystemMetrics collects system-wide metrics
func (mo *MetacognitiveOrchestrator) collectSystemMetrics() {
	mo.metrics.mu.Lock()
	defer mo.metrics.mu.Unlock()
	
	// Calculate average confidence over last minute
	if len(mo.metrics.ConfidenceScores) > 0 {
		recent := mo.metrics.ConfidenceScores
		if len(recent) > 60 {
			recent = recent[len(recent)-60:]
		}
		
		sum := 0.0
		for _, conf := range recent {
			sum += conf
		}
		mo.metrics.ThroughputMetrics["avg_confidence"] = sum / float64(len(recent))
	}
}

// GetMetrics returns current metrics
func (mo *MetacognitiveOrchestrator) GetMetrics() OrchestratorMetrics {
	mo.metrics.mu.RLock()
	defer mo.metrics.mu.RUnlock()
	
	// Create a copy to avoid data races
	metrics := OrchestratorMetrics{
		ProcessingTimes:   make([]time.Duration, len(mo.metrics.ProcessingTimes)),
		ConfidenceScores:  make([]float64, len(mo.metrics.ConfidenceScores)),
		ThroughputMetrics: make(map[string]float64),
		ErrorRates:        make(map[string]float64),
		LayerPerformance:  make(map[string]LayerMetrics),
	}
	
	copy(metrics.ProcessingTimes, mo.metrics.ProcessingTimes)
	copy(metrics.ConfidenceScores, mo.metrics.ConfidenceScores)
	
	for k, v := range mo.metrics.ThroughputMetrics {
		metrics.ThroughputMetrics[k] = v
	}
	
	for k, v := range mo.metrics.ErrorRates {
		metrics.ErrorRates[k] = v
	}
	
	for k, v := range mo.metrics.LayerPerformance {
		metrics.LayerPerformance[k] = v
	}
	
	return metrics
}

// GetState returns current state
func (mo *MetacognitiveOrchestrator) GetState() OrchestratorState {
	mo.mu.RLock()
	defer mo.mu.RUnlock()
	
	return *mo.state
}

// IsRunning returns whether the orchestrator is running
func (mo *MetacognitiveOrchestrator) IsRunning() bool {
	mo.mu.RLock()
	defer mo.mu.RUnlock()
	
	return mo.running
}

