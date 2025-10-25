# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

k8s-simple-logs is a simple Kubernetes log aggregation utility that exposes logs from all pods in the current namespace via HTTP. It's designed to run as an in-cluster deployment with minimal configuration.

## Build & Development Commands

```bash
# Build the binary locally
go build -o k8s-simple-logs

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

### Single-File Application
This is a monolithic Go application contained entirely in [main.go](main.go). All functionality lives in one file with two main functions:

- `setupRouter()`: Initializes the Gin HTTP server, Kubernetes client, and defines endpoints
- `main()`: Entry point that calls setupRouter and starts the server

### Kubernetes Client Configuration
The application supports both in-cluster and local execution ([main.go:31-67](main.go#L31-L67)):
1. **In-cluster mode** (production): Uses `rest.InClusterConfig()` and reads namespace from `/var/run/secrets/kubernetes.io/serviceaccount/namespace`
2. **Kubeconfig mode** (development/testing): Falls back to `~/.kube/config` if not in-cluster
3. Requires RBAC permissions (Role + RoleBinding) to list pods and read pod logs in both modes

### Request Flow
When `/logs` is accessed:
1. Optional LOGKEY authentication check (via `?key=` query parameter)
2. Parse `?lines=` parameter (default: 20 lines per container)
3. List all pods in the namespace using K8s client
4. For each pod, iterate through all containers
5. Stream logs from each container via `GetLogs()` API
6. Concatenate all logs with headers showing pod/container metadata
7. Return as plain text response

### Multi-Container Pod Handling
The nested loop structure ([main.go:78-105](main.go#L78-L105)) is critical:
- Outer loop iterates pods
- Inner loop iterates containers within each pod
- Each container's logs are fetched independently and labeled with ID numbers

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

### Helm Chart Publishing

[.github/workflows/release-helm-chart.yml](.github/workflows/release-helm-chart.yml) automatically publishes charts:
1. Triggers when changes are pushed to `helm/` directory
2. Packages the Helm chart
3. Creates GitHub Releases
4. Updates the Helm repository index on GitHub Pages

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
