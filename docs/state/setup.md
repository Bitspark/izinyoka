# Setup Guide for Izinyoka

This document provides detailed instructions for setting up the Izinyoka system for development and production use.

## Prerequisites

- Go 1.20 or later
- At least 8GB RAM for development, 16GB+ recommended for production
- Git
- Docker (optional, for containerized deployment)
- Redis (for distributed deployment)

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/izinyoka.git
cd izinyoka
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Environment Variables

Create a `.env` file in the project root with the following variables:

```
IZINYOKA_LOG_LEVEL=debug  # Options: debug, info, warn, error
IZINYOKA_KNOWLEDGE_PATH=./data/knowledge
IZINYOKA_DREAM_INTERVAL=10m  # Time between dream cycles
IZINYOKA_MAX_CONCURRENT_TASKS=8
IZINYOKA_CONTEXT_THRESHOLD=0.3  # Minimum confidence to forward partial results
```

### 4. Initialize Knowledge Base

For first-time setup, initialize the knowledge base:

```bash
go run cmd/izinyoka/main.go --init-knowledge
```

This will create the necessary directory structure and seed the initial knowledge repository.

## Configuration Options

### Core System Configuration

Edit `config/system.yaml` to adjust system parameters:

```yaml
metacognitive:
  context_layer:
    threshold: 0.3
    buffer_size: 256
  reasoning_layer:
    max_depth: 5
    timeout: 200ms
  intuition_layer:
    heuristic_weight: 0.7
    pattern_threshold: 0.5

metabolic:
  glycolytic:
    task_batch_size: 16
    resource_timeout: 500ms
  dreaming:
    interval: 10m
    diversity: 0.8
    duration: 2m
  lactate:
    max_incomplete_tasks: 100
    recycle_threshold: 0.4
```

### Domain-Specific Configuration

For domain-specific applications (e.g., genomic variant calling), edit the appropriate config file:

```yaml
# config/domains/genomic.yaml
reference_genome: ./data/reference/hg38.fa
min_read_quality: 20
variant_calling:
  confidence_threshold: 0.95
  min_coverage: 10
  structural_sensitivity: 0.8
```

## Running the System

### Development Mode

```bash
go run cmd/izinyoka/main.go --config=config/system.yaml --domain=genomic
```

### Production Mode

```bash
go build -o izinyoka cmd/izinyoka/main.go
./izinyoka --config=config/system.yaml --domain=genomic --production
```

### Containerized Deployment

Build the Docker image:

```bash
docker build -t izinyoka:latest .
```

Run the container:

```bash
docker run -p 8080:8080 -v /path/to/data:/app/data izinyoka:latest
```

## Monitoring

The system exposes metrics via Prometheus at `/metrics` endpoint. Major metrics include:

- `izinyoka_tasks_processed_total`: Counter of total tasks processed
- `izinyoka_processing_time_seconds`: Histogram of processing time
- `izinyoka_dream_cycles_total`: Counter of dream cycles executed
- `izinyoka_knowledge_size_bytes`: Gauge of knowledge base size
- `izinyoka_incomplete_tasks`: Gauge of tasks in lactate cycle

## Next Steps

- [Project Structure](structure.md) - Understand the codebase organization
- [Installation](installation.md) - Details on installing in various environments
