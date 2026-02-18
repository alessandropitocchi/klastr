# Plugin: Cert-Manager

Installa [cert-manager](https://cert-manager.io/) per la gestione automatica dei certificati TLS nel cluster.

## Configurazione

```yaml
plugins:
  certManager:
    enabled: true
    version: v1.16.3
```

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | si | Abilita il plugin |
| `version` | string | `v1.16.3` | no | Versione di cert-manager |

## Come funziona

Il plugin:
1. Applica il manifest ufficiale cert-manager dalla release GitHub
2. Attende che i deployment `cert-manager-webhook` e `cert-manager` siano pronti

## Utilizzo

Dopo l'installazione puoi creare risorse per ottenere certificati TLS automatici.

### ClusterIssuer self-signed (sviluppo)

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned
spec:
  selfSigned: {}
```

### ClusterIssuer Let's Encrypt (produzione)

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
```

### Certificato su Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - myapp.example.com
      secretName: myapp-tls
  rules:
    - host: myapp.example.com
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

## Integrazione con ArgoCD

Se l'ingress ArgoCD ha `tls: true`, il plugin aggiunge automaticamente le annotation per cert-manager:

```yaml
argocd:
  ingress:
    enabled: true
    host: argocd.example.com
    tls: true   # Aggiunge annotation cert-manager.io/cluster-issuer
```

## Verifica

```bash
kubectl get pods -n cert-manager
kubectl get clusterissuers
kubectl get certificates --all-namespaces
```
