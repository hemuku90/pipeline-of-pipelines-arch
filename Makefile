# =============================================================================
# Makefile for Pipeline Architecture
# Build, test, and deployment targets
# =============================================================================

# Variables
APP_NAME := pipeline-arch
APP_VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(APP_VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)

# Go commands
GO := go
GOC := $(GO) build
GOT := $(GO) test
GOL := $(GO) lint
GOF := $(GO) fmt

# Docker
DOCKER := docker
DOCKER_BUILD := $(DOCKER) build
DOCKER_PUSH := $(DOCKER) push

# Kubernetes
KUBECTL := kubectl
KUSTOMIZE := kustomize

# Directories
BIN_DIR := bin
SRC_DIR := cmd/server
PKG_DIR := internal/...

# Default target
.PHONY: help
help:
	@echo "Pipeline Architecture Build Commands"
	@echo ""
	@echo "Build Targets:"
	@echo "  build        - Build the application binary"
	@echo "  build-docker - Build Docker image"
	@echo "  build-all    - Build binary and Docker image"
	@echo ""
	@echo "Test Targets:"
	@echo "  test         - Run unit and integration tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-coverage- Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Vet Go code"
	@echo ""
	@echo "Kubernetes:"
	@echo "  k8s-deploy-dev   - Deploy to dev environment"
	@echo "  k8s-deploy-staging - Deploy to staging"
	@echo "  k8s-deploy-prod  - Deploy to production"
	@echo "  k8s-diff         - Show K8s changes (dry-run)"
	@echo "  k8s-delete       - Delete all K8s resources"
	@echo ""
	@echo "Helm:"
	@echo "  helm-install     - Install Helm chart"
	@echo "  helm-upgrade     - Upgrade Helm release"
	@echo "  helm-delete      - Uninstall Helm release"
	@echo ""
	@echo "Utility:"
	@echo "  tidy             - Download and tidy Go modules"
	@echo "  clean            - Clean build artifacts"
	@echo "  run              - Run application locally"
	@echo "  run-docker       - Run Docker container locally"
	@echo "  deps             - Install dependencies"

# =============================================================================
# Build Targets
# =============================================================================

.PHONY: build
build: ## Build the application binary
	@echo "Building binary..."
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) ./cmd/server
	@echo "Binary built: $(BIN_DIR)/$(APP_NAME)"

.PHONY: build-docker
build-docker: ## Build Docker image
	@echo "Building Docker image..."
	$(DOCKER_BUILD) -t $(APP_NAME):$(APP_VERSION) -t $(APP_NAME):latest .
	@echo "Image built: $(APP_NAME):$(APP_VERSION)"

.PHONY: build-all
build-all: build build-docker ## Build binary and Docker image
	@echo "Build complete!"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BIN_DIR)
	rm -rf coverage.txt
	rm -rf $(APP_NAME)
	@echo "Clean complete!"

# =============================================================================
# Test Targets
# =============================================================================

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	$(GOT) -v -race ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	$(GOT) -v -race ./internal/... ./pkg/...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOT) -v ./tests/integration/...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOT) -coverprofile=coverage.txt -covermode=atomic ./...
	@echo ""
	@echo "Coverage Summary:"
	@go tool cover -func=coverage.txt | tail -5 || true

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "Running e2e tests..."
	@echo "Note: Requires running application and Kubernetes cluster"
	$(GOT) -v ./tests/e2e/...

# =============================================================================
# Code Quality
# =============================================================================

.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	$(GOF) run ./...
	$(GOF) ./...

.PHONY: vet
vet: ## Vet Go code
	@echo "Vetting Go code..."
	$(GO) vet ./...

.PHONY: tidy
tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	$(GO) mod tidy
	$(GO) mod verify

.PHONY: deps
deps: ## Install dependencies
	@echo "Installing dependencies..."
	$(GO) mod download
	@echo "Dependencies installed!"

# =============================================================================
# Run Targets
# =============================================================================

.PHONY: run
run: ## Run application locally
	@echo "Running application..."
	$(GO) run ./cmd/server

.PHONY: run-docker
run-docker: build-docker ## Run Docker container locally
	@echo "Running Docker container..."
	$(DOCKER) run -p 8080:8080 -p 9090:9090 \
		-e DATABASE_URL="" \
		$(APP_NAME):latest

# =============================================================================
# Kubernetes Targets
# =============================================================================

.PHONY: k8s-deploy-dev
k8s-deploy-dev: ## Deploy to dev environment
	@echo "Deploying to dev..."
	$(KUBECTL) kustomize k8s/overlays/dev | $(KUBECTL) apply -f -
	$(KUBECTL) rollout status deployment/pipeline-arch-dev -n app-dev

.PHONY: k8s-deploy-staging
k8s-deploy-staging: ## Deploy to staging environment
	@echo "Deploying to staging..."
	$(KUBECTL) kustomize k8s/overlays/staging | $(KUBECTL) apply -f -
	$(KUBECTL) rollout status deployment/pipeline-arch-staging -n app-staging

.PHONY: k8s-deploy-prod
k8s-deploy-prod: ## Deploy to production environment
	@echo "Deploying to production..."
	$(KUBECTL) kustomize k8s/overlays/prod | $(KUBECTL) apply -f -
	$(KUBECTL) rollout status deployment/pipeline-arch-prod -n app-prod

.PHONY: k8s-diff
k8s-diff: ## Show K8s changes (dry-run)
	@echo "Showing K8s diff..."
	@echo "=== Dev ==="
	$(KUBECTL) kustomize k8s/overlays/dev | $(KUBECTL) apply --dry-run=server -f -
	@echo ""
	@echo "=== Staging ==="
	$(KUBECTL) kustomize k8s/overlays/staging | $(KUBECTL) apply --dry-run=server -f -
	@echo ""
	@echo "=== Production ==="
	$(KUBECTL) kustomize k8s/overlays/prod | $(KUBECTL) apply --dry-run=server -f -

.PHONY: k8s-delete
k8s-delete: ## Delete all K8s resources
	@echo "Deleting K8s resources..."
	$(KUBECTL) kustomize k8s/overlays/dev | $(KUBECTL) delete -f -
	$(KUBECTL) kustomize k8s/overlays/staging | $(KUBECTL) delete -f -
	$(KUBECTL) kustomize k8s/overlays/prod | $(KUBECTL) delete -f -

.PHONY: k8s-logs
k8s-logs: ## View application logs
	@echo "Viewing logs..."
	$(KUBECTL) logs -l app=pipeline-arch -n app --tail=100 -f

.PHONY: k8s-status
k8s-status: ## Show K8s resource status
	@echo "Showing K8s resource status..."
	@echo "=== Pods ==="
	$(KUBECTL) get pods -l app=pipeline-arch -A
	@echo ""
	@echo "=== Deployments ==="
	$(KUBECTL) get deployment -l app=pipeline-arch -A
	@echo ""
	@echo "=== Services ==="
	$(KUBECTL) get svc -l app=pipeline-arch -A

# =============================================================================
# Helm Targets
# =============================================================================

.PHONY: helm-install
helm-install: ## Install Helm chart
	@echo "Installing Helm chart..."
	helm install $(APP_NAME) ./helm/app -f ./helm/app/values.yaml

.PHONY: helm-upgrade
helm-upgrade: ## Upgrade Helm release
	@echo "Upgrading Helm release..."
	helm upgrade $(APP_NAME) ./helm/app -f ./helm/app/values.yaml

.PHONY: helm-delete
helm-delete: ## Uninstall Helm release
	@echo "Uninstalling Helm release..."
	helm uninstall $(APP_NAME)

.PHONY: helm-template
helm-template: ## Render Helm templates (dry-run)
	@echo "Rendering Helm templates..."
	helm template $(APP_NAME) ./helm/app -f ./helm/app/values.yaml

.PHONY: helm-values
helm-values: ## Show Helm values
	@echo "Showing Helm values..."
	helm show values ./helm/app

# =============================================================================
# Docker Registry
# =============================================================================

.PHONY: docker-login
docker-login: ## Login to Docker registry
	@echo "Logging in to Docker registry..."
	$(DOCKER) login ghcr.io/$(shell echo $(GITHUB_REPOSITORY) | cut -d'/' -f1)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image..."
	$(DOCKER) tag $(APP_NAME):latest $(REGISTRY)/$(GITHUB_REPOSITORY):$(APP_VERSION)
	$(DOCKER) tag $(APP_NAME):latest $(REGISTRY)/$(GITHUB_REPOSITORY):latest
	$(DOCKER) push $(REGISTRY)/$(GITHUB_REPOSITORY):$(APP_VERSION)
	$(DOCKER) push $(REGISTRY)/$(GITHUB_REPOSITORY):latest

# =============================================================================
# Release
# =============================================================================

.PHONE: release
release: ## Create a new release
	@echo "Creating release v$(APP_VERSION)..."
	@echo "Please manually create the release on GitHub with the following:"
	@echo "  Tag: v$(APP_VERSION)"
	@echo "  Title: Release v$(APP_VERSION)"
	@echo "  Notes: See CHANGELOG.md"

# =============================================================================
# CI/CD
# =============================================================================

.PHONY: ci
ci: lint test tidy ## Run CI pipeline locally

.PHONY: cd
cd: build-docker docker-login docker-push ## Run CD pipeline locally

# =============================================================================
# Documentation
# =============================================================================

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go --output docs/; \
		echo "API docs generated in docs/"; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# =============================================================================
# Security
# =============================================================================

.PHONY: security-scan
security-scan: ## Run security scans
	@echo "Running security scans..."
	@if command -v trivy >/dev/null 2>&1; then \
		trivy fs . --severity CRITICAL,HIGH; \
	else \
		echo "trivy not found. Install from: https://github.com/aquasecurity/trivy"; \
	fi

.PHONY: dependency-check
dependency-check: ## Check for vulnerable dependencies
	@echo "Checking dependencies..."
	$(GO) list -m -json all | jq -r '.Dir + " " + (.Version // "unknown")' | while read dir version; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "Checking $$dir@$$version..."; \
		fi; \
	done
	@echo "Use 'go list -m all | nancy' for vulnerability scanning"

# =============================================================================
# Development
# =============================================================================

.PHONY: dev
dev: ## Run development server with hot reload
	@echo "Running development server with hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "air not found. Install with: go install github.com/air-verse/air@latest"; \
		echo "Falling back to 'go run'..."; \
		$(GO) run ./cmd/server; \
	fi

.PHONY: generate
generate: ## Run code generators
	@echo "Running code generators..."
	@echo "No generators configured"

.PHONY: mock
mock: ## Generate mocks
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/repository/user_repo.go -destination=mocks/repository/mock_user_repo.go; \
		echo "Mocks generated"; \
	else \
		echo "mockgen not found. Install with: go install go.uber.org/mock/mockgen@latest"; \
	fi

# =============================================================================
# Debug
# =============================================================================

.PHONY: version
version: ## Show version information
	@echo "Application: $(APP_NAME)"
	@echo "Version: $(APP_VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell $(GO) version)"

.PHONY: env
env: ## Show environment variables
	@echo "Environment:"
	@echo "  APP_NAME: $(APP_NAME)"
	@echo "  APP_VERSION: $(APP_VERSION)"
	@echo "  GIT_COMMIT: $(GIT_COMMIT)"
	@echo "  BUILD_TIME: $(BUILD_TIME)"
	@echo "  GO_VERSION: $(shell $(GO) version | cut -d' ' -f3)"
	@echo "  DOCKER_VERSION: $(shell $(DOCKER) version --format '{{.Server.Version}}' 2>/dev/null || echo 'not installed')"