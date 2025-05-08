# Build stage
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make build-base

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o izinyoka ./cmd/izinyoka

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user
RUN adduser -D -g '' izinyoka
USER izinyoka

# Set working directory
WORKDIR /home/izinyoka

# Copy the binary from builder
COPY --from=builder /app/izinyoka .

# Copy necessary configuration files
COPY --from=builder /app/config ./config

# Create directories for runtime data
RUN mkdir -p data/knowledge

# Set environment variables
ENV IZINYOKA_LOG_LEVEL=info \
    IZINYOKA_KNOWLEDGE_PATH=/home/izinyoka/data/knowledge

# Expose ports if your application listens on any
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

# Command to run when the container starts
ENTRYPOINT ["./izinyoka"]
CMD ["--config", "./config/config.yaml"]
