# Plugin: Ingress

Installs an ingress controller in the cluster to expose services via HTTP/HTTPS through hostnames.

## Configuration

```yaml
plugins:
  ingress:
    enabled: true
    type: nginx
```

| Field | Type | Default | Required | Description |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | yes | Enable the plugin |
| `type` | string | - | yes | Controller type |

## Supported Types

### `nginx`

[ingress-nginx](https://kubernetes.github.io/ingress-nginx/) — official NGINX controller for Kubernetes.

Uses the kind-specific manifest (`deploy/static/provider/kind/deploy.yaml`) which:
- Configures the controller as a `DaemonSet` with `hostPort`
- Uses `nodeSelector` with the `ingress-ready=true` label
- Integrates with kind node port mappings

## Prerequisites

The kind provider automatically configures the control-plane node with:
- Label `ingress-ready=true`
- Port mapping `80:80` and `443:443`

This only happens at cluster creation. If you add ingress afterwards, you need the label manually:

```bash
kubectl label node <cluster>-control-plane ingress-ready=true
```

## How It Works

After installation, create `Ingress` resources to expose services:

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

The service becomes reachable at `http://myapp.localhost`.

## Integration with Other Plugins

When ingress is enabled, several plugins can automatically create an Ingress for their UI:

| Plugin | Default Hostname | Configuration |
|--------|-----------------|---------------|
| Monitoring (Grafana) | `grafana.localhost` | `monitoring.ingress.host` |
| Dashboard (Headlamp) | `headlamp.localhost` | `dashboard.ingress.host` |
| ArgoCD | `argocd.localhost` | `argocd.ingress.host` |
| Custom Apps | configurable | `customApps[].ingress.host` |

## Verification

```bash
# Controller
kubectl get pods -n ingress-nginx
kubectl get svc -n ingress-nginx

# Created ingresses
kubectl get ingress --all-namespaces

# Test
curl http://argocd.localhost
```

## Notes

- Host ports 80 and 443 must be free
- On macOS/Linux, `*.localhost` resolves automatically to `127.0.0.1`
- For custom hostnames, add an entry to `/etc/hosts`
