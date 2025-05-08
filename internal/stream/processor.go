package stream

import (
	"context"
)

// StreamProcessor defines the interface for each processing layer
type StreamProcessor interface {
	// Process takes a stream of input data and returns a stream of output data
	Process(ctx context.Context, in <-chan StreamData) <-chan StreamData
}

// ProcessorFunc is a function that implements StreamProcessor
type ProcessorFunc func(ctx context.Context, in <-chan StreamData) <-chan StreamData

// Process implements the StreamProcessor interface for ProcessorFunc
func (f ProcessorFunc) Process(ctx context.Context, in <-chan StreamData) <-chan StreamData {
	return f(ctx, in)
}

// Chain connects multiple processors in sequence
func Chain(processors ...StreamProcessor) StreamProcessor {
	return ProcessorFunc(func(ctx context.Context, in <-chan StreamData) <-chan StreamData {
		current := in
		for _, p := range processors {
			current = p.Process(ctx, current)
		}
		return current
	})
}

// Buffer creates a buffered processor that can hold up to bufferSize items
func Buffer(bufferSize int) StreamProcessor {
	return ProcessorFunc(func(ctx context.Context, in <-chan StreamData) <-chan StreamData {
		out := make(chan StreamData, bufferSize)

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
					select {
					case out <- data:
					case <-ctx.Done():
						return
					}
				}
			}
		}()

		return out
	})
}

// Filter creates a processor that only passes data meeting the condition
func Filter(condition func(StreamData) bool) StreamProcessor {
	return ProcessorFunc(func(ctx context.Context, in <-chan StreamData) <-chan StreamData {
		out := make(chan StreamData)

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
					if condition(data) {
						select {
						case out <- data:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()

		return out
	})
}

// Map creates a processor that transforms each data item
func Map(transform func(StreamData) StreamData) StreamProcessor {
	return ProcessorFunc(func(ctx context.Context, in <-chan StreamData) <-chan StreamData {
		out := make(chan StreamData)

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
					transformed := transform(data)
					select {
					case out <- transformed:
					case <-ctx.Done():
						return
					}
				}
			}
		}()

		return out
	})
}
