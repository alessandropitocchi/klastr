# Plugin: ArgoCD

Installa [ArgoCD](https://argo-cd.readthedocs.io/) per il GitOps — sincronizzazione automatica delle applicazioni da repository Git.

## Configurazione

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

## Campi base

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | si | Abilita il plugin |
| `namespace` | string | `argocd` | no | Namespace di installazione |
| `version` | string | `stable` | no | Versione di ArgoCD |

## Ingress

| Campo | Tipo | Default | Descrizione |
|-------|------|---------|-------------|
| `ingress.enabled` | bool | `false` | Crea un Ingress per la UI |
| `ingress.host` | string | - | Hostname (es. `argocd.localhost`) |
| `ingress.tls` | bool | `false` | Abilita TLS via cert-manager |

Quando l'ingress e abilitato, il plugin:
1. Configura `server.insecure: "true"` nella ConfigMap `argocd-cmd-params-cm` (disabilita TLS interno)
2. Esegue `rollout restart` di `argocd-server` per applicare la modifica
3. Crea una risorsa Ingress con `ingressClassName: nginx`

## Accesso alla UI

### Con ingress

```
http://argocd.localhost
```

### Senza ingress (port-forward)

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
# https://localhost:8080
```

### Password admin

```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

---

## Repository (`repos`)

I repository vengono configurati come Secret Kubernetes con label `argocd.argoproj.io/secret-type: repository`.

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `name` | string | auto-generato dall'URL | no | Nome della repository |
| `url` | string | - | si | URL della repository |
| `type` | string | `git` | no | Tipo: `git` o `helm` |
| `insecure` | bool | auto | no | Salta verifica TLS. Auto: `true` per URL non HTTPS |
| `username` | string | - | no | Username per repo HTTPS |
| `password` | string | - | no | Password/token per repo HTTPS |
| `sshKeyEnv` | string | - | no | Nome variabile d'ambiente con chiave SSH |
| `sshKeyFile` | string | - | no | Path al file della chiave SSH |

### Autenticazione SSH da file

```yaml
repos:
  - name: private-repo
    url: git@github.com:user/repo.git
    sshKeyFile: ~/.ssh/id_ed25519
```

### Autenticazione SSH da variabile d'ambiente

File `.env`:

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

Lancia con `--env .env`:

```bash
deploy-cluster create --config cluster.yaml --env .env
```

### Autenticazione HTTPS con token

```yaml
repos:
  - name: private-repo
    url: https://github.com/user/repo.git
    username: git
    password: ghp_xxxxxxxxxxxxx
```

### Repository Helm

```yaml
repos:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
    type: helm
```

---

## Applicazioni (`apps`)

Le applicazioni vengono create come risorse ArgoCD `Application`.

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `name` | string | - | si | Nome dell'Application |
| `namespace` | string | `default` | no | Namespace di destinazione |
| `project` | string | `default` | no | Progetto ArgoCD |
| `repoURL` | string | - | si | URL del chart repo o del repo Git |
| `chart` | string | - | no | Nome del chart Helm (per Helm repo) |
| `path` | string | `.` | no | Path nel repo Git (per sorgenti Git) |
| `targetRevision` | string | `HEAD` | no | Versione del chart o branch/tag |
| `values` | map | - | no | Valori Helm inline |
| `valuesFile` | string | - | no | Path a un file di values esterno |
| `autoSync` | bool | `true` | no | Abilita sync automatico con prune e selfHeal |

### Helm chart da repository pubblica

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

### Helm chart con values da file

```yaml
apps:
  - name: nginx
    namespace: demo-app
    repoURL: https://charts.bitnami.com/bitnami
    chart: nginx
    targetRevision: 18.2.4
    valuesFile: ./nginx-values.yaml
```

### Manifesti da Git repo

```yaml
apps:
  - name: my-app
    namespace: demo-app
    repoURL: git@github.com:user/gitops-repo.git
    path: environments/dev
    targetRevision: main
```

### Sync manuale

```yaml
apps:
  - name: my-app
    namespace: demo-app
    repoURL: git@github.com:user/gitops-repo.git
    path: environments/prod
    targetRevision: main
    autoSync: false    # Richiede sync manuale dalla UI o CLI
```

---

## Upgrade (diff-based)

Il comando `upgrade` per ArgoCD e intelligente:

1. **Manifest ArgoCD**: ri-applica il manifest (aggiorna la versione se cambiata)
2. **Ingress**: ri-configura se abilitato
3. **Repository**: applica tutti quelli desiderati (idempotente). Rimuove quelli non piu presenti nel config
4. **Applicazioni**: applica tutte quelle desiderate (idempotente). Rimuove quelle non piu presenti nel config

### Dry-run

```bash
deploy-cluster upgrade --config cluster.yaml --dry-run
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

### Disabilitazione

Se ArgoCD e disabilitato nel config ma ancora installato nel cluster, il tool mostra un warning senza disinstallarlo automaticamente:

```
[argocd] WARNING: ArgoCD is installed but disabled in config. It will NOT be automatically uninstalled.
[argocd] To uninstall manually: kubectl delete namespace argocd --context kind-my-cluster
```

## Verifica

```bash
# Pod ArgoCD
kubectl get pods -n argocd

# Repository configurati
kubectl get secrets -n argocd -l argocd.argoproj.io/secret-type=repository

# Applicazioni
kubectl get applications -n argocd

# Stato sync di un'app
kubectl get application nginx -n argocd -o jsonpath='{.status.sync.status}'
```
