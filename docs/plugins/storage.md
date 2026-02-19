# Plugin: Storage

Installs a StorageClass provisioner in the cluster so that PersistentVolumeClaims (PVCs) are fulfilled automatically.

## Configuration

```yaml
plugins:
  storage:
    enabled: true
    type: local-path
```

| Field | Type | Default | Required | Description |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | yes | Enable the plugin |
| `type` | string | - | yes | Provisioner type |

## Supported Types

### `local-path`

[Rancher local-path-provisioner](https://github.com/rancher/local-path-provisioner) — creates volumes on the node filesystem. Ideal for local development clusters.

What it does:
- Installs the provisioner in the `local-path-storage` namespace
- Sets `local-path` as the cluster's **default** StorageClass
- Removes the default flag from kind's `standard` StorageClass

## Verification

```bash
kubectl get storageclass
# NAME                   PROVISIONER             RECLAIMPOLICY   DEFAULT
# local-path (default)   rancher.io/local-path   Delete          Yes
# standard               rancher.io/local-path   Delete          No
```

Any PVC without an explicit `storageClassName` will automatically use `local-path`.

## Installation Order

Storage is installed **first** among all plugins, so that PVCs required by monitoring, dashboard, or other components are already available.

## Notes

- The provisioner creates volumes as directories on the Docker node. Data does not survive cluster destruction.
- Not suitable for production environments — designed for local development.
