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

### 1. Generate the template

```bash
./deploy-cluster init
```

The interactive wizard guides you through:
- Cluster name and topology (control planes, workers, K8s version)
- Plugins to enable (storage, ingress, cert-manager, monitoring, dashboard, ArgoCD)
- Hostnames for ingresses (if ingress is enabled)
- ArgoCD configuration (namespace, version)

The result is a ready-to-use `template.yaml` file.

### 2. Validate the template

Before creating the cluster, validate the template for errors and best practices:

```bash
./deploy-cluster lint
```

This checks for:
- Valid cluster name and Kubernetes version
- Correct topology configuration
- Ingress host uniqueness
- Resource dependencies
- Best practices recommendations

### 3. Create the cluster

```bash
./deploy-cluster create --template template.yaml
```

The tool creates the kind cluster and automatically installs all configured plugins in the correct order.

### 4. Check status

```bash
./deploy-cluster status --template template.yaml
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

### 5. Update the template

Edit `template.yaml` and apply changes without recreating the cluster:

```bash
# Preview changes
./deploy-cluster upgrade --template template.yaml --dry-run

# Apply
./deploy-cluster upgrade --template template.yaml
```

### 6. Destroy the cluster

```bash
./deploy-cluster destroy --template template.yaml
```

## Next Steps

- [CLI Commands](commands.md) — complete command reference
- [Configuration](configuration.md) — `template.yaml` file structure
- [Provider](providers/kind.md) — kind provider details
- [Plugins](plugins/) — documentation for each plugin
