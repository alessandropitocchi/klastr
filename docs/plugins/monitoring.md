# Plugin: Monitoring

Installa lo stack [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) via Helm, che include Prometheus, Grafana, Alertmanager, node-exporter e kube-state-metrics.

## Configurazione

```yaml
plugins:
  monitoring:
    enabled: true
    type: prometheus
    version: "72.6.2"          # opzionale
    ingress:
      enabled: true
      host: grafana.localhost
```

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | si | Abilita il plugin |
| `type` | string | - | si | Tipo di stack. Valori: `prometheus` |
| `version` | string | `72.6.2` | no | Versione del chart Helm |
| `ingress.enabled` | bool | `false` | no | Crea un Ingress per Grafana |
| `ingress.host` | string | - | se ingress enabled | Hostname per Grafana |

## Come funziona

Il plugin usa Helm per installare il chart OCI:

```
oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
```

- Namespace: `monitoring`
- Release name: `kube-prometheus-stack`
- Flag `--wait` per attendere che tutti i pod siano pronti

## Accesso a Grafana

### Con ingress

Se `ingress.enabled: true`, Grafana e accessibile direttamente:

```
http://grafana.localhost
```

Credenziali: `admin` / `prom-operator`

### Senza ingress (port-forward)

```bash
kubectl port-forward svc/kube-prometheus-stack-grafana -n monitoring 3000:80
# http://localhost:3000
```

## Accesso a Prometheus

```bash
kubectl port-forward svc/kube-prometheus-stack-prometheus -n monitoring 9090:9090
# http://localhost:9090
```

## Accesso ad Alertmanager

```bash
kubectl port-forward svc/kube-prometheus-stack-alertmanager -n monitoring 9093:9093
# http://localhost:9093
```

## Componenti installati

| Componente | Descrizione |
|------------|-------------|
| Prometheus Operator | Gestisce le risorse Prometheus/ServiceMonitor/AlertmanagerConfig |
| Prometheus | Raccolta metriche e query PromQL |
| Grafana | Dashboard e visualizzazione (con dashboard preconfigurate) |
| Alertmanager | Gestione e routing degli alert |
| node-exporter | Metriche del sistema operativo dei nodi |
| kube-state-metrics | Metriche sullo stato degli oggetti Kubernetes |

## Dashboard Grafana preconfigurate

Il chart include dashboard per:
- Kubernetes cluster overview
- Node metrics
- Pod metrics
- Namespace resources
- Persistent volumes
- API server
- etcd
- CoreDNS

## Upgrade

L'upgrade usa `helm upgrade` che aggiorna la release esistente. Cambiare `version` nel config e lanciare `upgrade` aggiorna lo stack.

## Verifica

```bash
# Pod
kubectl get pods -n monitoring

# Servizi
kubectl get svc -n monitoring

# ServiceMonitor
kubectl get servicemonitors -n monitoring

# Helm release
helm list -n monitoring
```
