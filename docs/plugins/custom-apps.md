# Plugin: Custom Apps

Allows installing any Helm chart without creating a dedicated plugin. Each entry in the list becomes a `helm upgrade --install`.

## Configuration

```yaml
plugins:
  customApps:
    - name: redis
      chart: oci://registry-1.docker.io/bitnamicharts/redis
      version: "21.1.5"
      namespace: redis
      values:
        architecture: standalone
        auth:
          enabled: false
```

| Field | Type | Default | Required | Description |
|-------|------|---------|:---:|-------------|
| `name` | string | - | yes | Helm release name |
| `chart` | string | - | yes | Chart reference (OCI, URL, local path) |
| `version` | string | latest | no | Chart version |
| `namespace` | string | same as `name` | no | Installation namespace |
| `values` | map | - | no | Inline Helm values |
| `valuesFile` | string | - | no | Path to an external values file |
| `ingress.enabled` | bool | `false` | no | Create an Ingress for the app |
| `ingress.host` | string | - | if ingress enabled | Hostname |
| `ingress.serviceName` | string | same as `name` | no | Backend service name |
| `ingress.servicePort` | int | `80` | no | Backend service port |

## Supported Chart Formats

### OCI Chart (recommended)

```yaml
chart: oci://registry-1.docker.io/bitnamicharts/redis
chart: oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
```

### Helm Repository Chart

```yaml
chart: https://charts.bitnami.com/bitnami/redis
```

## Values

### Inline

Values written directly in the config:

```yaml
customApps:
  - name: redis
    chart: oci://registry-1.docker.io/bitnamicharts/redis
    values:
      architecture: standalone
      auth:
        enabled: false
      replica:
        replicaCount: 0
```

### From External File

Path to a YAML file with values:

```yaml
customApps:
  - name: redis
    chart: oci://registry-1.docker.io/bitnamicharts/redis
    valuesFile: ./redis-values.yaml
```

If both are specified, `valuesFile` takes **precedence**.

## Ingress

If the ingress plugin is enabled in the cluster, you can expose apps via hostname:

```yaml
customApps:
  - name: rabbitmq
    chart: oci://registry-1.docker.io/bitnamicharts/rabbitmq
    version: "14.0.0"
    namespace: rabbitmq
    ingress:
      enabled: true
      host: rabbitmq.localhost
      serviceName: rabbitmq        # Kubernetes Service name
      servicePort: 15672           # management UI port
```

| Ingress Field | Default | Description |
|---------------|---------|-------------|
| `host` | required | Hostname for the Ingress |
| `serviceName` | release name | Backend service. Change it if the chart creates a service with a different name than the release |
| `servicePort` | `80` | Service port. Check the chart documentation for the correct port |

## Examples

### Standalone Redis

```yaml
- name: redis
  chart: oci://registry-1.docker.io/bitnamicharts/redis
  version: "21.1.5"
  namespace: redis
  values:
    architecture: standalone
    auth:
      enabled: false
```

### RabbitMQ with Management UI

```yaml
- name: rabbitmq
  chart: oci://registry-1.docker.io/bitnamicharts/rabbitmq
  version: "14.0.0"
  namespace: rabbitmq
  values:
    auth:
      username: admin
      password: admin
  ingress:
    enabled: true
    host: rabbitmq.localhost
    serviceName: rabbitmq
    servicePort: 15672
```

### PostgreSQL

```yaml
- name: postgres
  chart: oci://registry-1.docker.io/bitnamicharts/postgresql
  version: "16.4.4"
  namespace: postgres
  values:
    auth:
      postgresPassword: localdev
      database: myapp
```

### MinIO (Object Storage)

```yaml
- name: minio
  chart: oci://registry-1.docker.io/bitnamicharts/minio
  version: "14.8.5"
  namespace: minio
  values:
    auth:
      rootUser: admin
      rootPassword: localdev123
  ingress:
    enabled: true
    host: minio.localhost
    serviceName: minio
    servicePort: 9001
```

## Upgrade

The `upgrade` command runs `helm upgrade --install` for each app in the list. It is idempotent:
- If the release doesn't exist, it is installed
- If it exists, it is updated with the new values/version

## Dry-run

```bash
deploy-cluster upgrade --config cluster.yaml --dry-run
```

Output:

```
[customApps] Custom apps:
  ~ redis (oci://registry-1.docker.io/bitnamicharts/redis@21.1.5) (update)
  + rabbitmq (oci://registry-1.docker.io/bitnamicharts/rabbitmq@14.0.0) (install)
```

## Status

```bash
deploy-cluster status --config cluster.yaml
```

```
Custom Apps (2 configured):
  - redis: installed
  - rabbitmq: not installed
```

## Verification

```bash
# Helm releases
helm list --all-namespaces

# Pods for a specific app
kubectl get pods -n redis

# Services
kubectl get svc -n redis
```
