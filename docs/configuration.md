# Configurazione

Il file `cluster.yaml` definisce l'intera configurazione del cluster e dei plugin. Viene generato dal comando [`init`](commands.md#init) e può essere personalizzato manualmente.

## Struttura

```yaml
name: my-cluster              # Nome del cluster
provider:
  type: kind                  # Provider Kubernetes
cluster:
  controlPlanes: 1            # Numero di control plane
  workers: 2                  # Numero di worker
  version: v1.31.0            # Versione Kubernetes
plugins:
  storage: { ... }            # StorageClass provisioner
  ingress: { ... }            # Ingress controller
  certManager: { ... }        # Gestione certificati TLS
  monitoring: { ... }         # Prometheus + Grafana
  dashboard: { ... }          # Dashboard Kubernetes
  customApps: [ ... ]         # Chart Helm personalizzati
  argocd: { ... }             # GitOps con ArgoCD
```

## Campi base

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `name` | string | `my-cluster` | si | Nome del cluster (usato come nome kind) |
| `provider.type` | string | - | si | Provider Kubernetes. Valori: `kind` |
| `cluster.controlPlanes` | int | `1` | si | Numero di nodi control plane (min: 1) |
| `cluster.workers` | int | `2` | si | Numero di nodi worker (min: 0) |
| `cluster.version` | string | - | no | Versione Kubernetes (es. `v1.31.0`) |

## Plugin

Ogni plugin ha un campo `enabled: true/false`. I plugin non presenti nel YAML sono considerati disabilitati.

La documentazione dettagliata di ogni plugin si trova nella cartella [plugins/](plugins/):

| Plugin | Documentazione | Descrizione |
|--------|---------------|-------------|
| Storage | [plugins/storage.md](plugins/storage.md) | StorageClass provisioner |
| Ingress | [plugins/ingress.md](plugins/ingress.md) | Ingress controller NGINX |
| Cert-Manager | [plugins/cert-manager.md](plugins/cert-manager.md) | Gestione certificati TLS |
| Monitoring | [plugins/monitoring.md](plugins/monitoring.md) | Prometheus + Grafana via Helm |
| Dashboard | [plugins/dashboard.md](plugins/dashboard.md) | Headlamp via Helm |
| Custom Apps | [plugins/custom-apps.md](plugins/custom-apps.md) | Chart Helm personalizzati |
| ArgoCD | [plugins/argocd.md](plugins/argocd.md) | GitOps con ArgoCD |

## Variabili d'ambiente

Il flag `--env` (default: `.env`) carica variabili d'ambiente da un file. Utile per i secret (es. chiavi SSH per ArgoCD).

Formato del file `.env`:

```bash
ARGOCD_SSH_KEY="-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----"
```

Riferimento nel config:

```yaml
argocd:
  repos:
    - name: private-repo
      url: git@github.com:user/repo.git
      sshKeyEnv: ARGOCD_SSH_KEY
```

## Esempio completo

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
    version: v1.16.3
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
