# CLI Commands

## Overview

| Command | Description |
|---------|-------------|
| [`init`](#init) | Generate a `template.yaml` file via interactive wizard |
| [`create`](#create) | Create the cluster and install configured plugins |
| [`upgrade`](#upgrade) | Update plugins on an existing cluster |
| [`status`](#status) | Show cluster and plugin status |
| [`destroy`](#destroy) | Destroy the cluster |
| [`get`](#get) | Subcommands for retrieving cluster information |
| [`check`](#check) | Verify that all prerequisites are installed |
| [`switch`](#switch) | Switch kubectl context between clusters |

---

## `init`

Generates a `template.yaml` template file via an interactive wizard that guides you through plugin and option selection.

```bash
deploy-cluster init [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `template.yaml` | Output file path |

### Example

```bash
# Generate template.yaml in the current directory
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

Creates a new Kubernetes cluster and installs all plugins enabled in the template.

```bash
deploy-cluster create [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, --template` | `template.yaml` | Template file |
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
deploy-cluster create --template template.yaml
deploy-cluster create --template template.yaml --env production.env
deploy-cluster create --template template.yaml --timeout 10m
```

---

## `upgrade`

Updates an existing cluster by applying only the differences from the current template. The cluster is not recreated.

```bash
deploy-cluster upgrade [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, --template` | `template.yaml` | Template file |
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
deploy-cluster upgrade --template template.yaml --dry-run

# Apply
deploy-cluster upgrade --template template.yaml
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
| `-t, --template` | `template.yaml` | Template file |

### Example

```bash
deploy-cluster status --template template.yaml
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
| `-t, --template` | `template.yaml` | Template file |
| `-n, --name` | - | Cluster name (overrides template) |

### Example

```bash
deploy-cluster destroy --template template.yaml
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

---

## `check`

Verifies that all required tools are installed and shows their versions.

```bash
deploy-cluster check
```

### Example

```bash
$ deploy-cluster check

Prerequisites
─────────────
✓ docker     27.5.1
✓ kind       v0.25.0
✓ kubectl    v1.31.4
✓ helm       v3.16.3

All prerequisites satisfied!
```

If a tool is missing:

```
✗ kind       not found → https://kind.sigs.k8s.io/
```

The command exits with an error if any prerequisite is missing.

---

## `switch`

Switches the active kubectl context to a kind cluster. Without arguments, lists all clusters and highlights the current context.

```bash
deploy-cluster switch [cluster-name]
```

### Example

```bash
# List clusters with current context marker
$ deploy-cluster switch

KIND CLUSTERS
─────────────
  my-cluster
● dev-cluster (current)
  staging

# Switch to a specific cluster
$ deploy-cluster switch my-cluster
Switched to context 'kind-my-cluster'
```
