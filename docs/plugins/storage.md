# Plugin: Storage

Installa un provisioner per StorageClass nel cluster, in modo che i PersistentVolumeClaim (PVC) vengano soddisfatti automaticamente.

## Configurazione

```yaml
plugins:
  storage:
    enabled: true
    type: local-path
```

| Campo | Tipo | Default | Obbligatorio | Descrizione |
|-------|------|---------|:---:|-------------|
| `enabled` | bool | `false` | si | Abilita il plugin |
| `type` | string | - | si | Tipo di provisioner |

## Tipi supportati

### `local-path`

[Rancher local-path-provisioner](https://github.com/rancher/local-path-provisioner) — crea volumi sul filesystem del nodo. Ideale per cluster locali di sviluppo.

Cosa fa:
- Installa il provisioner nel namespace `local-path-storage`
- Imposta `local-path` come StorageClass di **default** del cluster
- Rimuove il flag default dalla StorageClass `standard` di kind

## Verifica

```bash
kubectl get storageclass
# NAME                   PROVISIONER             RECLAIMPOLICY   DEFAULT
# local-path (default)   rancher.io/local-path   Delete          Yes
# standard               rancher.io/local-path   Delete          No
```

Qualsiasi PVC senza `storageClassName` esplicito usera automaticamente `local-path`.

## Ordine di installazione

Storage viene installato **per primo** tra tutti i plugin, in modo che i PVC richiesti da monitoring, dashboard o altri componenti siano gia disponibili.

## Note

- Il provisioner crea i volumi come directory sul nodo Docker. I dati non sopravvivono alla distruzione del cluster.
- Non adatto per ambienti di produzione — progettato per sviluppo locale.
