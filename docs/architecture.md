# Architettura deploy-cluster

## Overview

deploy-cluster è un CLI tool scritto in Go per il provisioning di cluster Kubernetes locali con installazione automatica di plugin. L'architettura è basata su due astrazioni principali: **Provider** e **Plugin**.

## Struttura del progetto

```
deploy-cluster/
├── cmd/deploy-cluster/          # Entrypoint CLI
│   ├── main.go                  # Main
│   ├── root.go                  # Root command (cobra)
│   ├── init.go                  # Comando init: genera cluster.yaml
│   ├── create.go                # Comando create: crea cluster + plugin
│   ├── destroy.go               # Comando destroy: distrugge cluster
│   └── get.go                   # Comando get: info cluster/nodi/kubeconfig
├── pkg/
│   ├── config/
│   │   ├── config.go            # Strutture di configurazione e parsing YAML
│   │   └── env.go               # Loader per file .env
│   ├── provider/
│   │   ├── provider.go          # Interfaccia Provider
│   │   └── kind/
│   │       └── kind.go          # Implementazione provider kind
│   └── plugin/
│       ├── plugin.go            # Interfaccia Plugin
│       └── argocd/
│           └── argocd.go        # Implementazione plugin ArgoCD
├── example-app/                 # Esempi di applicazioni per ArgoCD
│   ├── app/                     # Manifesti plain YAML
│   ├── nginx-helm/              # Chart Helm di esempio
│   └── argocd/                  # Application manifest di esempio
├── cluster.yaml                 # Configurazione cluster (generato da init)
├── .env.example                 # Esempio file .env per i secret
└── CLAUDE.md                    # Istruzioni per Claude Code
```

## Flusso di esecuzione

### `deploy-cluster init`

1. Genera un `Config` con valori di default
2. Serializza in YAML e scrive su file

### `deploy-cluster create`

```
Carica .env → Carica cluster.yaml → Crea cluster (Provider) → Installa plugin
```

1. **Carica .env**: Legge variabili d'ambiente dal file `.env` (opzionale)
2. **Carica config**: Parsing del file `cluster.yaml`
3. **Crea cluster**: Chiama il provider (kind) che:
   - Verifica che kind sia installato
   - Verifica che il cluster non esista già
   - Genera la configurazione kind (nodi, versione, immagine)
   - Esegue `kind create cluster` con output streaming
4. **Installa plugin**: Per ogni plugin abilitato:
   - **ArgoCD**: Crea namespace → Applica manifest → Attende ready → Aggiunge repo → Crea Application

### `deploy-cluster destroy`

1. Carica config o usa il nome passato via flag
2. Chiama `provider.Delete()` che esegue `kind delete cluster`

## Interfacce

### Provider

```go
type Provider interface {
    Name() string
    Create(cfg *config.Config) error
    Delete(name string) error
    GetKubeconfig(name string) (string, error)
    Exists(name string) (bool, error)
}
```

Implementazioni: `kind` (attuale), `k3d` e `minikube` (future)

### Plugin

```go
type Plugin interface {
    Name() string
    Install(kubeconfig string) error
    Uninstall(kubeconfig string) error
    IsInstalled(kubeconfig string) (bool, error)
}
```

> Nota: il plugin ArgoCD usa una firma specifica `Install(cfg *config.ArgoCDConfig, kubecontext string)` per ricevere la configurazione tipizzata. In futuro si potrà standardizzare l'interfaccia.

## Configurazione

La configurazione è definita in `pkg/config/config.go` con le seguenti strutture principali:

- **Config**: Struttura root con nome, provider, cluster e plugin
- **ClusterConfig**: Topologia (controlPlanes, workers, version)
- **ArgoCDConfig**: Installazione ArgoCD con repos e apps
- **ArgoCDRepoConfig**: Repository Git/Helm con autenticazione (SSH, HTTPS)
- **ArgoCDAppConfig**: Application ArgoCD (Helm chart o Git path)

### Gestione dei secret

I secret (chiavi SSH, token) non vanno nel `cluster.yaml`. Due opzioni:

1. **sshKeyFile**: Punta a un file locale (es. `~/.ssh/id_ed25519`)
2. **sshKeyEnv**: Legge da variabile d'ambiente, caricabile da file `.env`

Il file `.env` viene caricato automaticamente prima del parsing della configurazione.

## Plugin ArgoCD - Dettagli

### Installazione

1. Crea il namespace specificato
2. Applica il manifest ufficiale ArgoCD con `--server-side --force-conflicts` (necessario per i CRD grandi)
3. Attende che il deployment `argocd-server` sia pronto

### Repository

Le repository vengono aggiunte come Secret Kubernetes con label `argocd.argoproj.io/secret-type: repository`. Supporta:

- Repository Git pubbliche
- Repository Git private via SSH key
- Repository Git private via HTTPS (username/password o token)
- Repository Helm
- Flag `insecure` auto-rilevato per URL non HTTPS

### Application

Le Application vengono create come risorse CRD `Application` di ArgoCD. Supporta:

- **Helm chart**: Da repository Helm pubblica con `chart` + `targetRevision`
- **Git repo**: Da repository Git con `path` + `targetRevision`
- **Values**: Inline nella configurazione o da file esterno (`valuesFile`)
- **Sync policy**: Auto-sync abilitato di default con prune e selfHeal

## Dipendenze

| Pacchetto | Uso |
|-----------|-----|
| `github.com/spf13/cobra` | Framework CLI |
| `gopkg.in/yaml.v3` | Parsing/serializzazione YAML |

## Roadmap

- [ ] Provider k3d
- [ ] Provider minikube
- [ ] Plugin Storage (local-path-provisioner, OpenEBS)
- [ ] Plugin Ingress (nginx, traefik)
- [ ] Plugin Monitoring (Prometheus, Grafana)
- [ ] Comando `apply` per aggiornare plugin su cluster esistente
- [ ] Standardizzazione interfaccia Plugin
