package metacognitive

import (
	"context"

	"github.com/yourusername/izinyoka/internal/stream"
)

// ContextProcessor defines the interface for the context layer
type ContextProcessor interface {
	// Process transforms input data into contextual information
	Process(ctx context.Context, in <-chan stream.StreamData) <-chan stream.StreamData
}

// ReasoningProcessor defines the interface for the reasoning layer
type ReasoningProcessor interface {
	// Process performs logical processing on contextualized data
	Process(ctx context.Context, in <-chan stream.StreamData) <-chan stream.StreamData
}

// IntuitionProcessor defines the interface for the intuition layer
type IntuitionProcessor interface {
	// Process applies pattern recognition and heuristic reasoning
	Process(ctx context.Context, in <-chan stream.StreamData) <-chan stream.StreamData
}
