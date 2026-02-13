# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-01-15

### Added

- Initial production-grade Pipeline of Pipelines architecture
- Complete Go REST API application with CRUD operations
- Multi-stage Dockerfile with distroless base and security hardening
- Kubernetes manifests:
  - Deployment with security context, health checks, resource limits
  - Service (ClusterIP)
  - HorizontalPodAutoscaler
  - PodDisruptionBudget
  - NetworkPolicy
  - ConfigMap and Secret templates
  - ServiceAccount with RBAC
- Kustomize overlays for dev, staging, and production environments
- Helm chart with all templates
- GitHub Actions CI pipeline:
  - Code linting (golangci-lint, hadolint)
  - Unit and integration tests with coverage
  - Security scanning (Trivy, GoSec, Gitleaks)
  - Docker build and push
- GitHub Actions CD pipelines:
  - CI-CD-dev.yml: Auto-deploy to development
  - CI-CD-staging.yml: Semi-auto with QA verification
  - CI-CD-prod.yml: Manual approval required for production
- Human verification integration:
  - PR comments for deployment verification
  - Slack notifications for deployment status
- Unit tests with testify framework
- Makefile with comprehensive build commands
- Build and test scripts
- Comprehensive README documentation

### Features

- **Pipeline-of-Pipelines Pattern**: Parent pipeline orchestrates child pipelines per service
- **GitOps with ArgoCD**: Declarative GitOps approach for Kubernetes manifests
- **Human-in-the-Loop**: Verification gates for staging and production deployments
- **Multi-Environment**: Support for dev, staging, and production clusters
- **Security Hardened**: Non-root execution, read-only filesystem, network policies
- **Auto-Scaling**: HPA with CPU and memory metrics
- **Monitoring Ready**: Prometheus metrics endpoint, ServiceMonitor configuration
- **Production Ready**: PDB for controlled disruptions, graceful shutdown

### Technology Stack

- **Language**: Go 1.21
- **Framework**: chi/v5 for routing
- **Container**: Docker with multi-stage builds
- **Orchestration**: Kubernetes 1.28+
- **CI/CD**: GitHub Actions
- **GitOps**: ArgoCD with ApplicationSet
- **Monitoring**: Prometheus, Grafana
- **Helm**: Helm 3 charts

### Repository Structure

```
pipeline-of-pipelines-arch/
├── .github/workflows/     # GitHub Actions CI/CD
├── cmd/server/            # Application entrypoint
├── internal/              # Application code
│   ├── api/               # HTTP handlers & middleware
│   ├── config/            # Configuration loading
│   └── models/            # Data models
├── pkg/                   # Shared packages
├── tests/                 # Unit & integration tests
├── k8s/                   # Kubernetes manifests
│   ├── base/              # Base resources
│   └── overlays/          # Environment overlays (dev/staging/prod)
├── helm/                  # Helm charts
├── scripts/               # Build & deployment scripts
└── gitops/                # ArgoCD configurations
```

### Getting Started

```bash
# Clone the repository
git clone https://github.com/hemuku90/pipeline-of-pipelines-arch.git
cd pipeline-of-pipelines-arch

# Run locally
make run

# Run tests
make test

# Build Docker
make docker-build

# Deploy to dev
make k8s-deploy-dev
```

### GitOps Setup

```bash
# Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.8.4/manifests/install.yaml

# Apply ApplicationSet
kubectl apply -f gitops/argocd/application-set.yaml -n argocd
```

### CI/CD Pipeline Flow

```
Developer Push → CI Pipeline → Dev Deploy → QA Verify → Staging Deploy → 
Manual Approval → Production Deploy
     ↓                ↓            ↓            ↓            ↓              ↓
  Tests          Security       Auto       Comment        Auto         /approve
  Lint           Scan          Deploy    "/verify-stage" Deploy         required
```

### Security Considerations

- Secrets managed via External Secrets Operator
- Non-root container execution
- Read-only root filesystem
- Network policies restricting egress
- Regular dependency updates
- Security scanning in CI pipeline

### License

This project is licensed under the MIT License - see the LICENSE file for details.

### Support

For issues and feature requests, please use the GitHub issue tracker.

---

## [Unreleased]

### Planned

- Add integration tests with real database
- Add end-to-end tests
- Prometheus rules for alerting
- Grafana dashboards
- Backup and restore procedures
- Multi-region support
- Blue-green deployment strategy
- Canary deployment with Argo Rollouts