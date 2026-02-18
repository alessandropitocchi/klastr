# Plugin: Custom Apps

Permette di installare qualsiasi chart Helm senza dover creare un plugin dedicato. Ogni entry nella lista diventa un `helm upgrade --install`.

## Configurazione

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

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `name` | string | - | si | Nome della release Helm |
| `chart` | string | - | si | Riferimento al chart (OCI, URL, path locale) |
| `version` | string | latest | no | Versione del chart |
| `namespace` | string | uguale a `name` | no | Namespace di installazione |
| `values` | map | - | no | Valori Helm inline |
| `valuesFile` | string | - | no | Path a un file di values esterno |
| `ingress.enabled` | bool | `false` | no | Crea un Ingress per l'app |
| `ingress.host` | string | - | se ingress enabled | Hostname |
| `ingress.serviceName` | string | uguale a `name` | no | Nome del service backend |
| `ingress.servicePort` | int | `80` | no | Porta del service backend |

## Formati chart supportati

### Chart OCI (consigliato)

```yaml
chart: oci://registry-1.docker.io/bitnamicharts/redis
chart: oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
```

### Chart da repository Helm

```yaml
chart: https://charts.bitnami.com/bitnami/redis
```

## Values

### Inline

Valori scritti direttamente nel config:

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

### Da file esterno

Path a un file YAML con i values:

```yaml
customApps:
  - name: redis
    chart: oci://registry-1.docker.io/bitnamicharts/redis
    valuesFile: ./redis-values.yaml
```

Se entrambi sono specificati, `valuesFile` ha la **precedenza**.

## Ingress

Se il plugin ingress e abilitato nel cluster, puoi esporre le app via hostname:

```yaml
customApps:
  - name: rabbitmq
    chart: oci://registry-1.docker.io/bitnamicharts/rabbitmq
    version: "14.0.0"
    namespace: rabbitmq
    ingress:
      enabled: true
      host: rabbitmq.localhost
      serviceName: rabbitmq        # nome del Service Kubernetes
      servicePort: 15672           # porta della management UI
```

| Campo Ingress | Default | Descrizione |
|---------------|---------|-------------|
| `host` | obbligatorio | Hostname per l'Ingress |
| `serviceName` | nome della release | Service backend. Cambialo se il chart crea un servizio con nome diverso dalla release |
| `servicePort` | `80` | Porta del service. Controlla la documentazione del chart per la porta corretta |

## Esempi

### Redis standalone

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

### RabbitMQ con management UI

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

### MinIO (object storage)

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

Il comando `upgrade` esegue `helm upgrade --install` per ogni app nella lista. E idempotente:
- Se la release non esiste, viene installata
- Se esiste, viene aggiornata con i nuovi values/versione

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

## Verifica

```bash
# Release Helm
helm list --all-namespaces

# Pod di una specifica app
kubectl get pods -n redis

# Servizi
kubectl get svc -n redis
```
