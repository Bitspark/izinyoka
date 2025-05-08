package stream

import (
	"context"
	"sync"
)

// Pipeline represents a complete data processing pipeline
type Pipeline struct {
	processors []StreamProcessor
	source     <-chan StreamData
	output     <-chan StreamData
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewPipeline creates a new pipeline from a source and processors
func NewPipeline(source <-chan StreamData, processors ...StreamProcessor) *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pipeline{
		processors: processors,
		source:     source,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start begins the pipeline processing
func (p *Pipeline) Start() <-chan StreamData {
	current := p.source

	for _, processor := range p.processors {
		current = processor.Process(p.ctx, current)
	}

	p.output = current
	return p.output
}

// Output returns the output channel of the pipeline
func (p *Pipeline) Output() <-chan StreamData {
	return p.output
}

// Stop halts the pipeline processing
func (p *Pipeline) Stop() {
	p.cancel()
}

// Fork creates a split in the pipeline, sending data to multiple processors
func Fork(in <-chan StreamData, n int) []<-chan StreamData {
	outputs := make([]chan StreamData, n)
	for i := 0; i < n; i++ {
		outputs[i] = make(chan StreamData)
	}

	go func() {
		defer func() {
			for _, out := range outputs {
				close(out)
			}
		}()

		for data := range in {
			for _, out := range outputs {
				// Create a copy to avoid race conditions
				dataCopy := data
				out <- dataCopy
			}
		}
	}()

	// Convert to read-only channels
	readOnly := make([]<-chan StreamData, n)
	for i, ch := range outputs {
		readOnly[i] = ch
	}

	return readOnly
}

// Merge combines multiple input streams into a single output stream
func Merge(ctx context.Context, inputs ...<-chan StreamData) <-chan StreamData {
	out := make(chan StreamData)
	var wg sync.WaitGroup

	// Start a goroutine for each input channel
	for _, input := range inputs {
		wg.Add(1)
		go func(in <-chan StreamData) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case data, ok := <-in:
					if !ok {
						return
					}
					select {
					case out <- data:
					case <-ctx.Done():
						return
					}
				}
			}
		}(input)
	}

	// Start a goroutine to close the output channel when all inputs are done
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
