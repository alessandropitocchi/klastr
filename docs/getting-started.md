# Getting Started

## Requisiti

| Strumento | Versione | Note |
|-----------|----------|------|
| [Go](https://go.dev/) | 1.21+ | Per compilare il binario |
| [Docker](https://www.docker.com/) | - | Runtime per kind |
| [kind](https://kind.sigs.k8s.io/) | - | Provider Kubernetes locale |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | - | Interazione con il cluster |
| [Helm](https://helm.sh/) | 3.x | Per monitoring, dashboard e customApps |

## Installazione

```bash
git clone https://github.com/alepito/deploy-cluster.git
cd deploy-cluster
go build -o deploy-cluster ./cmd/deploycluster
```

## Quick Start

### 1. Genera la configurazione

```bash
./deploy-cluster init
```

Il wizard interattivo ti guida nella scelta di:
- Nome del cluster e topologia (control planes, workers, versione K8s)
- Plugin da abilitare (storage, ingress, cert-manager, monitoring, dashboard, ArgoCD)
- Hostname per gli ingress (se ingress è abilitato)
- Configurazione ArgoCD (namespace, versione)

Il risultato è un file `cluster.yaml` pronto all'uso.

### 2. Crea il cluster

```bash
./deploy-cluster create --config cluster.yaml
```

Il tool crea il cluster kind e installa automaticamente tutti i plugin configurati nell'ordine corretto.

### 3. Verifica lo stato

```bash
./deploy-cluster status --config cluster.yaml
```

Output di esempio:

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

### 4. Aggiorna la configurazione

Modifica `cluster.yaml` e applica le modifiche senza ricreare il cluster:

```bash
# Preview delle modifiche
./deploy-cluster upgrade --config cluster.yaml --dry-run

# Applica
./deploy-cluster upgrade --config cluster.yaml
```

### 5. Distruggi il cluster

```bash
./deploy-cluster destroy --config cluster.yaml
```

## Prossimi passi

- [Comandi CLI](commands.md) — riferimento completo dei comandi
- [Configurazione](configuration.md) — struttura del file `cluster.yaml`
- [Provider](providers/kind.md) — dettagli sul provider kind
- [Plugin](plugins/) — documentazione di ogni plugin
