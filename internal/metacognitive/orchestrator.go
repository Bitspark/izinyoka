package metacognitive

import (
	"context"
	"sync"

	"github.com/yourusername/izinyoka/internal/knowledge"
	"github.com/yourusername/izinyoka/internal/metabolic"
	"github.com/yourusername/izinyoka/internal/stream"
)

// Orchestrator manages the nested processing layers
type Orchestrator struct {
	contextLayer   ContextProcessor
	reasoningLayer ReasoningProcessor
	intuitionLayer IntuitionProcessor
	glycolytic     *metabolic.GlycolicCycle
	dreaming       *metabolic.DreamingModule
	lactateCycle   *metabolic.LactateCycle
	knowledge      *knowledge.KnowledgeBase
	mu             sync.RWMutex
}

// NewOrchestrator creates a new metacognitive orchestrator
func NewOrchestrator(
	kb *knowledge.KnowledgeBase,
	contextLayer ContextProcessor,
	reasoningLayer ReasoningProcessor,
	intuitionLayer IntuitionProcessor,
	glycolytic *metabolic.GlycolicCycle,
	dreaming *metabolic.DreamingModule,
	lactateCycle *metabolic.LactateCycle,
) *Orchestrator {
	return &Orchestrator{
		knowledge:      kb,
		contextLayer:   contextLayer,
		reasoningLayer: reasoningLayer,
		intuitionLayer: intuitionLayer,
		glycolytic:     glycolytic,
		dreaming:       dreaming,
		lactateCycle:   lactateCycle,
	}
}

// Process starts the streaming processing pipeline
func (o *Orchestrator) Process(
	ctx context.Context,
	input <-chan stream.StreamData,
) <-chan stream.StreamData {
	// Use the glycolytic cycle to break down and allocate tasks
	taskInput := o.glycolytic.AllocateTasks(ctx, input)

	// Context layer processing
	contextOut := o.contextLayer.Process(ctx, taskInput)

	// Reasoning layer processing
	reasoningOut := o.reasoningLayer.Process(ctx, contextOut)

	// Intuition layer processing
	intuitionOut := o.intuitionLayer.Process(ctx, reasoningOut)

	// Create a pipeline to handle the process
	pipeline := stream.NewPipeline(
		intuitionOut,
		stream.ProcessorFunc(func(ctx context.Context, in <-chan stream.StreamData) <-chan stream.StreamData {
			out := make(chan stream.StreamData)

			go func() {
				defer close(out)
				for {
					select {
					case <-ctx.Done():
						return
					case data, ok := <-in:
						if !ok {
							return
						}

						// Check if task is incomplete
						confidence := data.Confidence
						if confidence < 0.7 { // Threshold for completion
							// Store in lactate cycle for later processing
							o.lactateCycle.StoreIncompleteTask(data)
							continue
						}

						// Pass completed tasks through
						select {
						case out <- data:
						case <-ctx.Done():
							return
						}
					}
				}
			}()

			return out
		}),
	)

	return pipeline.Start()
}

// Shutdown gracefully terminates processing
func (o *Orchestrator) Shutdown(ctx context.Context) error {
	// Stop the dreaming module
	o.dreaming.Stop()

	// Wait for the glycolytic cycle to complete current tasks
	o.glycolytic.Shutdown(ctx)

	// Persist any incomplete tasks
	o.lactateCycle.PersistIncompleteTasks()

	return nil
}
