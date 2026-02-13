# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**deploy-cluster** is a CLI tool for deploying Kubernetes clusters on various providers, starting with kind (Kubernetes in Docker). The tool allows users to define cluster topology (number of workers, control plane nodes) and install plugins for storage, CD, and other cluster components.

## Planned Architecture

### Core Concepts

- **Providers**: Abstraction layer for different cluster providers (kind, k3d, etc.)
- **Plugins**: Modular components that can be installed on clusters (storage, ArgoCD, etc.)
- **Cluster Config**: YAML-based configuration for defining cluster topology

### Target Features

1. **Cluster Creation**: Deploy clusters with configurable:
   - Number of worker nodes
   - Number of control plane nodes
   - Kubernetes version
   - Network configuration

2. **Plugin System**: Post-deployment installation of components with their configuration:
   - **ArgoCD**: Install and configure with target Git repository, project settings, and Application definitions
   - **Storage**: local-path-provisioner, OpenEBS, etc.
   - **Ingress**: nginx, traefik
   - **Monitoring**: Prometheus, Grafana stack

3. **Plugin Configuration Example** (ArgoCD):
   ```yaml
   plugins:
     argocd:
       enabled: true
       repo: https://github.com/user/gitops-repo.git
       path: environments/dev
       targetRevision: main
       project: default
   ```

4. **Provider Abstraction**: Support multiple local Kubernetes providers:
   - kind (initial target)
   - k3d (future)
   - minikube (future)

## Development

### Tech Stack
- **Language**: Go
- **CLI Framework**: cobra
- **Kubernetes Client**: client-go, kind API
- **Config Format**: YAML (parsed with gopkg.in/yaml.v3)

### Build & Run
```bash
go build -o deploy-cluster ./cmd/deploy-cluster
./deploy-cluster create --config cluster.yaml
```

### Project Structure (planned)
```
cmd/deploy-cluster/     # CLI entrypoint
pkg/
  provider/             # Provider interface and implementations (kind, k3d)
  plugin/               # Plugin interface and implementations (argocd, storage)
  config/               # YAML config parsing and validation
  cluster/              # Cluster lifecycle management
```

### Key Design Decisions
- Primary provider target: kind
- Plugin installation should be idempotent
- Plugins receive cluster kubeconfig for kubectl/client-go operations
