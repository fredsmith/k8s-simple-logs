# k8s simple logs utility

A simple log viewer for kubernetes hosted applications.

This tool exposes all the logs from every pod in the current namespace over http.

## Installation

Choose one of the following installation methods:

### Option 1: Helm (Recommended)

```bash
# Install directly from the repository
helm install k8s-simple-logs ./helm/k8s-simple-logs

# Or with custom values
helm install k8s-simple-logs ./helm/k8s-simple-logs \
  --set logkey=your-secret-key \
  --set service.type=NodePort

# Install to a specific namespace
helm install k8s-simple-logs ./helm/k8s-simple-logs -n my-namespace --create-namespace
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

### Endpoints

- **`GET /logs`** - Retrieve logs from all pods
  - `lines=N`: Number of log lines to fetch per container (default: 20)
  - `key=<value>`: Authentication key (required if LOGKEY is set)
  - Example: `http://localhost:8080/logs?lines=50&key=mysecret`

- **`GET /version`** - Check application version and namespace
  - Returns JSON: `{"version":"2025.1.0","namespace":"default"}`

- **`GET /healthcheck`** - Health check endpoint
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

