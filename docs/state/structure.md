# Project Structure

This document outlines the organization of the Izinyoka codebase to help developers navigate and contribute to the project.

## Root Directory Structure

```
izinyoka/
├── cmd/                     # Command-line applications
├── internal/                # Private application code
├── pkg/                     # Public libraries that can be imported by other projects
├── api/                     # API definitions and documentation
├── config/                  # Configuration files
├── data/                    # Data files used by the application
├── docs/                    # Documentation
├── scripts/                 # Utility scripts
├── test/                    # Test files and resources
├── tools/                   # Development tools
├── web/                     # Web interface assets
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── Dockerfile               # Docker build definition
├── .env.example             # Example environment variables
├── README.md                # Project overview
└── LICENSE                  # License information
```

## Core Packages

### cmd/

```
cmd/
└── izinyoka/
    └── main.go              # Main application entry point
```

### internal/

```
internal/
├── metacognitive/           # Metacognitive orchestrator components
│   ├── context/             # Context layer implementation
│   ├── reasoning/           # Reasoning layer implementation
│   ├── intuition/           # Intuition layer implementation
│   └── orchestrator.go      # Orchestrator implementation
├── metabolic/               # Metabolic-inspired components
│   ├── glycolytic/          # Glycolytic cycle implementation
│   ├── dreaming/            # Dreaming module implementation
│   ├── lactate/             # Lactate cycle implementation
│   └── manager.go           # Metabolic component manager
├── knowledge/               # Knowledge base implementation
│   ├── repository.go        # Knowledge repository
│   ├── query.go             # Knowledge query mechanisms
│   └── update.go            # Knowledge update mechanisms
├── stream/                  # Streaming infrastructure
│   ├── processor.go         # Stream processor interface
│   ├── data.go              # Stream data structures
│   └── pipeline.go          # Pipeline construction
├── domain/                  # Domain-specific implementations
│   ├── genomic/             # Genomic variant calling domain
│   └── common.go            # Common domain interfaces
└── server/                  # Server implementation for API
```

### pkg/

```
pkg/
├── metrics/                 # Metrics collection and reporting
├── config/                  # Configuration management
└── utils/                   # Utility functions and helpers
```

## Data Organization

```
data/
├── knowledge/               # Knowledge base storage
│   ├── models/              # Learned models
│   ├── patterns/            # Recognized patterns
│   └── heuristics/          # Heuristic rules
├── reference/               # Reference data for domains
│   └── genomic/             # Genomic reference data
└── results/                 # Processing results
```

## Configuration

```
config/
├── system.yaml              # Core system configuration
├── domains/                 # Domain-specific configurations
│   ├── genomic.yaml         # Genomic variant calling configuration
│   └── template.yaml        # Template for new domains
└── logging.yaml             # Logging configuration
```

## Documentation

```
docs/
├── architecture/            # Architecture documentation
│   ├── metacognitive.md     # Metacognitive layer details
│   ├── metabolic.md         # Metabolic components details
│   └── streaming.md         # Streaming implementation details
├── api/                     # API documentation
├── state/                   # Setup and structure documentation
├── development/             # Development guides
└── examples/                # Usage examples
```

## Key Interfaces

### StreamProcessor

The core interface that defines how components process streaming data:

```go
type StreamProcessor interface {
    Process(ctx context.Context, in <-chan StreamData) <-chan StreamData
}
```

### DomainAdapter

Interface for implementing domain-specific processing:

```go
type DomainAdapter interface {
    Initialize(config *DomainConfig) error
    GetContextProcessor() metacognitive.ContextProcessor
    GetReasoningProcessor() metacognitive.ReasoningProcessor
    GetIntuitionProcessor() metacognitive.IntuitionProcessor
    GetDreamGenerator() metabolic.DreamGenerator
}
```

### KnowledgeRepository

Interface for knowledge storage and retrieval:

```go
type KnowledgeRepository interface {
    Query(query *KnowledgeQuery) ([]KnowledgeItem, error)
    Store(item KnowledgeItem) error
    Update(id string, item KnowledgeItem) error
    Delete(id string) error
}
```

## Development Workflow

The codebase follows a modular design where:

1. New domain implementations can be added in `internal/domain/`
2. Core functionality enhancements should be made in the appropriate package under `internal/`
3. Public APIs and utilities should be placed in `pkg/`
4. Command-line tools and applications should be added under `cmd/`

Each component should implement the appropriate interfaces to maintain compatibility with the streaming architecture and metacognitive framework.
