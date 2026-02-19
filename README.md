# deploy-cluster

CLI tool for deploying local Kubernetes clusters with plugin support.

Create clusters with configurable topology and automatically install components like storage, ingress, cert-manager, monitoring, dashboard, ArgoCD and custom Helm apps — all defined in a single configuration file.

## Requirements

- Go 1.21+ | Docker | [kind](https://kind.sigs.k8s.io/) | kubectl | [Helm](https://helm.sh/) 3.x

## Installation

```bash
go build -o deploy-cluster ./cmd/deploycluster
```

## Quick Start

```bash
# Check prerequisites
./deploy-cluster check

# Interactive wizard to generate the configuration
./deploy-cluster init

# Create the cluster with all configured plugins
./deploy-cluster create --config cluster.yaml

# Check status
./deploy-cluster status --config cluster.yaml

# Update plugins without recreating the cluster
./deploy-cluster upgrade --config cluster.yaml

# Switch kubectl context between clusters
./deploy-cluster switch my-cluster

# Destroy the cluster
./deploy-cluster destroy --config cluster.yaml
```

## Configuration Example

```yaml
name: my-cluster
provider:
  type: kind
cluster:
  controlPlanes: 1
  workers: 2
  version: v1.31.0
plugins:
  storage:
    enabled: true
    type: local-path
  ingress:
    enabled: true
    type: nginx
  certManager:
    enabled: true
  monitoring:
    enabled: true
    type: prometheus
    ingress:
      enabled: true
      host: grafana.localhost
  dashboard:
    enabled: true
    type: headlamp
    ingress:
      enabled: true
      host: headlamp.localhost
  customApps:
    - name: redis
      chart: oci://registry-1.docker.io/bitnamicharts/redis
      version: "21.1.5"
      namespace: redis
      values:
        architecture: standalone
  argocd:
    enabled: true
    ingress:
      enabled: true
      host: argocd.localhost
```

## Available Plugins

| Plugin | Description | Installation |
|--------|-------------|--------------|
| [Storage](docs/plugins/storage.md) | StorageClass provisioner (local-path) | kubectl apply |
| [Ingress](docs/plugins/ingress.md) | NGINX controller for kind | kubectl apply |
| [Cert-Manager](docs/plugins/cert-manager.md) | TLS certificate management | kubectl apply |
| [Monitoring](docs/plugins/monitoring.md) | Prometheus + Grafana | Helm |
| [Dashboard](docs/plugins/dashboard.md) | Headlamp | Helm |
| [Custom Apps](docs/plugins/custom-apps.md) | Custom Helm charts | Helm |
| [ArgoCD](docs/plugins/argocd.md) | GitOps | kubectl apply |

## UI Access

With ingress enabled:

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | `http://grafana.localhost` | admin / prom-operator |
| Headlamp | `http://headlamp.localhost` | `kubectl create token headlamp -n headlamp` |
| ArgoCD | `http://argocd.localhost` | admin / `kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" \| base64 -d` |

## Documentation

| Section | Description |
|---------|-------------|
| [Getting Started](docs/getting-started.md) | Installation and first use |
| [CLI Commands](docs/commands.md) | Complete command reference |
| [Configuration](docs/configuration.md) | `cluster.yaml` file structure |
| [Provider: kind](docs/providers/kind.md) | kind provider details |
| [Plugins](docs/plugins/) | Documentation for each plugin |
