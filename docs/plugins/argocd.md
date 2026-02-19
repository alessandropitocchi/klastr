# Plugin: ArgoCD

Installs [ArgoCD](https://argo-cd.readthedocs.io/) for GitOps — automatic application synchronization from Git repositories.

## Configuration

```yaml
plugins:
  argocd:
    enabled: true
    namespace: argocd
    version: stable
    ingress:
      enabled: true
      host: argocd.localhost
    repos:
      - name: my-gitops-repo
        url: git@github.com:user/gitops-repo.git
        type: git
        sshKeyFile: ~/.ssh/id_ed25519
    apps:
      - name: nginx
        namespace: demo-app
        repoURL: https://charts.bitnami.com/bitnami
        chart: nginx
        targetRevision: 18.2.4
        values:
          replicaCount: 2
```

## Base Fields

| Field | Type | Default | Required | Description |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | yes | Enable the plugin |
| `namespace` | string | `argocd` | no | Installation namespace |
| `version` | string | `stable` | no | ArgoCD version |

## Ingress

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ingress.enabled` | bool | `false` | Create an Ingress for the UI |
| `ingress.host` | string | - | Hostname (e.g., `argocd.localhost`) |
| `ingress.tls` | bool | `false` | Enable TLS via cert-manager |

When ingress is enabled, the plugin:
1. Configures `server.insecure: "true"` in the `argocd-cmd-params-cm` ConfigMap (disables internal TLS)
2. Runs `rollout restart` on `argocd-server` to apply the change
3. Creates an Ingress resource with `ingressClassName: nginx`

## UI Access

### With ingress

```
http://argocd.localhost
```

### Without ingress (port-forward)

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
# https://localhost:8080
```

### Admin password

```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

---

## Repositories (`repos`)

Repositories are configured as Kubernetes Secrets with the label `argocd.argoproj.io/secret-type: repository`.

| Field | Type | Default | Required | Description |
|-------|------|---------|:---:|-------------|
| `name` | string | auto-generated from URL | no | Repository name |
| `url` | string | - | yes | Repository URL |
| `type` | string | `git` | no | Type: `git` or `helm` |
| `insecure` | bool | auto | no | Skip TLS verification. Auto: `true` for non-HTTPS URLs |
| `username` | string | - | no | Username for HTTPS repos |
| `password` | string | - | no | Password/token for HTTPS repos |
| `sshKeyEnv` | string | - | no | Environment variable name containing the SSH key |
| `sshKeyFile` | string | - | no | Path to the SSH key file |

### SSH Authentication from File

```yaml
repos:
  - name: private-repo
    url: git@github.com:user/repo.git
    sshKeyFile: ~/.ssh/id_ed25519
```

### SSH Authentication from Environment Variable

`.env` file:

```bash
ARGOCD_SSH_KEY="-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----"
```

Config:

```yaml
repos:
  - name: private-repo
    url: git@github.com:user/repo.git
    sshKeyEnv: ARGOCD_SSH_KEY
```

Run with `--env .env`:

```bash
deploy-cluster create --template template.yaml --env .env
```

### HTTPS Authentication with Token

```yaml
repos:
  - name: private-repo
    url: https://github.com/user/repo.git
    username: git
    password: ghp_xxxxxxxxxxxxx
```

### Helm Repository

```yaml
repos:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
    type: helm
```

---

## Applications (`apps`)

Applications are created as ArgoCD `Application` resources.

| Field | Type | Default | Required | Description |
|-------|------|---------|:---:|-------------|
| `name` | string | - | yes | Application name |
| `namespace` | string | `default` | no | Destination namespace |
| `project` | string | `default` | no | ArgoCD project |
| `repoURL` | string | - | yes | Chart repo URL or Git repo URL |
| `chart` | string | - | no | Helm chart name (for Helm repos) |
| `path` | string | `.` | no | Path in the Git repo (for Git sources) |
| `targetRevision` | string | `HEAD` | no | Chart version or branch/tag |
| `values` | map | - | no | Inline Helm values |
| `valuesFile` | string | - | no | Path to an external values file |
| `autoSync` | bool | `true` | no | Enable automatic sync with prune and selfHeal |

### Helm Chart from Public Repository

```yaml
apps:
  - name: nginx
    namespace: demo-app
    repoURL: https://charts.bitnami.com/bitnami
    chart: nginx
    targetRevision: 18.2.4
    values:
      replicaCount: 3
```

### Helm Chart with Values from File

```yaml
apps:
  - name: nginx
    namespace: demo-app
    repoURL: https://charts.bitnami.com/bitnami
    chart: nginx
    targetRevision: 18.2.4
    valuesFile: ./nginx-values.yaml
```

### Manifests from Git Repo

```yaml
apps:
  - name: my-app
    namespace: demo-app
    repoURL: git@github.com:user/gitops-repo.git
    path: environments/dev
    targetRevision: main
```

### Manual Sync

```yaml
apps:
  - name: my-app
    namespace: demo-app
    repoURL: git@github.com:user/gitops-repo.git
    path: environments/prod
    targetRevision: main
    autoSync: false    # Requires manual sync from UI or CLI
```

---

## Upgrade (diff-based)

The `upgrade` command for ArgoCD is intelligent:

1. **ArgoCD manifest**: re-applies the manifest (updates version if changed)
2. **Ingress**: re-configures if enabled
3. **Repositories**: applies all desired ones (idempotent). Removes those no longer present in template
4. **Applications**: applies all desired ones (idempotent). Removes those no longer present in template

### Dry-run

```bash
deploy-cluster upgrade --template template.yaml --dry-run
```

```
[argocd] Dry-run: version stable, namespace argocd

  Repositories:
    ~ app-repo (update)
    + new-repo (add)
    - old-repo (remove)

  Applications:
    ~ nginx (update)
    + my-app (add)
    - deprecated-app (remove)
```

### Disabling

If ArgoCD is disabled in the template but still installed in the cluster, the tool shows a warning without automatically uninstalling it:

```
[argocd] WARNING: ArgoCD is installed but disabled in template. It will NOT be automatically uninstalled.
[argocd] To uninstall manually: kubectl delete namespace argocd --context kind-my-cluster
```

## Verification

```bash
# ArgoCD pods
kubectl get pods -n argocd

# Configured repositories
kubectl get secrets -n argocd -l argocd.argoproj.io/secret-type=repository

# Applications
kubectl get applications -n argocd

# Sync status of an app
kubectl get application nginx -n argocd -o jsonpath='{.status.sync.status}'
```
