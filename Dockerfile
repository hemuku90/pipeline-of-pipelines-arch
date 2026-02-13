# =============================================================================
# Multi-stage Docker build for production-grade Go application
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder - Compile the application
# -----------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# Install dependencies for building
RUN apk add --no-cache --virtual .build-deps \
    git \
    tar \
    ca-certificates

# Set working directory
WORKDIR /app

# Download Go modules (use cache for faster builds)
COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

# Copy source code
COPY . .

# Build the application
# -ldflags: Embed version info and disable symbol table for smaller binary
# -w: Disable DWARF debugging information
# -s: Strip symbol table
# CGO_ENABLED=0: Static linking for distroless image
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG GIT_COMMIT=unknown

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags "-X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -s -w" \
    -tags "osusergo netgo" \
    -o /app/server \
    -a \
    ./cmd/server

# -----------------------------------------------------------------------------
# Stage 2: Production image - Distroless base for security
# -----------------------------------------------------------------------------
# Using gcr.io/distroless/static:nonroot for minimal attack surface
FROM gcr.io/distroless/static:nonroot AS production

# Add CA certificates for HTTPS support
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Create non-root user and group for security
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Copy built binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/server /app/

# Copy config file if needed
COPY --chown=appuser:appgroup config.yaml /app/config.yaml 2>/dev/null || true

# Change to non-root user
USER appuser

# Expose application port
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Set environment variables
ENV APP_HOST=0.0.0.0 \
    APP_PORT=8080 \
    ENVIRONMENT=production \
    GIN_MODE=release

# Run the application
ENTRYPOINT ["./server"]

# =============================================================================
# Development stage (optional, for local development)
# =============================================================================
FROM golang:1.21-alpine AS development

# Install development tools
RUN apk add --no-cache git make vim curl wget

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Install Air for hot reload (optional)
RUN go install github.com/air-verse/air@latest

# Default command
CMD ["air", "-c", ".air.toml"]

# =============================================================================
# Test stage (for CI/CD pipeline)
# -----------------------------------------------------------------------------
FROM builder AS test-runner

# Run tests with coverage
RUN CGO_ENABLED=0 go test -v -cover -race -coverprofile=coverage.txt ./...

# =============================================================================
# Security Scanning stage (for CI/CD pipeline)
# -----------------------------------------------------------------------------
FROM alpine:3.19 AS security-scan

# Install security scanning tools
RUN apk add --no-cache \
    trivy \
    checksec \
    syft \
    grype

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Run Trivy vulnerability scan
RUN trivy image --exit-code 1 --severity HIGH,CRITICAL /app/server

# =============================================================================
# Documentation: Build Arguments
# -----------------------------------------------------------------------------
# --build-arg VERSION=1.0.0    # Set application version
# --build-arg GIT_COMMIT=abc123 # Set Git commit SHA
# --build-arg TARGETOS=linux    # Target OS (linux, darwin, windows)
# --build-arg TARGETARCH=amd64  # Target architecture (amd64, arm64)

# =============================================================================
# Documentation: How to Build
# -----------------------------------------------------------------------------
#
# Build for current platform:
#   docker build -t pipeline-arch:latest .
#
# Build for specific platform:
#   docker buildx build --platform linux/amd64,linux/arm64 -t pipeline-arch:latest .
#
# Build with version info:
#   docker build --build-arg VERSION=1.0.0 --build-arg GIT_COMMIT=$(git rev-parse HEAD) -t pipeline-arch:1.0.0 .
#
# Run locally:
#   docker run -p 8080:8080 -p 9090:9090 -e DATABASE_URL=postgres://... pipeline-arch:latest
#
# =============================================================================