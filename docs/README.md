# klastr Documentation

Welcome to the klastr documentation. This guide will help you get started with deploying Kubernetes clusters using klastr.

## Quick Navigation

### Getting Started
- [Getting Started Guide](getting-started.md) - Installation and first cluster deployment
- [Configuration](configuration.md) - Complete configuration reference
- [CLI Commands](commands.md) - All available commands

### Installation
```bash
# Using install script (recommended)
curl -fsSL https://raw.githubusercontent.com/alessandropitocchi/deploy-cluster/main/install.sh | bash

# Or download from GitHub Releases
# https://github.com/alessandropitocchi/deploy-cluster/releases
```

### Providers
- [kind](providers/kind.md) - Local Kubernetes clusters in Docker
- [k3d](providers/k3d.md) - Lightweight k3s clusters
- [Existing](providers/existing.md) - Use existing clusters (EKS, GKE, AKS, etc.)

### Plugins
| Plugin | Description |
|--------|-------------|
| [Storage](plugins/storage.md) | StorageClass provisioner |
| [Ingress](plugins/ingress.md) | NGINX/Traefik ingress controller |
| [Cert-Manager](plugins/cert-manager.md) | TLS certificate management |
| [External DNS](plugins/external-dns.md) | Automatic DNS management |
| [Istio](plugins/istio.md) | Service mesh with mTLS |
| [Monitoring](plugins/monitoring.md) | Prometheus + Grafana |
| [Dashboard](plugins/dashboard.md) | Headlamp Kubernetes UI |
| [ArgoCD](plugins/argocd.md) | GitOps continuous delivery |
| [Custom Apps](plugins/custom-apps.md) | Deploy custom Helm charts |

### Advanced Topics
- [Multi-Environment](configuration.md#multi-environment-configuration) - Manage dev/staging/prod
- [Directory Structure](configuration.md#directory-structure-recommended) - Split config across files
- [Templating](templating.md) - Dynamic configuration with Go templates
- [Drift Detection](drift.md) - Detect configuration drift
- [Snapshots](commands.md#snapshot) - Backup and restore clusters

### Examples
- [Basic Cluster](../examples/simple/) - Simple single-file configuration
- [Directory Config](../examples/directory-config/) - Multi-file setup
- [Multi-Environment](../examples/multi-env/) - Dev/staging/prod setup

## Quick Start

```bash
# 1. Check prerequisites
klastr check

# 2. Generate configuration
klastr init

# 3. Validate
klastr lint

# 4. Deploy
klastr run

# 5. Check status
klastr status
```

## Help & Support

- [GitHub Issues](https://github.com/alessandropitocchi/deploy-cluster/issues) - Report bugs or request features
- [GitHub Discussions](https://github.com/alessandropitocchi/deploy-cluster/discussions) - Ask questions

---

**Note**: Replace `klastr` with `./klastr` if running from the local directory without installing.
