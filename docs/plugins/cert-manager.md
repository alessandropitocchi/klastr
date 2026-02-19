# Plugin: Cert-Manager

Installs [cert-manager](https://cert-manager.io/) for automatic TLS certificate management in the cluster.

## Configuration

```yaml
plugins:
  certManager:
    enabled: true
    version: v1.16.3
```

| Field | Type | Default | Required | Description |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | yes | Enable the plugin |
| `version` | string | `v1.16.3` | no | cert-manager version |

## How It Works

The plugin:
1. Applies the official cert-manager manifest from the GitHub release
2. Waits for the `cert-manager-webhook` and `cert-manager` deployments to be ready

## Usage

After installation, you can create resources to obtain automatic TLS certificates.

### Self-signed ClusterIssuer (development)

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned
spec:
  selfSigned: {}
```

### Let's Encrypt ClusterIssuer (production)

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

### TLS on Ingress

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

## Integration with ArgoCD

If the ArgoCD ingress has `tls: true`, the plugin automatically adds cert-manager annotations:

```yaml
argocd:
  ingress:
    enabled: true
    host: argocd.example.com
    tls: true   # Adds cert-manager.io/cluster-issuer annotation
```

## Verification

```bash
kubectl get pods -n cert-manager
kubectl get clusterissuers
kubectl get certificates --all-namespaces
```
