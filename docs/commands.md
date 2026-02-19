# CLI Commands

## Overview

| Command | Description |
|---------|-------------|
| [`init`](#init) | Generate a `cluster.yaml` file via interactive wizard |
| [`create`](#create) | Create the cluster and install configured plugins |
| [`upgrade`](#upgrade) | Update plugins on an existing cluster |
| [`status`](#status) | Show cluster and plugin status |
| [`destroy`](#destroy) | Destroy the cluster |
| [`get`](#get) | Subcommands for retrieving cluster information |

---

## `init`

Generates a `cluster.yaml` configuration file via an interactive wizard that guides you through plugin and option selection.

```bash
deploy-cluster init [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `cluster.yaml` | Output file path |

### Example

```bash
# Generate cluster.yaml in the current directory
deploy-cluster init

# Generate with custom name
deploy-cluster init -o my-cluster.yaml
```

The wizard asks in sequence:
1. **Cluster** — name, Kubernetes version, number of control planes and workers
2. **Plugins** — multi-select plugins to enable
3. **Ingress** — hostname for each service with a UI (if ingress enabled)
4. **ArgoCD** — namespace and version (if ArgoCD enabled)

---

## `create`

Creates a new Kubernetes cluster and installs all plugins enabled in the configuration.

```bash
deploy-cluster create [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | Configuration file |
| `-e, --env` | `.env` | File with environment variables for secrets |
| `--timeout` | `5m` | Timeout for plugin operations (kubectl/helm) |

### Installation Order

Plugins are installed in this order:

1. **Storage** — to make PVCs available
2. **Ingress** — to expose services via hostname
3. **Cert-Manager** — for TLS certificates
4. **Monitoring** — Prometheus + Grafana
5. **Dashboard** — Headlamp
6. **Custom Apps** — custom Helm charts
7. **ArgoCD** — GitOps (last, as it may depend on others)

### Example

```bash
deploy-cluster create --config cluster.yaml
deploy-cluster create --config cluster.yaml --env production.env
deploy-cluster create --config cluster.yaml --timeout 10m
```

---

## `upgrade`

Updates an existing cluster by applying only the differences from the current configuration. The cluster is not recreated.

```bash
deploy-cluster upgrade [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | Configuration file |
| `-e, --env` | `.env` | File with environment variables for secrets |
| `--dry-run` | `false` | Show changes without applying them |
| `--timeout` | `5m` | Timeout for plugin operations (kubectl/helm) |

### Per-plugin Behavior

| Plugin | Behavior |
|--------|----------|
| Storage | Re-apply manifest (idempotent) |
| Ingress | Re-apply manifest (idempotent) |
| Cert-Manager | Re-apply manifest (updates version if changed) |
| Monitoring | `helm upgrade` (idempotent) |
| Dashboard | `helm upgrade` (idempotent) |
| Custom Apps | `helm upgrade --install` for each app |
| ArgoCD | Re-apply manifest + diff repos/apps (adds new, removes those deleted from config) |

### Example

```bash
# Preview
deploy-cluster upgrade --config cluster.yaml --dry-run

# Apply
deploy-cluster upgrade --config cluster.yaml
```

---

## `status`

Shows the current cluster status: existence, installed plugins, ArgoCD repositories and applications.

```bash
deploy-cluster status [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | Configuration file |

### Example

```bash
deploy-cluster status --config cluster.yaml
```

---

## `destroy`

Destroys the cluster. Deletes all associated Docker containers.

```bash
deploy-cluster destroy [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | Configuration file |
| `-n, --name` | - | Cluster name (overrides config) |

### Example

```bash
deploy-cluster destroy --config cluster.yaml
deploy-cluster destroy --name my-cluster
```

---

## `get`

Subcommands for retrieving cluster information.

### `get clusters`

List all existing kind clusters.

```bash
deploy-cluster get clusters
```

### `get nodes <name>`

List nodes of a specific cluster.

```bash
deploy-cluster get nodes my-cluster
```

### `get kubeconfig <name>`

Print the kubeconfig of a cluster.

```bash
deploy-cluster get kubeconfig my-cluster

# Save to file
deploy-cluster get kubeconfig my-cluster > kubeconfig.yaml
```
