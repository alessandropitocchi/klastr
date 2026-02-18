# Plugin: Ingress

Installa un ingress controller nel cluster per esporre i servizi via HTTP/HTTPS attraverso hostname.

## Configurazione

```yaml
plugins:
  ingress:
    enabled: true
    type: nginx
```

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | si | Abilita il plugin |
| `type` | string | - | si | Tipo di controller |

## Tipi supportati

### `nginx`

[ingress-nginx](https://kubernetes.github.io/ingress-nginx/) — controller ufficiale NGINX per Kubernetes.

Usa il manifest specifico per kind (`deploy/static/provider/kind/deploy.yaml`) che:
- Configura il controller come `DaemonSet` con `hostPort`
- Usa `nodeSelector` con label `ingress-ready=true`
- Si integra con i port mapping del nodo kind

## Prerequisiti

Il provider kind configura automaticamente il nodo control-plane con:
- Label `ingress-ready=true`
- Port mapping `80:80` e `443:443`

Questo avviene solo alla creazione del cluster. Se aggiungi ingress dopo, serve la label manuale:

```bash
kubectl label node <cluster>-control-plane ingress-ready=true
```

## Come funziona

Dopo l'installazione, crei risorse `Ingress` per esporre i servizi:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
spec:
  ingressClassName: nginx
  rules:
    - host: myapp.localhost
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 80
```

Il servizio diventa raggiungibile su `http://myapp.localhost`.

## Integrazione con altri plugin

Quando ingress e abilitato, diversi plugin possono creare automaticamente un Ingress per la propria UI:

| Plugin | Hostname di default | Configurazione |
|--------|-------------------|----------------|
| Monitoring (Grafana) | `grafana.localhost` | `monitoring.ingress.host` |
| Dashboard (Headlamp) | `headlamp.localhost` | `dashboard.ingress.host` |
| ArgoCD | `argocd.localhost` | `argocd.ingress.host` |
| Custom Apps | configurabile | `customApps[].ingress.host` |

## Verifica

```bash
# Controller
kubectl get pods -n ingress-nginx
kubectl get svc -n ingress-nginx

# Ingress creati
kubectl get ingress --all-namespaces

# Test
curl http://argocd.localhost
```

## Note

- Le porte 80 e 443 dell'host devono essere libere
- Su macOS/Linux, `*.localhost` risolve automaticamente a `127.0.0.1`
- Per hostname custom, aggiungi l'entry in `/etc/hosts`
