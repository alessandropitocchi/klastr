# Plugin: Dashboard

Installa [Headlamp](https://headlamp.dev/) come dashboard web per Kubernetes, via Helm.

## Configurazione

```yaml
plugins:
  dashboard:
    enabled: true
    type: headlamp
    version: "0.25.0"          # opzionale
    ingress:
      enabled: true
      host: headlamp.localhost
```

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | si | Abilita il plugin |
| `type` | string | - | si | Tipo di dashboard. Valori: `headlamp` |
| `version` | string | `0.25.0` | no | Versione del chart Helm |
| `ingress.enabled` | bool | `false` | no | Crea un Ingress per Headlamp |
| `ingress.host` | string | - | se ingress enabled | Hostname per Headlamp |

## Come funziona

Il plugin:
1. Installa Headlamp via Helm dal chart OCI `oci://ghcr.io/headlamp-k8s/charts/headlamp`
2. Crea un `ClusterRoleBinding` che assegna `cluster-admin` al service account `headlamp`
3. Se ingress abilitato, crea una risorsa Ingress

- Namespace: `headlamp`
- Release name: `headlamp`

## Accesso

### Con ingress

```
http://headlamp.localhost
```

### Senza ingress (port-forward)

```bash
kubectl port-forward svc/headlamp -n headlamp 4466:80
# http://localhost:4466
```

## Autenticazione

Headlamp richiede un token per il login. Il plugin crea un ClusterRoleBinding con `cluster-admin`, quindi puoi usare il token del service account:

```bash
# Crea un token per il service account headlamp
kubectl create token headlamp -n headlamp
```

Copia il token e incollalo nella schermata di login di Headlamp.

Per un token di lunga durata:

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: headlamp-token
  namespace: headlamp
  annotations:
    kubernetes.io/service-account.name: headlamp
type: kubernetes.io/service-account-token
EOF

# Recupera il token
kubectl get secret headlamp-token -n headlamp -o jsonpath="{.data.token}" | base64 -d
```

## Funzionalita

Headlamp offre:
- Vista su cluster, namespace, workload, servizi, storage
- Log dei pod in tempo reale
- Exec nei container
- Editor YAML integrato
- Supporto multi-cluster
- Sistema di plugin estensibile

## Verifica

```bash
kubectl get pods -n headlamp
kubectl get svc -n headlamp
helm list -n headlamp
```
