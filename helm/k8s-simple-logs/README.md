# k8s-simple-logs Helm Chart

This Helm chart deploys k8s-simple-logs, a simple Kubernetes log aggregation utility that exposes logs from all pods in the namespace via HTTP.

## Installation

```bash
# Install with default values
helm install k8s-simple-logs . -n my-namespace --create-namespace

# Install with custom values
helm install k8s-simple-logs . \
  --set logkey=mysecretkey \
  --set service.type=NodePort \
  -n my-namespace --create-namespace

# Install with values file
helm install k8s-simple-logs . -f custom-values.yaml
```

## Configuration

The following table lists the configurable parameters and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `docker.io/derf/k8s-simple-logs` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Container image tag | `latest` |
| `logkey` | Authentication key for /logs endpoint | `""` (disabled) |
| `debug` | Enable debug mode | `false` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` (uses release name) |
| `rbac.create` | Create RBAC resources | `true` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `100Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `100Mi` |

See [values.yaml](values.yaml) for all available options.

## Usage

After installation, access logs via:

```bash
# Port forward
kubectl port-forward svc/k8s-simple-logs 8080:8080 -n my-namespace

# Access logs
curl http://localhost:8080/logs?lines=50

# With authentication (if logkey is set)
curl http://localhost:8080/logs?lines=50&key=mysecretkey
```

## Uninstallation

```bash
helm uninstall k8s-simple-logs -n my-namespace
```
