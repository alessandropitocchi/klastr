# Getting Started

## Requirements

| Tool | Version | Notes |
|------|---------|-------|
| [Go](https://go.dev/) | 1.21+ | To build the binary |
| [Docker](https://www.docker.com/) | - | Runtime for kind |
| [kind](https://kind.sigs.k8s.io/) | - | Local Kubernetes provider |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | - | Cluster interaction |
| [Helm](https://helm.sh/) | 3.x | For monitoring, dashboard, and customApps |

## Installation

```bash
git clone https://github.com/alepito/deploy-cluster.git
cd deploy-cluster
go build -o deploy-cluster ./cmd/deploycluster
```

## Quick Start

### 1. Generate the configuration

```bash
./deploy-cluster init
```

The interactive wizard guides you through:
- Cluster name and topology (control planes, workers, K8s version)
- Plugins to enable (storage, ingress, cert-manager, monitoring, dashboard, ArgoCD)
- Hostnames for ingresses (if ingress is enabled)
- ArgoCD configuration (namespace, version)

The result is a ready-to-use `cluster.yaml` file.

### 2. Create the cluster

```bash
./deploy-cluster create --config cluster.yaml
```

The tool creates the kind cluster and automatically installs all configured plugins in the correct order.

### 3. Check status

```bash
./deploy-cluster status --config cluster.yaml
```

Example output:

```
Cluster: my-cluster
Provider: kind
Status: running

Storage: installed (local-path-provisioner)

Ingress: installed (nginx)

Cert-manager: installed

Monitoring: installed (prometheus)

Dashboard: installed (headlamp)

Custom Apps (1 configured):
  - redis: installed

ArgoCD: installed (namespace: argocd)
  Repos (1):
    - app-repo
  Apps (1):
    - nginx
```

### 4. Update the configuration

Edit `cluster.yaml` and apply changes without recreating the cluster:

```bash
# Preview changes
./deploy-cluster upgrade --config cluster.yaml --dry-run

# Apply
./deploy-cluster upgrade --config cluster.yaml
```

### 5. Destroy the cluster

```bash
./deploy-cluster destroy --config cluster.yaml
```

## Next Steps

- [CLI Commands](commands.md) — complete command reference
- [Configuration](configuration.md) — `cluster.yaml` file structure
- [Provider](providers/kind.md) — kind provider details
- [Plugins](plugins/) — documentation for each plugin
