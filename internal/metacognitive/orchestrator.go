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
		Priority:    job.Priority,
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
