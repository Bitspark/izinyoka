package context

import (
	"context"
	"sync"

	"github.com/yourusername/izinyoka/internal/knowledge"
	"github.com/yourusername/izinyoka/internal/metacognitive"
	"github.com/yourusername/izinyoka/internal/stream"
)

// ContextLayer implements the context processing stage
type ContextLayer struct {
	knowledge *knowledge.KnowledgeBase
	buffer    []stream.StreamData
	threshold float64
	mu        sync.Mutex
}

// NewContextLayer creates a new context processing layer
func NewContextLayer(kb *knowledge.KnowledgeBase, threshold float64) *ContextLayer {
	return &ContextLayer{
		knowledge: kb,
		threshold: threshold,
		buffer:    make([]stream.StreamData, 0, 100),
	}
}

// Process implements metacognitive.ContextProcessor interface
func (cl *ContextLayer) Process(
	ctx context.Context,
	in <-chan stream.StreamData,
) <-chan stream.StreamData {
	out := make(chan stream.StreamData)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return

			case data, ok := <-in:
				if !ok {
					// Process remaining buffer on channel close
					if len(cl.buffer) > 0 {
						cl.mu.Lock()
						result := cl.processBuffer(cl.buffer, true)
						cl.buffer = cl.buffer[:0] // Clear buffer
						cl.mu.Unlock()

						out <- result
					}
					return
				}

				// Add to buffer
				cl.mu.Lock()
				cl.buffer = append(cl.buffer, data)

				// Process partial results if enough data available or buffer gets too large
				var partial stream.StreamData
				if len(cl.buffer) >= 10 || len(cl.buffer) > 100 {
					partial = cl.processBuffer(cl.buffer, false)
					// Keep a sliding window of the last 5 items
					if len(cl.buffer) > 5 {
						cl.buffer = cl.buffer[len(cl.buffer)-5:]
					}
				}
				cl.mu.Unlock()

				// Send partial results if confidence threshold met
				if partial.Content != nil && partial.Confidence >= cl.threshold {
					select {
					case out <- partial:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out
}

// processBuffer handles the processing of buffered data
func (cl *ContextLayer) processBuffer(
	buffer []stream.StreamData,
	isComplete bool,
) stream.StreamData {
	if len(buffer) == 0 {
		return stream.StreamData{}
	}

	// Extract domain context from the buffer
	contextData := extractContextFromData(buffer)

	// Enrich with knowledge from the knowledge base
	enrichedContext := cl.enrichWithKnowledge(contextData)

	// Calculate confidence based on data quantity and consistency
	confidence := calculateContextConfidence(buffer, isComplete)

	return stream.StreamData{
		Content:    enrichedContext,
		Confidence: confidence,
		Metadata: map[string]interface{}{
			"buffer_size":    len(buffer),
			"is_complete":    isComplete,
			"context_source": "context_layer",
		},
		Timestamp: buffer[len(buffer)-1].Timestamp,
		ID:        "ctx-" + buffer[len(buffer)-1].ID,
	}
}

// Extract context information from the data buffer
func extractContextFromData(buffer []stream.StreamData) map[string]interface{} {
	context := make(map[string]interface{})

	// Extract common patterns and metadata across buffer items
	for _, data := range buffer {
		for k, v := range data.Metadata {
			if existing, ok := context[k]; ok {
				// If key already exists, prefer the most frequent or recent value
				// This is a simplified approach - real implementation would be more sophisticated
				if data.Timestamp.After(buffer[0].Timestamp) {
					context[k] = v
				}
			} else {
				context[k] = v
			}
		}

		// Extract content-specific context
		// This would be domain-specific in a real implementation
	}

	return context
}

// Enrich the context with knowledge from the knowledge base
func (cl *ContextLayer) enrichWithKnowledge(contextData map[string]interface{}) map[string]interface{} {
	enriched := make(map[string]interface{})
	for k, v := range contextData {
		enriched[k] = v
	}

	// Query knowledge base for relevant context enrichment
	// This is a placeholder - real implementation would query based on contextData
	query := &knowledge.KnowledgeQuery{
		Type:       "context",
		Attributes: contextData,
		Limit:      10,
	}

	items, err := cl.knowledge.Query(query)
	if err == nil {
		for _, item := range items {
			// Add knowledge items to the context
			enriched["knowledge_"+item.ID] = item.Content
		}
	}

	return enriched
}

// Calculate confidence score for the context
func calculateContextConfidence(buffer []stream.StreamData, isComplete bool) float64 {
	if len(buffer) == 0 {
		return 0.0
	}

	// Base confidence from buffer size
	sizeConfidence := float64(len(buffer)) / 100.0
	if sizeConfidence > 1.0 {
		sizeConfidence = 1.0
	}

	// Average confidence of items in buffer
	var totalConfidence float64
	for _, data := range buffer {
		totalConfidence += data.Confidence
	}
	avgConfidence := totalConfidence / float64(len(buffer))

	// Combine factors (simple weighted average)
	confidence := (sizeConfidence * 0.3) + (avgConfidence * 0.7)

	// Higher confidence if this is a complete processing
	if isComplete {
		confidence += 0.1
		if confidence > 1.0 {
			confidence = 1.0
		}
	}

	return confidence
}

// Ensure ContextLayer implements ContextProcessor
var _ metacognitive.ContextProcessor = (*ContextLayer)(nil)
