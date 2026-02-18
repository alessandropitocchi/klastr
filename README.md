# deploy-cluster

CLI tool per il deploy di cluster Kubernetes locali con supporto plugin.

Permette di creare cluster con topologia configurabile e installare automaticamente componenti come storage, ingress, cert-manager, monitoring, dashboard, ArgoCD e applicazioni custom via Helm, definendo tutto da un singolo file di configurazione.

## Requisiti

- Go 1.21+ | Docker | [kind](https://kind.sigs.k8s.io/) | kubectl | [Helm](https://helm.sh/) 3.x

## Installazione

```bash
go build -o deploy-cluster ./cmd/deploycluster
```

## Quick Start

```bash
# Wizard interattivo per generare la configurazione
./deploy-cluster init

# Crea il cluster con tutti i plugin configurati
./deploy-cluster create --config cluster.yaml

# Verifica lo stato
./deploy-cluster status --config cluster.yaml

# Aggiorna i plugin senza ricreare il cluster
./deploy-cluster upgrade --config cluster.yaml

# Distruggi il cluster
./deploy-cluster destroy --config cluster.yaml
```

## Esempio configurazione

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

## Plugin disponibili

| Plugin | Descrizione | Installazione |
|--------|-------------|---------------|
| [Storage](docs/plugins/storage.md) | StorageClass provisioner (local-path) | kubectl apply |
| [Ingress](docs/plugins/ingress.md) | Controller NGINX per kind | kubectl apply |
| [Cert-Manager](docs/plugins/cert-manager.md) | Gestione certificati TLS | kubectl apply |
| [Monitoring](docs/plugins/monitoring.md) | Prometheus + Grafana | Helm |
| [Dashboard](docs/plugins/dashboard.md) | Headlamp | Helm |
| [Custom Apps](docs/plugins/custom-apps.md) | Chart Helm personalizzati | Helm |
| [ArgoCD](docs/plugins/argocd.md) | GitOps | kubectl apply |

## Accesso alle UI

Con ingress abilitato:

| Servizio | URL | Credenziali |
|----------|-----|-------------|
| Grafana | `http://grafana.localhost` | admin / prom-operator |
| Headlamp | `http://headlamp.localhost` | `kubectl create token headlamp -n headlamp` |
| ArgoCD | `http://argocd.localhost` | admin / `kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" \| base64 -d` |

## Documentazione

| Sezione | Descrizione |
|---------|-------------|
| [Getting Started](docs/getting-started.md) | Installazione e primo utilizzo |
| [Comandi CLI](docs/commands.md) | Riferimento completo dei comandi |
| [Configurazione](docs/configuration.md) | Struttura del file `cluster.yaml` |
| [Provider: kind](docs/providers/kind.md) | Dettagli sul provider kind |
| [Plugin](docs/plugins/) | Documentazione di ogni plugin |
