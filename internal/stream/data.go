package stream

import (
	"time"
)

// StreamData represents a chunk of data flowing through the system
type StreamData struct {
	// Content holds the actual data being processed
	Content interface{}

	// Confidence represents the certainty level of the data
	Confidence float64

	// Metadata contains additional information about the data
	Metadata map[string]interface{}

	// Timestamp when the data was created or last modified
	Timestamp time.Time

	// ID uniquely identifies this chunk of data
	ID string
}

// NewStreamData creates a new StreamData instance with defaults
func NewStreamData(content interface{}) StreamData {
	return StreamData{
		Content:    content,
		Confidence: 0.0,
		Metadata:   make(map[string]interface{}),
		Timestamp:  time.Now(),
		ID:         generateID(),
	}
}

// WithConfidence sets the confidence and returns the StreamData
func (d StreamData) WithConfidence(confidence float64) StreamData {
	d.Confidence = confidence
	return d
}

// WithMetadata adds metadata and returns the StreamData
func (d StreamData) WithMetadata(key string, value interface{}) StreamData {
	if d.Metadata == nil {
		d.Metadata = make(map[string]interface{})
	}
	d.Metadata[key] = value
	return d
}

// Merge combines two StreamData objects, preferring the higher confidence
func (d StreamData) Merge(other StreamData) StreamData {
	result := d

	if other.Confidence > d.Confidence {
		result.Content = other.Content
		result.Confidence = other.Confidence
	}

	// Merge metadata
	for k, v := range other.Metadata {
		result.Metadata[k] = v
	}

	// Use most recent timestamp
	if other.Timestamp.After(d.Timestamp) {
		result.Timestamp = other.Timestamp
	}

	return result
}

// Generate a unique ID for the stream data
func generateID() string {
	return time.Now().Format("20060102150405.000") + "-" + randomString(8)
}

// Generate a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}
