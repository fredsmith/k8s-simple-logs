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

```bash
# Deploy to Kubernetes
kubectl apply -f k8s-deployment.yaml

# Important: Update namespace in k8s-deployment.yaml
# Change lines 46, 54, and 61 from "default" to your target namespace
```

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

GitHub Actions workflow in [.github/workflows/test.yml](.github/workflows/test.yml) runs on every push and pull request to `main`. The workflow:
1. Sets up a kind (Kubernetes in Docker) cluster
2. Builds the Docker image and loads it into kind
3. Deploys the application with proper RBAC
4. Runs all unit tests with race detection and coverage
5. Uploads coverage reports to Codecov (optional)

The CI environment simulates the in-cluster deployment to ensure tests run successfully.

## Dependencies

- **Gin**: HTTP framework for routing and request handling
- **k8s.io/client-go**: Kubernetes client library (v0.24.0)
- **testify/assert**: Testing assertions library
- Go 1.18

## Kubernetes Resources

The [k8s-deployment.yaml](k8s-deployment.yaml) defines four resources that must be deployed together:
1. Deployment (single replica)
2. Service (NodePort on 8080)
3. Role (viewlogs - grants pod and pod/log read access)
4. RoleBinding (connects service account to role)
