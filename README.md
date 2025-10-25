# k8s simple logs utility

A simple log viewer for kubernetes hosted applications.

This tool exposes all the logs from every pod in the current namespace over http.

## Installation

Choose one of the following installation methods:

### Option 1: Helm (Recommended)

```bash
# Add the Helm repository
helm repo add k8s-simple-logs https://fredsmith.github.io/k8s-simple-logs

# Update your local Helm chart repository cache
helm repo update

# Install the chart
helm install my-release k8s-simple-logs/k8s-simple-logs

# Install with custom values
helm install my-release k8s-simple-logs/k8s-simple-logs \
  --set logkey=mysecret \
  --set service.type=NodePort \
  -n my-namespace

```

### Option 2: Kustomize

```bash
# Deploy to the default namespace
kubectl apply -k https://github.com/fredsmith/k8s-simple-logs/kustomize/overlays/production

# Or deploy to a specific namespace
kubectl apply -k https://github.com/fredsmith/k8s-simple-logs/kustomize/base -n my-namespace
```

### Option 3: Plain YAML

```bash
# Deploy to default namespace
kubectl apply -f https://raw.githubusercontent.com/fredsmith/k8s-simple-logs/main/k8s-deployment.yaml

# Deploy to a specific namespace
kubectl apply -f https://raw.githubusercontent.com/fredsmith/k8s-simple-logs/main/k8s-deployment.yaml -n my-namespace
```

## Configuration

### Helm Values

Key configuration options when installing with Helm:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `docker.io/derf/k8s-simple-logs` |
| `image.tag` | Container image tag | `latest` |
| `logkey` | Authentication key for /logs endpoint | `""` (disabled) |
| `debug` | Enable debug mode | `false` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `100Mi` |

See [values.yaml](helm/k8s-simple-logs/values.yaml) for all available options.

### Environment Variables

- `LOGKEY`: If set, requires `?key=LOGKEY` in requests to `/logs`
- `DEBUG`: If set to any value, enables Gin debug mode

## Accessing

### Via Ingress

Add the service to your ingress:

```yaml
- path: /logs
  pathType: Prefix
  backend:
    service:
      name: k8s-simple-logs  # or release name if using Helm
      port:
        number: 8080
```

Make sure it's before any less specific rules.

### Via Port Forward

```bash

kubectl port-forward svc/k8s-simple-logs 8080:8080

# Then access at http://localhost:8080/logs
```

### Web UI

Access the modern web interface at `http://localhost:8080/`

Features:
- **Container sidebar** - Browse and search all pods/containers in the namespace
- **Real-time log streaming** - WebSocket-based live log updates with automatic reconnection
- **Auto-scroll** - Toggle automatic scrolling to latest logs
- **Search** - Filter containers by name
- **Dark theme** - Terminal-style log display
- **Resilient connections** - Automatic reconnection on disconnect (up to 5 attempts with exponential backoff)

### API Endpoints

- **`GET /`** - Modern web UI for log viewing
  - Interactive dashboard with container selection
  - Real-time WebSocket log streaming

- **`GET /api/containers`** - List all pods and containers
  - Returns JSON with container list
  - Authentication: query param `?key=<value>` or header `X-API-Key`

- **`GET /api/logs/:pod/:container`** - Get logs for specific container
  - Query params: `lines=N` (default: 100), `key=<value>`
  - Returns JSON with log content

- **`WS /ws/logs/:pod/:container`** - WebSocket for real-time log streaming
  - Streams logs as JSON messages: `{"timestamp":"...", "log":"..."}`
  - Authentication: query param `?key=<value>`

- **`GET /logs`** - Legacy endpoint (backward compatible)
  - Returns all logs from all containers as plain text
  - Query params: `lines=N` (default: 20), `key=<value>`

- **`GET /version`** - Application version and namespace
  - Returns JSON: `{"version":"2025.1.0","namespace":"default"}`

- **`GET /healthcheck`** - Health check
  - Returns: `still alive`

## Development

### Running Tests

```bash
# Run all tests (requires Kubernetes cluster or kubeconfig)
go test -v

# Run tests with coverage
go test -v -cover

# Run a specific test
go test -v -run TestHealthcheck
```

### CI/CD

The project uses GitHub Actions for automated testing and validation:

- **[Test Workflow](.github/workflows/test.yml)** - Runs Go tests in a kind cluster
- **[Lint Workflow](.github/workflows/lint-helm-chart.yml)** - Validates Helm charts with kube-linter
- **[Release Workflow](.github/workflows/release.yml)** - Releases Versioned docker image and publishes Helm Chart

See [CLAUDE.md](CLAUDE.md) for detailed CI/CD documentation.

