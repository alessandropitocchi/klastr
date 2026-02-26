# CLI Commands

## Overview

| Command | Description |
|---------|-------------|
| [`init`](#init) | Generate a `template.yaml` file via interactive wizard |
| [`lint`](#lint) | Validate template for errors and best practices |
| [`run`](#run) | Deploy cluster and install configured plugins |
| [`upgrade`](#upgrade) | Update plugins on an existing cluster |
| [`drift`](#drift) | Detect drift between cluster and template |
| [`uninstall`](#uninstall) | Uninstall plugins from a cluster (keeps cluster) |
| [`status`](#status) | Show cluster and plugin status |
| [`destroy`](#destroy) | Destroy the cluster |
| [`env`](#env) | Manage multi-environment configurations |
| [`get`](#get) | Subcommands for retrieving cluster information |
| [`check`](#check) | Verify that all prerequisites are installed |
| [`switch`](#switch) | Switch kubectl context between clusters |
| [`snapshot`](#snapshot) | Save, restore, list, and delete cluster snapshots |

---

## `init`

Generates a starter configuration via an interactive wizard. Can create either a single `template.yaml` file or a directory structure.

```bash
klastr init [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `template.yaml` | Output file or directory path |
| `--dir` | `-d` | `false` | Generate directory structure instead of single file |
| `--provider` | `-p` | `kind` | Provider type: `kind`, `k3d`, or `existing` |

### Single File (Default)

Generates a `template.yaml` file with all configuration in one place.

```bash
# Generate template.yaml in the current directory
klastr init

# Generate with custom name
klastr init -o my-cluster.yaml

# Generate for existing cluster
klastr init --provider existing -o eks-cluster.yaml
```

### Directory Structure

Generates a directory with configuration split across multiple files. Recommended for complex setups or team environments.

```bash
# Generate directory structure
klastr init --dir --output my-cluster/

# Generate for specific provider
klastr init --dir --provider k3d --output my-k3d-cluster/
```

Generated structure:

```
my-cluster/
├── klastr.yaml          # Main configuration
├── plugins/             # Plugin configurations
│   ├── storage.yaml
│   ├── ingress.yaml
│   ├── cert-manager.yaml
│   ├── external-dns.yaml
│   ├── istio.yaml
│   ├── monitoring.yaml
│   ├── dashboard.yaml
│   └── argocd.yaml
├── apps/                # Custom application configs
├── .env.example         # Environment variables template
└── README.md            # Documentation
```

### Usage with Directory

All commands accept a directory path instead of a file:

```bash
klastr lint --template my-cluster/
klastr run --template my-cluster/
klastr upgrade --template my-cluster/
```

### Template Flag Aliases

All commands that accept a `--template` flag also support `-t` and `-f` as shorthand:

```bash
# These are all equivalent:
klastr run --template production.yaml
klastr run -t production.yaml
klastr run -f production.yaml

# Works with directories too:
klastr lint -f my-cluster/
klastr upgrade -t my-cluster/
```

---

## `lint`

Validates the template file for errors, warnings, and best practices before creating or upgrading a cluster.

```bash
klastr lint [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `--strict` | `false` | Treat warnings as errors |

### Checks Performed

| Category | Checks |
|----------|--------|
| **Cluster Name** | Valid DNS subdomain format, length (3-63 chars) |
| **Kubernetes Version** | Valid format (vX.Y.Z), EOL version detection |
| **Topology** | Control planes (odd numbers recommended), workers |
| **Ingress Hosts** | Uniqueness, valid hostname format |
| **Dependencies** | Ingress plugin required if ingress hosts configured |
| **Best Practices** | Storage for multi-node, monitoring, ArgoCD repos |

### Example

```bash
# Basic lint
$ klastr lint

  ✓ No issues found!

# Lint with warnings
$ klastr lint --template my-cluster.yaml

  ⚠ [WARN] cluster.controlPlanes: using 2 control planes (even number) is not recommended for etcd quorum, use odd numbers (1, 3, 5)
  ℹ [INFO] plugins.storage: consider enabling storage plugin for multi-node clusters (PVC support)

Total: 2 issues (0 errors, 1 warnings, 1 info)

# Strict mode (treats warnings as errors)
$ klastr lint --strict
```

---

## `run`

Deploys a Kubernetes cluster and installs all plugins enabled in the template.

This command handles both:
- **Creating new clusters** (kind, k3d)
- **Deploying to existing clusters** (EKS, GKE, AKS, or existing kind/k3d)

For `existing` provider, it skips cluster creation and only installs plugins.

```bash
klastr run [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `-e, --env` | `.env` | File with environment variables for secrets |
| `--timeout` | `5m` | Timeout for plugin operations (kubectl/helm) |
| `--fail-fast` | `false` | Stop at first plugin failure |

### Installation Order

Plugins are installed in this order:

1. **Storage** — to make PVCs available
2. **Ingress** — to expose services via hostname
3. **Cert-Manager** — for TLS certificates
4. **External DNS** — for automatic DNS management
5. **Istio** — service mesh (installs before apps that may use it)
6. **Monitoring** — Prometheus + Grafana
7. **Dashboard** — Headlamp
8. **Custom Apps** — custom Helm charts
9. **ArgoCD** — GitOps (last, as it may depend on others)

### Examples

```bash
# Deploy from template (creates cluster if needed)
klastr run --template template.yaml

# Deploy with environment variables
klastr run --template template.yaml --env production.env

# Deploy with extended timeout
klastr run --template template.yaml --timeout 10m

# Deploy to existing cluster (provider: existing)
klastr run --template template-existing.yaml
```

---

## `upgrade`

Updates an existing cluster by applying only the differences from the current template. The cluster is not recreated.

```bash
klastr upgrade [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `-e, --env` | `.env` | File with environment variables for secrets |
| `--dry-run` | `false` | Show changes without applying them |
| `--timeout` | `5m` | Timeout for plugin operations (kubectl/helm) |

### Per-plugin Behavior

| Plugin | Behavior |
|--------|----------|
| Storage | Re-apply manifest (idempotent) |
| Ingress | Re-apply manifest (idempotent) |
| Cert-Manager | Re-apply manifest (updates version if changed) |
| External DNS | `helm upgrade` (idempotent) |
| Istio | `istioctl install` (handles upgrades) |
| Monitoring | `helm upgrade` (idempotent) |
| Dashboard | `helm upgrade` (idempotent) |
| Custom Apps | `helm upgrade --install` for each app |
| ArgoCD | Re-apply manifest + diff repos/apps (adds new, removes those deleted from config) |

### Example

```bash
# Preview
klastr upgrade --template template.yaml --dry-run

# Apply
klastr upgrade --template template.yaml
```

---

## `drift`

Detects drift between the cluster's actual state and the desired state defined in the template. Identifies missing resources, orphan resources, and configuration differences.

```bash
klastr drift [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `-e, --env` | `.env` | Environment file |
| `--exit-error` | `false` | Exit with error code if drift detected |

### Drift Types

| Type | Description | Icon |
|------|-------------|------|
| **Missing** | In template but not in cluster | `-` |
| **Orphan** | In cluster but not in template | `?` |
| **Modified** | Different configuration | `~` |

### Example

```bash
# Basic drift detection
klastr drift

# With specific template
klastr drift --template production.yaml

# Exit with error if drift detected (useful for CI/CD)
klastr drift --exit-error
```

### Sample Output

```
Drift Detection Results:
------------------------------------------------------------

  Missing (in template, not in cluster):
    - [storage] local-path-provisioner: Storage plugin is enabled but not installed

  Orphans (in cluster, not in template):
    ? [custom-apps] redis: Custom app "redis" is installed but not in template

  Modified (drift detected):
    ~ [monitoring] kube-prometheus-stack: Version drift: cluster=72.6.0, template=72.6.2

------------------------------------------------------------
Total: 3 drift items (1 missing, 1 modified, 1 orphans)
```

See [Drift Detection](drift.md) for detailed documentation.

---

## `uninstall`

Uninstalls all enabled plugins from an existing cluster **without destroying the cluster itself**. Useful for resetting plugin state or cleaning up before a fresh plugin install.

```bash
klastr uninstall [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `-e, --env` | `.env` | Environment file for secrets |
| `--fail-fast` | `false` | Stop at first plugin failure |

### Uninstall Order

Plugins are uninstalled in reverse installation order:

1. **ArgoCD**
2. **Custom Apps**
3. **Dashboard**
4. **Monitoring**
5. **Cert-Manager**
6. **Ingress**
7. **Storage**

### Example

```bash
# Uninstall all plugins
klastr uninstall --template template.yaml

# Uninstall with fail-fast
klastr uninstall --template template.yaml --fail-fast
```

---

## `status`

Shows the current cluster status: existence, installed plugins, ArgoCD repositories and applications.

```bash
klastr status [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |

### Example

```bash
klastr status --template template.yaml
```

---

## `destroy`

Destroys the cluster. Deletes all associated Docker containers.

```bash
klastr destroy [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `-n, --name` | - | Cluster name (overrides template) |

### Example

```bash
klastr destroy --template template.yaml
klastr destroy --name my-cluster
```

---

## `env`

Manage multi-environment configurations for different deployment targets (dev, staging, production).

Uses an overlay system similar to Kustomize where a base configuration is patched with environment-specific values.

```bash
klastr env [command]
```

### Directory Structure

```
my-cluster/
├── klastr.yaml              # Base configuration
└── environments/
    ├── dev/
    │   └── overlay.yaml     # Dev-specific patches
    ├── staging/
    │   └── overlay.yaml     # Staging-specific patches
    └── production/
        └── overlay.yaml     # Production-specific patches
```

### `env list`

List all available environments.

```bash
klastr env list
```

**Example output:**
```
ENVIRONMENT  DESCRIPTION
-----------  -----------
dev          Development environment
staging      Staging environment  
production   Production environment
```

### `env create <name>`

Create a new environment overlay.

```bash
klastr env create <name> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--base` | `../../` | Path to base configuration |

**Examples:**
```bash
# Create dev environment
klastr env create dev

# Create production with custom base path
klastr env create production --base ./base-config
```

### `env show <name>`

Show the effective configuration for an environment (base + patches applied).

```bash
klastr env show <name>
```

**Example:**
```bash
klastr env show production
```

### Using Environments

All main commands support the `--environment` (or `-E`) flag:

```bash
# Deploy dev environment
klastr run --environment dev

# Validate staging configuration
klastr lint --environment staging

# Upgrade production
klastr upgrade --environment production

# Check status
klastr status --environment dev
```

### Overlay Configuration

The `overlay.yaml` file defines environment-specific patches:

```yaml
name: production
description: Production environment
base: ../../                    # Relative path to base config
patches:
  - target: name                # Patch cluster name
    value: myapp-prod
  - target: cluster.workers     # Patch worker count
    value: 5
  - target: cluster.controlPlanes
    value: 3
  - target: plugins.monitoring.enabled
    value: true
  - target: plugins.certManager.enabled
    value: true
values:                         # Variables for templating
  DOMAIN: example.com
  LOG_LEVEL: warn
```

### Patch Targets

Available patch targets:

| Target | Type | Description |
|--------|------|-------------|
| `name` | string | Cluster name |
| `cluster.controlPlanes` | int | Number of control plane nodes |
| `cluster.workers` | int | Number of worker nodes |
| `cluster.version` | string | Kubernetes version |
| `provider.type` | string | Provider type |
| `provider.kubeconfig` | string | Kubeconfig path |
| `provider.context` | string | Kubectl context |
| `plugins.storage.enabled` | bool | Enable storage plugin |
| `plugins.ingress.enabled` | bool | Enable ingress plugin |
| `plugins.monitoring.enabled` | bool | Enable monitoring plugin |
| `plugins.dashboard.enabled` | bool | Enable dashboard plugin |
| `plugins.certManager.enabled` | bool | Enable cert-manager plugin |
| `snapshot.enabled` | bool | Enable S3 snapshots |
| `snapshot.bucket` | string | S3 bucket name |
| `snapshot.prefix` | string | S3 key prefix |
| `snapshot.region` | string | AWS region |

---

## `get`

Subcommands for retrieving cluster information.

### `get clusters`

List all existing kind clusters.

```bash
klastr get clusters
```

### `get nodes <name>`

List nodes of a specific cluster.

```bash
klastr get nodes my-cluster
```

### `get kubeconfig <name>`

Print the kubeconfig of a cluster.

```bash
klastr get kubeconfig my-cluster

# Save to file
klastr get kubeconfig my-cluster > kubeconfig.yaml
```

---

## `check`

Verifies that all required tools are installed and shows their versions.

```bash
klastr check
```

### Example

```bash
$ klastr check

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
klastr switch [cluster-name]
```

### Example

```bash
# List clusters with current context marker
$ klastr switch

KIND CLUSTERS
─────────────
  my-cluster
● dev-cluster (current)
  staging

# Switch to a specific cluster
$ klastr switch my-cluster
Switched to context 'kind-my-cluster'
```

---

## `snapshot`

Manage cluster snapshots — export Kubernetes resources to disk or S3 and restore them later.

### `snapshot save <name>`

Export all non-system resources from the cluster to a local snapshot or S3.

```bash
klastr snapshot save <name> [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `-e, --env` | `.env` | Environment file for secrets |
| `--namespace` | *(all non-system)* | Comma-separated list of namespaces to snapshot |
| `--exclude-secrets` | `false` | Exclude Kubernetes Secrets from the snapshot |
| `--s3` | `false` | Upload snapshot to S3 |
| `--s3-bucket` | *(env var)* | S3 bucket name |
| `--s3-prefix` | *(env var)* | S3 key prefix |
| `--s3-region` | *(env var)* | AWS region |
| `--s3-endpoint` | *(env var)* | S3 endpoint URL (for S3-compatible services) |

#### What gets exported

- Dynamic resource discovery via `kubectl api-resources` (captures CRDs too)
- One file per resource, stored at `~/.klastr/snapshots/<name>/` (local) or S3
- Sanitized: removes `resourceVersion`, `uid`, `managedFields`, `status`, etc.

#### What gets excluded

- **System namespaces**: `kube-system`, `kube-public`, `kube-node-lease`, `local-path-storage`
- **Transient resources**: events, endpoints, pods, replicasets, nodes, leases
- **Controller-managed**: resources with `ownerReferences`
- **Auto-created**: `kube-root-ca.crt` ConfigMaps, `default` ServiceAccounts, `kubernetes` Service

#### Examples

```bash
# Save all non-system resources locally
klastr snapshot save before-upgrade --template template.yaml

# Save only specific namespaces
klastr snapshot save my-snap --namespace app,monitoring --template template.yaml

# Save to S3 using flags
klastr snapshot save my-snap --s3 --s3-bucket my-backups --s3-prefix clusters/prod/

# Save to S3 using environment variables
export DEPLOY_CLUSTER_S3_BUCKET=my-backups
export DEPLOY_CLUSTER_S3_PREFIX=clusters/prod/
export AWS_REGION=us-east-1
klastr snapshot save my-snap --s3
```

### `snapshot restore <name>`

Restore resources from a saved snapshot to the cluster.

```bash
klastr snapshot restore <name> [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-t, -f, --template` | `template.yaml` | Template file or directory |
| `-e, --env` | `.env` | Environment file for secrets |
| `--dry-run` | `false` | Preview what would be applied without making changes |
| `--s3` | `false` | Restore snapshot from S3 |
| `--s3-bucket` | *(env var)* | S3 bucket name |
| `--s3-prefix` | *(env var)* | S3 key prefix |
| `--s3-region` | *(env var)* | AWS region |
| `--s3-endpoint` | *(env var)* | S3 endpoint URL |

#### Restore Order

Resources are applied in dependency order:

1. **CRDs** + 5s wait for propagation
2. **Namespaces**
3. **Cluster-scoped**: ClusterRoles → ClusterRoleBindings → PersistentVolumes → other
4. **Namespaced** (per namespace): ServiceAccounts → Secrets/ConfigMaps → PVCs → Services → Deployments/StatefulSets/DaemonSets → Ingresses → Jobs/CronJobs → custom resources

Each `kubectl apply` uses retry with exponential backoff for transient errors.

#### Examples

```bash
# Preview restore from local snapshot
klastr snapshot restore before-upgrade --dry-run --template template.yaml

# Restore from S3
klastr snapshot restore my-snap --s3 --s3-bucket my-backups --s3-prefix clusters/prod/

# Apply restore
klastr snapshot restore before-upgrade --template template.yaml
```

### `snapshot list`

Display all saved snapshots with metadata.

```bash
klastr snapshot list [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--s3` | `false` | List snapshots in S3 |
| `--s3-bucket` | *(env var)* | S3 bucket name |
| `--s3-prefix` | *(env var)* | S3 key prefix |

#### Examples

```bash
# List local snapshots
$ klastr snapshot list

SNAPSHOTS
─────────────────────────────────────────────────────────────
• before-upgrade
  Cluster: my-cluster (kind)
  Resources: 42
  Created: 2025-01-15 10:30:00

# List S3 snapshots
$ klastr snapshot list --s3 --s3-bucket my-backups

S3 SNAPSHOTS
─────────────────────────────────────────────────────────────
• before-upgrade
• after-migration
```

### `snapshot delete <name>`

Permanently delete a snapshot from disk or S3.

```bash
klastr snapshot delete <name> [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--s3` | `false` | Delete snapshot from S3 |
| `--s3-bucket` | *(env var)* | S3 bucket name |
| `--s3-prefix` | *(env var)* | S3 key prefix |

#### Examples

```bash
# Delete local snapshot
klastr snapshot delete before-upgrade

# Delete S3 snapshot
klastr snapshot delete my-snap --s3 --s3-bucket my-backups
```

---

## S3 Snapshot Storage

Snapshots can be stored in Amazon S3 or S3-compatible services (MinIO, Wasabi, etc.) for backup and disaster recovery.

### Configuration Methods

You can configure S3 in three ways (priority: flags > template config > env vars):

#### 1. Template Configuration (Recommended)

Add to your `template.yaml`:

```yaml
snapshot:
  enabled: true
  bucket: my-k8s-backups
  prefix: clusters/production/
  region: us-east-1
  # endpoint: http://localhost:9000  # For S3-compatible services
```

With this configuration, all snapshot commands automatically use S3:
```bash
klastr snapshot save my-backup    # Saves to S3 automatically
klastr snapshot list              # Lists from S3 automatically
klastr snapshot restore my-backup # Restores from S3 automatically
```

#### 2. Environment Variables

| Environment Variable | Description | Required |
|---------------------|-------------|----------|
| `DEPLOY_CLUSTER_S3_BUCKET` | S3 bucket name | Yes |
| `DEPLOY_CLUSTER_S3_PREFIX` | Key prefix (e.g., `clusters/prod/`) | No |
| `AWS_REGION` | AWS region | Yes |
| `AWS_ACCESS_KEY_ID` | AWS access key | Yes* |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | Yes* |
| `DEPLOY_CLUSTER_S3_ENDPOINT` | Custom endpoint for S3-compatible services | No |

\* Or use IAM roles/instance profiles when running on AWS.

#### 3. Command Flags

Use `--s3` flag with optional configuration flags:

```bash
klastr snapshot save my-backup --s3 \
  --s3-bucket my-k8s-backups \
  --s3-prefix clusters/production/ \
  --s3-region us-east-1
```

### Examples

#### AWS S3 with Template Config

```yaml
# template.yaml
name: production
provider:
  type: kind

snapshot:
  enabled: true
  bucket: my-k8s-backups
  prefix: clusters/production/
  region: us-east-1
```

```bash
# All commands automatically use S3
klastr snapshot save before-upgrade
klastr snapshot list
klastr snapshot restore before-upgrade
```

#### AWS S3 with Environment Variables

```bash
export DEPLOY_CLUSTER_S3_BUCKET=my-k8s-backups
export DEPLOY_CLUSTER_S3_PREFIX=clusters/production/
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...

# Must use --s3 flag with env var config
klastr snapshot save before-upgrade --s3
klastr snapshot list --s3
klastr snapshot restore before-upgrade --s3
```

#### MinIO (S3-Compatible)

```yaml
# template.yaml
snapshot:
  enabled: true
  bucket: k8s-backups
  prefix: clusters/
  region: us-east-1
  endpoint: http://localhost:9000
```

Or via environment:
```bash
export DEPLOY_CLUSTER_S3_BUCKET=k8s-backups
export DEPLOY_CLUSTER_S3_PREFIX=clusters/
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=minioadmin
export AWS_SECRET_ACCESS_KEY=minioadmin
export DEPLOY_CLUSTER_S3_ENDPOINT=http://localhost:9000

klastr snapshot save my-snap --s3
```

### S3 Storage Structure

Snapshots are stored in S3 with the following structure:

```
s3://bucket-name/prefix/snapshot-name/
├── metadata.yaml
├── crds/
├── namespaces/
├── cluster-scoped/
└── namespaced/
    └── namespace-name/
        ├── deployments/
        ├── services/
        └── ...
```

> **Security note:** Snapshots may contain Kubernetes Secrets in plain text. Always encrypt your S3 buckets and use IAM policies to restrict access.
