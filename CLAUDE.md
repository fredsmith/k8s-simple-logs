# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

k8s-simple-logs is a simple Kubernetes log aggregation utility that exposes logs from all pods in the current namespace via HTTP. It's designed to run as an in-cluster deployment with minimal configuration.

**Versioning**: The project uses Calendar Versioning (CalVer) with format `YYYY.M.PATCH`. See [VERSIONING.md](VERSIONING.md) for details on the automated release process.

## Build & Development Commands

```bash
# Build the binary locally (development version)
go build -o k8s-simple-logs

# Build with version
go build -ldflags "-X main.Version=2025.1.0" -o k8s-simple-logs

# Build Docker image
docker build -t k8s-simple-logs .

# Run locally (requires KUBECONFIG or in-cluster configuration)
go run main.go

# Enable debug mode
DEBUG=1 go run main.go

# Run tests (must be in Kubernetes cluster or tests will be skipped)
go test -v

# Run tests with coverage
go test -v -cover

# Run a specific test
go test -v -run TestHealthcheck
```

## Deployment

The application can be deployed using Helm (recommended), Kustomize, or plain YAML manifests.

### Helm Deployment (Recommended)

```bash
# Install with default values
helm install k8s-simple-logs ./helm/k8s-simple-logs -n my-namespace --create-namespace

# Install with authentication enabled
helm install k8s-simple-logs ./helm/k8s-simple-logs \
  --set logkey=mysecretkey \
  --set service.type=NodePort \
  -n my-namespace --create-namespace

# Upgrade an existing release
helm upgrade k8s-simple-logs ./helm/k8s-simple-logs

# Uninstall
helm uninstall k8s-simple-logs
```

The Helm chart is located in [helm/k8s-simple-logs/](helm/k8s-simple-logs/) with templates for all Kubernetes resources. Configuration options are documented in [values.yaml](helm/k8s-simple-logs/values.yaml).

### Kustomize Deployment

```bash
# Deploy using remote kustomization
kubectl apply -k https://github.com/fredsmith/k8s-simple-logs/kustomize/overlays/production

# Or deploy base to a specific namespace
kubectl apply -k https://github.com/fredsmith/k8s-simple-logs/kustomize/base -n my-namespace

# Local customization
cd kustomize/overlays/production
# Edit kustomization.yaml to set namespace and other options
kustomize build . | kubectl apply -f -
```

Kustomize files are in [kustomize/base/](kustomize/base/) with an example overlay in [kustomize/overlays/production/](kustomize/overlays/production/).

### Plain YAML

```bash
# Deploy to default namespace
kubectl apply -f k8s-deployment.yaml

# Deploy to a specific namespace
kubectl apply -f k8s-deployment.yaml -n my-namespace

# Or deploy directly from GitHub
kubectl apply -f https://raw.githubusercontent.com/fredsmith/k8s-simple-logs/main/k8s-deployment.yaml -n my-namespace
```

Note: The YAML file no longer hardcodes namespaces, so resources will be created in whatever namespace is specified with `-n` or in your current kubectl context.

## Architecture

### Application Structure

The application consists of two main Go files:

- **[main.go](main.go)** - Core application logic:
  - `setupRouter()`: Initializes Gin HTTP server, Kubernetes client, and defines all endpoints
  - `main()`: Entry point that calls setupRouter and starts the server
  - API endpoints for container listing, log retrieval, and WebSocket streaming
  - Authentication middleware
  - Legacy `/logs` endpoint for backward compatibility

- **[ui.go](ui.go)** - Frontend web interface:
  - `getHTMLUI()`: Returns embedded HTML/CSS/JavaScript for the modern web UI
  - Tailwind CSS-based responsive design
  - WebSocket client for real-time log streaming
  - Container search and selection functionality

### Kubernetes Client Configuration
The application supports both in-cluster and local execution ([main.go:31-67](main.go#L31-L67)):
1. **In-cluster mode** (production): Uses `rest.InClusterConfig()` and reads namespace from `/var/run/secrets/kubernetes.io/serviceaccount/namespace`
2. **Kubeconfig mode** (development/testing): Falls back to `~/.kube/config` if not in-cluster
3. Requires RBAC permissions (Role + RoleBinding) to list pods and read pod logs in both modes

### API Endpoints

The application exposes several endpoints:

1. **`GET /`** - Modern web UI (Tailwind CSS + WebSocket)
   - Displays interactive dashboard for container selection
   - Lists all pods/containers in sidebar
   - Streams logs in real-time via WebSocket

2. **`GET /api/containers`** - List all pods and containers
   - Returns JSON array of `{podName, containerName, namespace, id}`
   - Used by web UI to populate sidebar

3. **`GET /api/logs/:pod/:container`** - Fetch logs for specific container
   - Query param `lines=N` controls tail lines (default: 100)
   - Returns JSON: `{pod, container, logs}`

4. **`WS /ws/logs/:pod/:container`** - WebSocket for real-time log streaming
   - Streams logs line-by-line as JSON: `{timestamp, log}`
   - Uses K8s `Follow: true` option for continuous streaming
   - Automatically closes on client disconnect

5. **`GET /logs`** - Legacy plaintext endpoint (backward compatible)
   - Returns all logs from all containers as concatenated text
   - Query param `lines=N` (default: 20)

6. **`GET /version`** - Application version and namespace info

7. **`GET /healthcheck`** - Health probe endpoint

### Authentication

Authentication is optional via the `LOGKEY` environment variable:
- If set, all API endpoints require authentication
- Supports query parameter: `?key=VALUE`
- Supports HTTP header: `X-API-Key: VALUE`
- WebSocket authentication via query parameter only

## Configuration

Environment variables:
- `LOGKEY`: If set, requires `?key=LOGKEY` in requests to `/logs`
- `DEBUG`: If set to any value, enables Gin debug mode

Query parameters on `/logs`:
- `key`: Authentication key (required if LOGKEY is set)
- `lines`: Number of tail lines to fetch per container (default: 20)

## Testing

Unit tests are in [main_test.go](main_test.go) and use the testify/assert library. Tests cover:
- Health check endpoint
- Authentication with LOGKEY environment variable
- Query parameter parsing (key, lines)
- Invalid parameter handling

Tests work with any valid Kubernetes configuration (kubeconfig or in-cluster) and will automatically use the appropriate authentication method.

## CI/CD

The project uses GitHub Actions for continuous integration and deployment automation.

### Test Workflow

[.github/workflows/test.yml](.github/workflows/test.yml) runs on every push and pull request to `main`:
1. Sets up a kind (Kubernetes in Docker) cluster
2. Builds the Docker image and loads it into kind
3. Deploys the application with proper RBAC
4. Runs all unit tests with race detection and coverage
5. Generates test and coverage reports in the workflow summary

The CI environment simulates the in-cluster deployment to ensure tests run successfully.

### Helm Chart Linting

[.github/workflows/lint-helm-chart.yml](.github/workflows/lint-helm-chart.yml) validates the Helm chart:
1. Runs `helm lint` to check chart structure and syntax
2. Templates the chart with default and custom values
3. Runs `kube-linter` on the templated manifests to detect misconfigurations
4. Surfaces any regressions in the workflow summary
5. Uploads templated manifests as artifacts for inspection

Configuration for kube-linter is in [.kube-linter.yaml](.kube-linter.yaml). The linting workflow helps catch:
- Missing health probes
- Security issues (privileged containers, host mounts)
- Resource misconfigurations
- Service selector mismatches

To run linting locally:
```bash
# Lint the Helm chart
helm lint helm/k8s-simple-logs

# Template and lint with kube-linter
helm template test helm/k8s-simple-logs | kube-linter lint - --config .kube-linter.yaml
```

### Release Workflow

[.github/workflows/release.yml](.github/workflows/release.yml) handles automated releases:
1. **Triggers** on pushes to `main` (excluding docs) or manual workflow dispatch
2. **Generates CalVer version** using `scripts/generate-version.sh` (format: `YYYY.M.PATCH`)
3. **Updates Helm chart** version and appVersion in Chart.yaml and values.yaml
4. **Builds Go binary** with version embedded via `-ldflags`
5. **Runs tests** to ensure code quality
6. **Builds and pushes Docker image** with multi-arch support (amd64, arm64)
   - Tagged as both `YYYY.M.PATCH` and `latest`
7. **Packages Helm chart** with the new version
8. **Creates Git tag** (`vYYYY.M.PATCH`) and commits version changes
9. **Creates GitHub Release** with auto-generated release notes
10. **Updates Helm repository** index on GitHub Pages

The workflow can also be triggered manually with a custom version via workflow_dispatch.

See [VERSIONING.md](VERSIONING.md) for complete details on the versioning strategy.

## Dependencies

- **Gin**: HTTP framework for routing and request handling (v1.11.0)
- **k8s.io/client-go**: Kubernetes client library (v0.34.1)
- **testify/assert**: Testing assertions library (v1.11.1)
- Go 1.24.0

## Kubernetes Resources

All deployment methods (Helm, Kustomize, and plain YAML) deploy the same five Kubernetes resources:

1. **ServiceAccount**: `k8s-simple-logs` - dedicated service account for the pod
2. **Deployment**: Single replica running the application container
3. **Service**: ClusterIP on port 8080 (NodePort configurable via Helm)
4. **Role**: Grants `get` and `list` permissions for `pods` and `pods/log` resources
5. **RoleBinding**: Connects the service account to the role

### File Locations:
- **Helm**: [helm/k8s-simple-logs/templates/](helm/k8s-simple-logs/templates/)
- **Kustomize**: [kustomize/base/](kustomize/base/)
- **Plain YAML**: [k8s-deployment.yaml](k8s-deployment.yaml)

All YAML files are namespace-agnostic and will deploy to whatever namespace is specified via `-n` flag or the current kubectl context.
