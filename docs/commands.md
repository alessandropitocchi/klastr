# Comandi CLI

## Panoramica

| Comando | Descrizione |
|---------|-------------|
| [`init`](#init) | Genera un file `cluster.yaml` tramite wizard interattivo |
| [`create`](#create) | Crea il cluster e installa i plugin configurati |
| [`upgrade`](#upgrade) | Aggiorna i plugin di un cluster esistente |
| [`status`](#status) | Mostra lo stato del cluster e dei plugin |
| [`destroy`](#destroy) | Distrugge il cluster |
| [`get`](#get) | Sottcomandi per ottenere informazioni sui cluster |

---

## `init`

Genera un file di configurazione `cluster.yaml` tramite un wizard interattivo che guida nella scelta dei plugin e delle opzioni.

```bash
deploy-cluster init [flags]
```

### Flag

| Flag | Default | Descrizione |
|------|---------|-------------|
| `-o, --output` | `cluster.yaml` | Path del file di output |

### Esempio

```bash
# Genera cluster.yaml nella directory corrente
deploy-cluster init

# Genera con nome custom
deploy-cluster init -o my-cluster.yaml
```

Il wizard chiede in sequenza:
1. **Cluster** — nome, versione Kubernetes, numero di control planes e workers
2. **Plugin** — selezione multipla dei plugin da abilitare
3. **Ingress** — hostname per ogni servizio con UI (se ingress abilitato)
4. **ArgoCD** — namespace e versione (se ArgoCD abilitato)

---

## `create`

Crea un nuovo cluster Kubernetes e installa tutti i plugin abilitati nella configurazione.

```bash
deploy-cluster create [flags]
```

### Flag

| Flag | Default | Descrizione |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | File di configurazione |
| `-e, --env` | `.env` | File con variabili d'ambiente per i secret |

### Ordine di installazione

I plugin vengono installati in quest'ordine:

1. **Storage** — per rendere disponibili i PVC
2. **Ingress** — per esporre i servizi via hostname
3. **Cert-Manager** — per i certificati TLS
4. **Monitoring** — Prometheus + Grafana
5. **Dashboard** — Headlamp
6. **Custom Apps** — chart Helm personalizzati
7. **ArgoCD** — GitOps (per ultimo, potrebbe dipendere dagli altri)

### Esempio

```bash
deploy-cluster create --config cluster.yaml
deploy-cluster create --config cluster.yaml --env production.env
```

---

## `upgrade`

Aggiorna un cluster esistente applicando solo le differenze rispetto alla configurazione attuale. Il cluster non viene ricreato.

```bash
deploy-cluster upgrade [flags]
```

### Flag

| Flag | Default | Descrizione |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | File di configurazione |
| `-e, --env` | `.env` | File con variabili d'ambiente per i secret |
| `--dry-run` | `false` | Mostra le modifiche senza applicarle |

### Comportamento per plugin

| Plugin | Comportamento |
|--------|--------------|
| Storage | Re-apply manifest (idempotente) |
| Ingress | Re-apply manifest (idempotente) |
| Cert-Manager | Re-apply manifest (aggiorna versione se cambiata) |
| Monitoring | `helm upgrade` (idempotente) |
| Dashboard | `helm upgrade` (idempotente) |
| Custom Apps | `helm upgrade --install` per ogni app |
| ArgoCD | Re-apply manifest + diff repos/apps (aggiunge nuovi, rimuove rimossi dal config) |

### Esempio

```bash
# Preview
deploy-cluster upgrade --config cluster.yaml --dry-run

# Applica
deploy-cluster upgrade --config cluster.yaml
```

---

## `status`

Mostra lo stato corrente del cluster: esistenza, plugin installati, repository e applicazioni ArgoCD.

```bash
deploy-cluster status [flags]
```

### Flag

| Flag | Default | Descrizione |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | File di configurazione |

### Esempio

```bash
deploy-cluster status --config cluster.yaml
```

---

## `destroy`

Distrugge il cluster. Elimina tutti i container Docker associati.

```bash
deploy-cluster destroy [flags]
```

### Flag

| Flag | Default | Descrizione |
|------|---------|-------------|
| `-c, --config` | `cluster.yaml` | File di configurazione |
| `-n, --name` | - | Nome del cluster (override del config) |

### Esempio

```bash
deploy-cluster destroy --config cluster.yaml
deploy-cluster destroy --name my-cluster
```

---

## `get`

Sottocomandi per ottenere informazioni sui cluster.

### `get clusters`

Lista tutti i cluster kind esistenti.

```bash
deploy-cluster get clusters
```

### `get nodes <nome>`

Lista i nodi di un cluster specifico.

```bash
deploy-cluster get nodes my-cluster
```

### `get kubeconfig <nome>`

Stampa il kubeconfig di un cluster.

```bash
deploy-cluster get kubeconfig my-cluster

# Salva su file
deploy-cluster get kubeconfig my-cluster > kubeconfig.yaml
```
