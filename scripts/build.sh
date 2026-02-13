#!/bin/bash

# =============================================================================
# Build Script for Pipeline Architecture
# Builds the Go application and Docker image
# =============================================================================

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_NAME="pipeline-arch"
APP_VERSION="${VERSION:-dev}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}"
BUILD_TIME="${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        log_warn "Docker is not installed. Skipping Docker build."
        DOCKER_AVAILABLE=false
    else
        DOCKER_AVAILABLE=true
    fi

    log_info "Prerequisites check completed"
}

# Build Go application
build_go() {
    log_info "Building Go application..."

    cd "$SCRIPT_DIR/.."

    # Determine build tags
    BUILD_TAGS="osusergo netgo"
    if [[ "${CROSS_COMPILE:-}" == "true" ]]; then
        BUILD_TAGS="${BUILD_TAGS} netgo"
    fi

    # Build binary
    LDFLAGS="-s -w -X main.version=${APP_VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildTime=${BUILD_TIME}"

    log_info "Building binary with tags: ${BUILD_TAGS}"
    CGO_ENABLED=0 GOOS="${GOOS:-linux}" GOARCH="${GOARCH:-amd64}" \
        go build -ldflags "${LDFLAGS}" \
        -tags "${BUILD_TAGS}" \
        -o bin/server \
        ./cmd/server

    log_info "Binary built successfully"

    # Verify binary
    if [[ -f "bin/server" ]]; then
        log_info "Binary info:"
        file bin/server
        ls -lh bin/server
    else
        log_error "Binary not found"
        exit 1
    fi
}

# Build Docker image
build_docker() {
    if [[ "$DOCKER_AVAILABLE" == "false" ]]; then
        log_warn "Docker not available, skipping build"
        return 0
    fi

    log_info "Building Docker image..."

    cd "$SCRIPT_DIR/.."

    # Get image tags
    TAGS=()
    TAGS+=("${APP_NAME}:${APP_VERSION}")
    TAGS+=("${APP_NAME}:latest")

    # Add git short SHA tag if available
    if [[ "$GIT_COMMIT" != "unknown" ]]; then
        SHORT_SHA="${GIT_COMMIT:0:7}"
        TAGS+=("${APP_NAME}:${SHORT_SHA}")
    fi

    # Build each tag
    for tag in "${TAGS[@]}"; do
        log_info "Building ${tag}..."
        docker build \
            --build-arg VERSION="${APP_VERSION}" \
            --build-arg GIT_COMMIT="${GIT_COMMIT}" \
            --build-arg BUILD_TIME="${BUILD_TIME}" \
            -t "${tag}" \
            -f Dockerfile .

        if [[ $? -eq 0 ]]; then
            log_info "Successfully built ${tag}"
        else
            log_error "Failed to build ${tag}"
            exit 1
        fi
    done

    log_info "Docker images built successfully"
    docker images | grep "${APP_NAME}"
}

# Scan for vulnerabilities
scan_image() {
    if [[ "$DOCKER_AVAILABLE" == "false" ]]; then
        log_warn "Docker not available, skipping scan"
        return 0
    fi

    if ! command -v trivy &> /dev/null; then
        log_warn "Trivy not installed, skipping vulnerability scan"
        return 0
    fi

    log_info "Scanning image for vulnerabilities..."

    trivy image --exit-code 1 \
        --severity CRITICAL,HIGH \
        --format json \
        "${APP_NAME}:${APP_VERSION}" > trivy-report.json 2>&1

    if [[ $? -eq 0 ]]; then
        log_info "Vulnerability scan passed"
    else
        log_error "Vulnerability scan failed. See trivy-report.json for details"
        cat trivy-report.json
        exit 1
    fi
}

# Push Docker image
push_docker() {
    if [[ "$DOCKER_AVAILABLE" == "false" ]]; then
        log_warn "Docker not available, skipping push"
        return 0
    fi

    log_info "Pushing Docker image..."

    docker push "${APP_NAME}:${APP_VERSION}"
    docker push "${APP_NAME}:latest"

    log_info "Docker image pushed successfully"
}

# Run tests
run_tests() {
    log_info "Running tests..."

    cd "$SCRIPT_DIR/.."

    go test -v -race -cover ./...
}

# Build for multiple platforms
build_cross_platform() {
    log_info "Cross-compiling for multiple platforms..."

    cd "$SCRIPT_DIR/.."

    PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

    export CROSS_COMPILE=true

    for platform in "${PLATFORMS[@]}"; do
        OS=$(echo "$platform" | cut -d'/' -f1)
        ARCH=$(echo "$platform" | cut -d'/' -f2)

        log_info "Building for ${OS}/${ARCH}..."

        LDFLAGS="-s -w -X main.version=${APP_VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildTime=${BUILD_TIME}"

        CGO_ENABLED=0 GOOS="${OS}" GOARCH="${ARCH}" \
            go build -ldflags "${LDFLAGS}" \
            -o "bin/server-${OS}-${ARCH}" \
            ./cmd/server/

        log_info "Built binary: bin/server-${OS}-${ARCH}"
    done

    log_info "Cross-compilation complete"
    ls -lh bin/
}

# Main function
main() {
    cd "$SCRIPT_DIR/.."

    log_info "Starting build process..."
    log_info "Version: ${APP_VERSION}"
    log_info "Git Commit: ${GIT_COMMIT}"
    log_info "Build Time: ${BUILD_TIME}"

    # Parse arguments
    case "${1:-build}" in
        build)
            check_prerequisites
            build_go
            build_docker
            ;;
        binary)
            check_prerequisites
            build_go
            ;;
        docker)
            check_prerequisites
            build_docker
            ;;
        scan)
            check_prerequisites
            build_docker
            scan_image
            ;;
        push)
            check_prerequisites
            push_docker
            ;;
        test)
            run_tests
            ;;
        cross)
            check_prerequisites
            build_cross_platform
            ;;
        all)
            check_prerequisites
            build_go
            build_docker
            scan_image
            ;;
        *)
            echo "Usage: $0 [build|binary|docker|scan|push|test|cross|all]"
            echo ""
            echo "Commands:"
            echo "  build   - Build binary and Docker image (default)"
            echo "  binary  - Build Go binary only"
            echo "  docker  - Build Docker image only"
            echo "  scan    - Build and scan for vulnerabilities"
            echo "  push    - Push Docker image to registry"
            echo "  test    - Run tests"
            echo "  cross   - Cross-compile for multiple platforms"
            echo "  all     - Build, scan, and prepare for release"
            exit 1
            ;;
    esac

    log_info "Build process completed successfully!"
}

# Run main function with all arguments
main "$@"