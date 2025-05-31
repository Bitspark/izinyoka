# Installation Guide

This document provides detailed instructions for installing Izinyoka in various environments.

## Standard Installation

### System Requirements

- Operating System: Linux (Ubuntu 20.04+, CentOS 8+), macOS 11+, or Windows 10+
- CPU: 4+ cores recommended
- Memory: 8GB minimum, 16GB+ recommended
- Disk: 20GB free space
- Go: version 1.20 or higher

### Installation Steps

1. Install Go:
   ```bash
   # For Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install -y golang-1.20
   
   # For macOS with Homebrew
   brew install go
   
   # For Windows
   # Download and install from https://golang.org/dl/
   ```

2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/izinyoka.git
   cd izinyoka
   ```

3. Build the application:
   ```bash
   go build -o bin/izinyoka cmd/izinyoka/main.go
   ```

4. Create necessary directories:
   ```bash
   mkdir -p data/knowledge data/reference data/results
   ```

5. Copy example configuration:
   ```bash
   cp config/system.example.yaml config/system.yaml
   cp .env.example .env
   ```

6. Edit configuration files as needed (see [Setup](setup.md) for details)

7. Initialize the system:
   ```bash
   ./bin/izinyoka --init
   ```

## Docker Installation

### Requirements

- Docker Engine 20.10+
- Docker Compose 2.0+ (optional, for multi-container deployments)

### Steps

1. Build the Docker image:
   ```bash
   docker build -t izinyoka:latest .
   ```

2. Create a data volume:
   ```bash
   docker volume create izinyoka-data
   ```

3. Run the container:
   ```bash
   docker run -d \
     --name izinyoka \
     -p 8080:8080 \
     -v izinyoka-data:/app/data \
     -v $(pwd)/config:/app/config \
     --env-file .env \
     izinyoka:latest
   ```

## Kubernetes Deployment

### Requirements

- Kubernetes 1.19+
- kubectl configured to access your cluster
- Helm 3+ (optional, for chart-based deployment)

### Steps

1. Create namespace:
   ```bash
   kubectl create namespace izinyoka
   ```

2. Apply configuration:
   ```bash
   kubectl create configmap izinyoka-config --from-file=config/ -n izinyoka
   ```

3. Create secrets:
   ```bash
   # Example for creating secrets from .env file
   kubectl create secret generic izinyoka-secrets \
     --from-env-file=.env -n izinyoka
   ```

4. Apply deployment:
   ```bash
   kubectl apply -f k8s/deployment.yaml -n izinyoka
   kubectl apply -f k8s/service.yaml -n izinyoka
   ```

5. Check deployment status:
   ```bash
   kubectl get pods -n izinyoka
   ```

## Development Environment

For development, we recommend using a more streamlined setup:

1. Install Go:
   ```bash
   # Install Go as described in the Standard Installation
   ```

2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/izinyoka.git
   cd izinyoka
   ```

3. Install development dependencies:
   ```bash
   # Install golangci-lint for code quality
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Install mockgen for generating mocks
   go install github.com/golang/mock/mockgen@latest
   ```

4. Set up pre-commit hooks:
   ```bash
   cp scripts/pre-commit .git/hooks/
   chmod +x .git/hooks/pre-commit
   ```

5. Run in development mode:
   ```bash
   go run cmd/izinyoka/main.go --dev
   ```

## Troubleshooting

### Common Issues

1. **Missing Go modules**
   ```bash
   go mod tidy
   ```

2. **Permission issues with data directory**
   ```bash
   sudo chown -R $(whoami) data/
   ```

3. **Port conflicts**
   If port 8080 is already in use, you can specify a different port:
   ```bash
   ./bin/izinyoka --port=8081
   ```

4. **Configuration errors**
   Validate your configuration file:
   ```bash
   ./bin/izinyoka --validate-config
   ```

### Getting Help

If you encounter issues not covered here:

1. Check the logs:
   ```bash
   ./bin/izinyoka --log-level=debug
   ```

2. Run diagnostics:
   ```bash
   ./bin/izinyoka --diagnostics
   ```

3. Open an issue on GitHub:
   https://github.com/yourusername/izinyoka/issues

## Updating

To update an existing installation:

1. Pull the latest changes:
   ```bash
   git pull
   ```

2. Rebuild the application:
   ```bash
   go build -o bin/izinyoka cmd/izinyoka/main.go
   ```

3. Run migrations (if necessary):
   ```bash
   ./bin/izinyoka --migrate
   ```

## Next Steps

- [Setup](setup.md) - Configure the system after installation
- [Project Structure](structure.md) - Understand the codebase organization
