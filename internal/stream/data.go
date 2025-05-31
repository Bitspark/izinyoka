package stream

import (
	"fmt"
	"time"
)

// StreamData represents a unit of data flowing through the processing pipeline
type StreamData struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	Source        string                 `json:"source,omitempty"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	ProcessingLog []ProcessingLogEntry   `json:"processingLog,omitempty"`
}

// ProcessingLogEntry records a processing step performed on the data
type ProcessingLogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Processor string                 `json:"processor"`
	Operation string                 `json:"operation"`
	Duration  time.Duration          `json:"duration"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
	NextStage string                 `json:"nextStage,omitempty"`
	ErrorMsg  string                 `json:"errorMsg,omitempty"`
}

// Job represents a processing job submitted to the system
type Job struct {
	ID          string                 `json:"id"`
	InputData   map[string]interface{} `json:"inputData"`
	ProcessType string                 `json:"processType"`
	Priority    int                    `json:"priority"`
	Status      JobStatus              `json:"status"`
	Results     map[string]interface{} `json:"results,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	StartedAt   time.Time              `json:"startedAt,omitempty"`
	CompletedAt time.Time              `json:"completedAt,omitempty"`
	MaxDuration int                    `json:"maxDuration,omitempty"` // in seconds
	CallbackURL string                 `json:"callbackUrl,omitempty"`
}

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobPending    JobStatus = "pending"
	JobProcessing JobStatus = "processing"
	JobCompleted  JobStatus = "completed"
	JobFailed     JobStatus = "failed"
	JobCancelled  JobStatus = "cancelled"
)

// JobStatusInfo provides detailed information about job status
type JobStatusInfo struct {
	JobID       string                 `json:"jobId"`
	Status      JobStatus              `json:"status"`
	Progress    float64                `json:"progress"`
	CurrentStep string                 `json:"currentStep,omitempty"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
	StartedAt   time.Time              `json:"startedAt,omitempty"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	EstTimeLeft string                 `json:"estimatedTimeLeft,omitempty"`
}

// NewStreamData creates a new stream data instance
func NewStreamData(dataType string, data map[string]interface{}) *StreamData {
	now := time.Now()
	id := generateID(dataType, now)

	return &StreamData{
		ID:        id,
		Type:      dataType,
		Data:      data,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
	}
}

// AddMetadata adds or updates metadata in the stream data
func (sd *StreamData) AddMetadata(key string, value interface{}) {
	if sd.Metadata == nil {
		sd.Metadata = make(map[string]interface{})
	}
	sd.Metadata[key] = value
}

// LogProcessing adds a processing log entry
func (sd *StreamData) LogProcessing(processor, operation string, duration time.Duration) {
	if sd.ProcessingLog == nil {
		sd.ProcessingLog = make([]ProcessingLogEntry, 0)
	}

	entry := ProcessingLogEntry{
		Timestamp: time.Now(),
		Processor: processor,
		Operation: operation,
		Duration:  duration,
	}

	sd.ProcessingLog = append(sd.ProcessingLog, entry)
}

// LogError adds an error to the processing log
func (sd *StreamData) LogError(processor, operation, errMsg string) {
	if sd.ProcessingLog == nil {
		sd.ProcessingLog = make([]ProcessingLogEntry, 0)
	}

	entry := ProcessingLogEntry{
		Timestamp: time.Now(),
		Processor: processor,
		Operation: operation,
		ErrorMsg:  errMsg,
	}

	sd.ProcessingLog = append(sd.ProcessingLog, entry)
}

// Clone creates a copy of the stream data with a new ID
func (sd *StreamData) Clone() *StreamData {
	// Create new data maps to avoid shared references
	newData := make(map[string]interface{})
	for k, v := range sd.Data {
		newData[k] = v
	}

	newMetadata := make(map[string]interface{})
	for k, v := range sd.Metadata {
		newMetadata[k] = v
	}

	// Clone processing log
	newLog := make([]ProcessingLogEntry, len(sd.ProcessingLog))
	copy(newLog, sd.ProcessingLog)

	// Create new instance with copied data
	clone := &StreamData{
		ID:            generateID(sd.Type, time.Now()),
		Type:          sd.Type,
		Data:          newData,
		Metadata:      newMetadata,
		CreatedAt:     time.Now(),
		Source:        sd.Source,
		CorrelationID: sd.CorrelationID,
		ProcessingLog: newLog,
	}

	return clone
}

// NewJob creates a new job instance
func NewJob(processType string, inputData map[string]interface{}, priority int) *Job {
	now := time.Now()
	id := generateJobID(processType, now)

	return &Job{
		ID:          id,
		InputData:   inputData,
		ProcessType: processType,
		Priority:    priority,
		Status:      JobPending,
		CreatedAt:   now,
	}
}

// SetCompleted marks a job as completed with results
func (j *Job) SetCompleted(results map[string]interface{}) {
	j.Status = JobCompleted
	j.Results = results
	j.CompletedAt = time.Now()
}

// SetFailed marks a job as failed with an error message
func (j *Job) SetFailed(errorMsg string) {
	j.Status = JobFailed
	j.Error = errorMsg
	j.CompletedAt = time.Now()
}

// SetCancelled marks a job as cancelled
func (j *Job) SetCancelled() {
	j.Status = JobCancelled
	j.CompletedAt = time.Now()
}

// NewJobStatusInfo creates a new job status info instance
func NewJobStatusInfo(jobID string, status JobStatus) *JobStatusInfo {
	return &JobStatusInfo{
		JobID:     jobID,
		Status:    status,
		Progress:  0.0,
		UpdatedAt: time.Now(),
		Metrics:   make(map[string]interface{}),
	}
}

// UpdateProgress updates the job progress
func (jsi *JobStatusInfo) UpdateProgress(progress float64, currentStep string) {
	jsi.Progress = progress
	jsi.CurrentStep = currentStep
	jsi.UpdatedAt = time.Now()
}

// AddMetric adds a metric to the job status
func (jsi *JobStatusInfo) AddMetric(key string, value interface{}) {
	if jsi.Metrics == nil {
		jsi.Metrics = make(map[string]interface{})
	}
	jsi.Metrics[key] = value
}

// SetEstimatedTimeLeft sets the estimated time remaining
func (jsi *JobStatusInfo) SetEstimatedTimeLeft(duration time.Duration) {
	jsi.EstTimeLeft = formatDuration(duration)
}

// Helper functions

// generateID creates an ID for stream data
func generateID(dataType string, timestamp time.Time) string {
	return dataType + "-" + timestamp.Format("20060102-150405-") + randomString(6)
}

// generateJobID creates an ID for a job
func generateJobID(processType string, timestamp time.Time) string {
	return "job-" + processType + "-" + timestamp.Format("20060102-150405-") + randomString(6)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond) // Ensure uniqueness
	}
	return string(result)
}

// formatDuration formats a duration in a human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	} else if d < time.Hour {
		minutes := d.Minutes()
		seconds := d.Seconds() - minutes*60
		return fmt.Sprintf("%.0f minutes %.0f seconds", minutes, seconds)
	} else {
		hours := d.Hours()
		minutes := d.Minutes() - hours*60
		return fmt.Sprintf("%.0f hours %.0f minutes", hours, minutes)
	}
}
