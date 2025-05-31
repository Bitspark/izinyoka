---
title: "Getting Started"
layout: default
---

# Getting Started with Izinyoka

This guide will help you install, configure, and run your first genomic variant calling analysis with Izinyoka's biomimetic metacognitive architecture.

---

## Quick Start

For the impatient, here's how to get Izinyoka running in 5 minutes:

```bash
# Clone and build
git clone https://github.com/fullscreen-triangle/izinyoka.git
cd izinyoka && go build -o bin/izinyoka cmd/izinyoka/main.go

# Initialize and run
./bin/izinyoka --init && ./bin/izinyoka --example-data
```

---

## Prerequisites

### System Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| **OS** | Linux, macOS, Windows | Ubuntu 20.04+ | Container support available |
| **CPU** | 4 cores | 8+ cores | Benefits from parallelization |
| **Memory** | 8 GB | 16+ GB | Depends on data size |
| **Storage** | 20 GB | 100+ GB | For reference data and results |
| **Go** | 1.20+ | 1.21+ | Required for building from source |

### Knowledge Prerequisites

**Genomics Background**:
- Basic understanding of genomic variant calling
- Familiarity with VCF format and variant types
- Knowledge of reference genomes and sequencing technologies

**Technical Skills**:
- Command line proficiency
- Basic understanding of containerization (Docker) - optional
- Go programming experience - only for development

---

## Installation Options

### Option 1: Pre-built Binaries (Easiest)

Download the latest release for your platform:

```bash
# Linux x86_64
wget https://github.com/fullscreen-triangle/izinyoka/releases/latest/download/izinyoka-linux-amd64.tar.gz
tar -xzf izinyoka-linux-amd64.tar.gz

# macOS (Intel)
wget https://github.com/fullscreen-triangle/izinyoka/releases/latest/download/izinyoka-darwin-amd64.tar.gz
tar -xzf izinyoka-darwin-amd64.tar.gz

# macOS (Apple Silicon)
wget https://github.com/fullscreen-triangle/izinyoka/releases/latest/download/izinyoka-darwin-arm64.tar.gz
tar -xzf izinyoka-darwin-arm64.tar.gz
```

### Option 2: Build from Source

```bash
# 1. Install Go (if not already installed)
# See: https://golang.org/doc/install

# 2. Clone repository
git clone https://github.com/fullscreen-triangle/izinyoka.git
cd izinyoka

# 3. Build the application
go mod download
go build -o bin/izinyoka cmd/izinyoka/main.go

# 4. Verify installation
./bin/izinyoka --version
```

### Option 3: Docker Container

```bash
# Pull the official image
docker pull ghcr.io/fullscreen-triangle/izinyoka:latest

# Or build locally
git clone https://github.com/fullscreen-triangle/izinyoka.git
cd izinyoka
docker build -t izinyoka:local .

# Run with Docker
docker run -v $(pwd)/data:/app/data izinyoka:latest --help
```

### Option 4: Package Managers

```bash
# Homebrew (macOS/Linux)
brew install fullscreen-triangle/tap/izinyoka

# Conda
conda install -c bioconda izinyoka

# Snap (Linux)
sudo snap install izinyoka
```

---

## Initial Setup

### 1. Initialize the System

```bash
# Create necessary directories and download reference data
./bin/izinyoka --init

# This creates:
# ├── data/
# │   ├── knowledge/     # Knowledge base storage
# │   ├── reference/     # Reference genomes and databases
# │   └── results/       # Analysis outputs
# ├── config/
# │   └── system.yaml    # Main configuration file
# └── logs/              # System logs
```

### 2. Configuration

Edit the main configuration file:

```bash
# Copy and edit the example configuration
cp config/system.example.yaml config/system.yaml
nano config/system.yaml  # or your preferred editor
```

**Basic Configuration**:

```yaml
# config/system.yaml
server:
  host: "localhost"
  port: 8080
  
knowledge:
  storagePath: "./data/knowledge"
  backupEnabled: true
  maxItems: 10000

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
    max_concurrent_tasks: 8
    resource_timeout: 500ms
  dreaming:
    enable_auto_dreaming: true
    interval: 10m
    diversity: 0.8
  lactate:
    max_incomplete_tasks: 100
    recycle_threshold: 0.4
```

### 3. Download Reference Data

```bash
# Download human reference genome (GRCh38)
./bin/izinyoka --download-reference hg38

# Download variant databases
./bin/izinyoka --download-databases \
  --dbsnp \
  --gnomad \
  --clinvar

# Verify downloads
./bin/izinyoka --verify-data
```

---

## Your First Analysis

### Example 1: Single Sample Variant Calling

Let's analyze a sample FASTQ file:

```bash
# 1. Prepare input data
mkdir -p analysis/sample001
cd analysis/sample001

# 2. Create sample configuration
cat > sample.yaml << EOF
sample:
  id: "sample001"
  type: "germline"
  files:
    fastq1: "../../data/examples/sample001_R1.fastq.gz"
    fastq2: "../../data/examples/sample001_R2.fastq.gz"
  reference: "hg38"

analysis:
  variant_calling:
    confidence_threshold: 0.95
    min_coverage: 10
  quality_control:
    adapter_trimming: true
    quality_threshold: 20
EOF

# 3. Run the analysis
../../bin/izinyoka analyze \
  --config sample.yaml \
  --output results/ \
  --threads 4
```

**Expected Output**:
```
[INFO] Starting Izinyoka analysis pipeline
[INFO] Initializing metacognitive orchestrator
[INFO] Loading knowledge base (2.1 GB)
[INFO] Starting quality control...
[INFO] Adapter trimming completed (98.5% reads retained)
[INFO] Mapping reads to reference...
[INFO] Context layer: Processing alignment evidence
[INFO] Reasoning layer: Applying variant calling rules
[INFO] Intuition layer: Pattern recognition on read patterns
[INFO] Variant calling completed: 4,234,567 variants called
[INFO] Confidence calibration: ECE = 0.012
[INFO] Results written to: results/sample001.vcf.gz
[INFO] Analysis completed in 2.3 minutes
```

### Example 2: Multi-Sample Population Analysis

```bash
# 1. Create population study configuration
cat > population.yaml << EOF
study:
  name: "pilot_cohort"
  type: "population"
  samples:
    - id: "sample001"
      files: { fastq1: "data/sample001_R1.fq.gz", fastq2: "data/sample001_R2.fq.gz" }
    - id: "sample002" 
      files: { fastq1: "data/sample002_R1.fq.gz", fastq2: "data/sample002_R2.fq.gz" }
    - id: "sample003"
      files: { fastq1: "data/sample003_R1.fq.gz", fastq2: "data/sample003_R2.fq.gz" }

analysis:
  joint_calling: true
  population_filters: true
  structural_variants: true
  
output:
  format: ["vcf", "json", "tsv"]
  annotations: ["clinical", "population", "functional"]
EOF

# 2. Run population analysis
./bin/izinyoka analyze-population \
  --config population.yaml \
  --output cohort_results/ \
  --parallel 8
```

### Example 3: API Usage

Start the Izinyoka server:

```bash
# Start the API server
./bin/izinyoka server --config config/system.yaml &

# Check server status
curl http://localhost:8080/health
```

Submit a job via API:

```bash
# Submit variant calling job
curl -X POST http://localhost:8080/api/v1/process/job \
  -H "Content-Type: application/json" \
  -d '{
    "processType": "genomic.variant_analysis",
    "inputData": {
      "sampleId": "sample001",
      "referenceId": "hg38",
      "confidenceLevel": 95,
      "includeFiltered": false
    },
    "priority": 5,
    "maxDuration": 300
  }'

# Response: {"message": "Job submitted successfully", "jobId": "job-abc123"}

# Check job status
curl http://localhost:8080/api/v1/process/job/job-abc123

# Get results when complete
curl http://localhost:8080/api/v1/process/job/job-abc123/results
```

---

## Understanding the Output

### 1. VCF File Structure

Izinyoka produces enhanced VCF files with additional metacognitive annotations:

```vcf
##fileformat=VCFv4.3
##source=Izinyoka-1.0.0
##reference=GRCh38
##INFO=<ID=IC,Number=1,Type=Float,Description="Izinyoka Confidence Score">
##INFO=<ID=CL,Number=3,Type=Float,Description="Context,Reasoning,Intuition Layer Scores">
##INFO=<ID=EU,Number=1,Type=Float,Description="Epistemic Uncertainty">
##INFO=<ID=AU,Number=1,Type=Float,Description="Aleatoric Uncertainty">
##FORMAT=<ID=IC,Number=1,Type=Float,Description="Sample-level Izinyoka Confidence">

#CHROM  POS     ID      REF ALT QUAL    FILTER  INFO                            FORMAT      sample001
chr1    100001  .       T   C   99.9    PASS    IC=0.987;CL=0.92,0.95,0.89;EU=0.02;AU=0.01  GT:IC   1/0:0.987
chr1    100050  .       G   A   85.3    PASS    IC=0.923;CL=0.88,0.91,0.97;EU=0.05;AU=0.03  GT:IC   0/1:0.923
```

**Key Annotations**:
- **IC**: Overall Izinyoka confidence score (0-1)
- **CL**: Individual layer confidence scores [Context, Reasoning, Intuition]
- **EU**: Epistemic uncertainty (model uncertainty)
- **AU**: Aleatoric uncertainty (data noise)

### 2. JSON Results Format

```json
{
  "sample_id": "sample001",
  "analysis_type": "germline_variant_calling",
  "izinyoka_version": "1.0.0",
  "processing_time": "2.3 minutes",
  "statistics": {
    "total_variants": 4234567,
    "snvs": 3891234,
    "indels": 343333,
    "structural_variants": 1234,
    "high_confidence": 4098765,
    "medium_confidence": 135802,
    "low_confidence": 0
  },
  "quality_metrics": {
    "expected_calibration_error": 0.012,
    "brier_score": 0.089,
    "mean_confidence": 0.967,
    "uncertainty_breakdown": {
      "epistemic": 0.045,
      "aleatoric": 0.023
    }
  },
  "layer_contributions": {
    "context": 0.42,
    "reasoning": 0.38,
    "intuition": 0.20
  },
  "metabolic_stats": {
    "glycolytic_efficiency": 0.89,
    "dreams_generated": 23,
    "lactate_recoveries": 7
  }
}
```

### 3. Analysis Report

Izinyoka generates comprehensive HTML reports:

```bash
# Generate detailed analysis report
./bin/izinyoka report \
  --vcf results/sample001.vcf.gz \
  --json results/sample001.json \
  --output results/sample001_report.html
```

The report includes:
- **Executive Summary**: Key findings and quality metrics
- **Variant Overview**: Counts by type and confidence
- **Quality Assessment**: Calibration plots and uncertainty analysis
- **Metacognitive Analysis**: Layer contribution breakdown
- **Clinical Annotations**: Actionable variants and interpretations

---

## Common Use Cases

### 1. Clinical Diagnostics

**Mendelian Disease Analysis**:

```bash
./bin/izinyoka analyze \
  --sample trio_proband.yaml \
  --mode clinical \
  --phenotype "intellectual_disability" \
  --inheritance "autosomal_recessive" \
  --gene-panel "id_panel_v2.bed"
```

**Cancer Genomics**:

```bash
./bin/izinyoka analyze \
  --tumor tumor_sample.yaml \
  --normal normal_sample.yaml \
  --mode somatic \
  --cancer-type "breast" \
  --actionable-only
```

### 2. Population Genomics

**Large Cohort Analysis**:

```bash
./bin/izinyoka analyze-cohort \
  --samples cohort_manifest.txt \
  --population EUR \
  --joint-calling \
  --batch-size 50 \
  --distributed
```

### 3. Research Applications

**Rare Variant Discovery**:

```bash
./bin/izinyoka analyze \
  --samples research_cohort.yaml \
  --mode discovery \
  --af-threshold 0.001 \
  --enable-dreaming \
  --structural-variants
```

---

## Performance Optimization

### 1. Hardware Configuration

**CPU Optimization**:
```yaml
# config/system.yaml
metabolic:
  glycolytic:
    max_concurrent_tasks: 16  # Number of CPU cores
    worker_scaling: "auto"
    cpu_affinity: true
```

**Memory Management**:
```yaml
knowledge:
  cache_size: "8GB"
  memory_mapping: true
  
stream:
  buffer_size: 1024
  memory_pool: true
```

### 2. Distributed Processing

**Multi-Node Setup**:

```bash
# Node 1 (Controller)
./bin/izinyoka server \
  --mode controller \
  --workers node2:8080,node3:8080

# Node 2 & 3 (Workers)  
./bin/izinyoka server \
  --mode worker \
  --controller node1:8080
```

### 3. Cloud Deployment

**AWS Batch Configuration**:

```yaml
# aws-batch.yaml
compute_environment:
  name: "izinyoka-compute"
  type: "MANAGED"
  instance_types: ["m5.4xlarge", "c5.4xlarge"]
  min_vcpus: 0
  max_vcpus: 1000

job_definition:
  name: "izinyoka-job"
  container:
    image: "ghcr.io/fullscreen-triangle/izinyoka:latest"
    vcpus: 16
    memory: 32768
```

---

## Monitoring and Debugging

### 1. Real-time Monitoring

```bash
# Start with monitoring enabled
./bin/izinyoka analyze \
  --config sample.yaml \
  --monitor \
  --prometheus-port 9090

# View metrics in browser
open http://localhost:3000  # Grafana dashboard
```

### 2. Debugging Issues

**Enable Debug Logging**:

```bash
export IZINYOKA_LOG_LEVEL=debug
./bin/izinyoka analyze --config sample.yaml
```

**Common Issues**:

| Problem | Symptom | Solution |
|---------|---------|----------|
| **Out of Memory** | Process killed | Reduce `max_concurrent_tasks` |
| **Slow Performance** | High processing time | Enable `cpu_affinity`, increase `buffer_size` |
| **Low Confidence** | Many uncertain variants | Check input quality, adjust thresholds |
| **Missing Reference** | File not found errors | Run `--download-reference` |

### 3. Performance Profiling

```bash
# Enable profiling
./bin/izinyoka analyze \
  --config sample.yaml \
  --profile cpu,memory \
  --profile-output profile/

# Analyze results
go tool pprof profile/cpu.prof
go tool pprof profile/memory.prof
```

---

## Next Steps

### 1. Explore Advanced Features

- **[Custom Domains](../domains/)**: Extend beyond genomics
- **[API Development](../implementation/)**: Build custom integrations  
- **[Mathematical Framework](../mathematics/)**: Understand the theory

### 2. Join the Community

- **GitHub Discussions**: Ask questions and share experiences
- **Issue Tracker**: Report bugs and request features
- **Contributing Guide**: Help improve Izinyoka

### 3. Training and Certification

- **Online Tutorials**: Interactive learning modules
- **Workshop Materials**: Hands-on training sessions
- **Certification Program**: Demonstrate expertise

---

## Getting Help

### Documentation
- **[Architecture](../architecture/)**: System design and components
- **[Experiments](../experiments/)**: Performance benchmarks and validation
- **[Implementation](../implementation/)**: Technical implementation details

### Community Support
- **GitHub Issues**: [Report bugs](https://github.com/fullscreen-triangle/izinyoka/issues)
- **Discussions**: [Ask questions](https://github.com/fullscreen-triangle/izinyoka/discussions)
- **Stack Overflow**: Tag questions with `izinyoka`

### Professional Support
- **Commercial License**: Enterprise features and support
- **Training Services**: Custom workshops and consulting
- **Development Services**: Custom implementation and integration

---

## What's Next?

Now that you have Izinyoka running, explore these topics:

1. **[Understand the Architecture](../architecture/)** - Learn how the metacognitive layers work
2. **[Dive into the Mathematics](../mathematics/)** - Explore the theoretical foundations
3. **[Review Experimental Results](../experiments/)** - See performance comparisons
4. **[Explore Domain Applications](../domains/)** - Learn about genomic variant calling specifics

<style>
.code-block {
  background: #f6f8fa;
  border: 1px solid #e1e4e8;
  border-radius: 6px;
  padding: 16px;
  margin: 16px 0;
  overflow-x: auto;
  font-family: 'SFMono-Regular', 'Consolas', 'Liberation Mono', 'Menlo', monospace;
}

.quickstart-box {
  background: #f1f8ff;
  border: 1px solid #c8e1ff;
  border-radius: 6px;
  padding: 20px;
  margin: 20px 0;
}

.warning-box {
  background: #fff8c5;
  border: 1px solid #ffdf5d;
  border-radius: 6px;
  padding: 16px;
  margin: 16px 0;
}

.info-box {
  background: #f6f8fa;
  border-left: 4px solid #0366d6;
  padding: 16px;
  margin: 16px 0;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin: 20px 0;
}

table th, table td {
  border: 1px solid #e1e4e8;
  padding: 8px 12px;
  text-align: left;
}

table th {
  background: #f6f8fa;
  font-weight: 600;
}

table tr:nth-child(even) {
  background: #f9f9f9;
}

.nav-pills {
  display: flex;
  gap: 10px;
  margin: 20px 0;
}

.nav-pill {
  background: #f6f8fa;
  border: 1px solid #e1e4e8;
  border-radius: 20px;
  padding: 8px 16px;
  text-decoration: none;
  color: #0366d6;
  font-weight: 500;
}

.nav-pill:hover {
  background: #e1f5fe;
  border-color: #0366d6;
}
</style> 