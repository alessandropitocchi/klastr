# Provider: kind

[kind](https://kind.sigs.k8s.io/) (Kubernetes IN Docker) crea cluster Kubernetes utilizzando container Docker come nodi.

## Configurazione

```yaml
provider:
  type: kind
```

## Come funziona

Il provider kind:

1. Genera una configurazione kind (`kind.x-k8s.io/v1alpha4`) basata su `cluster.yaml`
2. Crea i nodi come container Docker (uno per control plane, uno per worker)
3. Configura automaticamente il kubeconfig con context `kind-<nome-cluster>`

## Configurazione automatica per Ingress

Quando il plugin **ingress** e abilitato, il provider aggiunge automaticamente al primo nodo control-plane:

- **Label** `ingress-ready=true` — richiesta dal manifest nginx per kind
- **Port mapping** `80:80` e `443:443` — espone le porte HTTP/HTTPS sull'host

Questa configurazione avviene al momento della creazione del cluster. Se abiliti ingress su un cluster gia esistente, devi aggiungere la label manualmente:

```bash
kubectl label node <cluster>-control-plane ingress-ready=true
```

## Esempio di config kind generata

Con 1 control plane, 2 worker e ingress abilitato:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    image: kindest/node:v1.31.0
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
  - role: worker
    image: kindest/node:v1.31.0
  - role: worker
    image: kindest/node:v1.31.0
```

## Kubeconfig

Il context Kubernetes segue il formato `kind-<nome-cluster>`:

```bash
# Usa il context
kubectl cluster-info --context kind-my-cluster

# Esporta kubeconfig
deploy-cluster get kubeconfig my-cluster > kubeconfig.yaml
```

## Comandi utili

```bash
# Lista cluster kind
deploy-cluster get clusters
# oppure: kind get clusters

# Lista nodi
deploy-cluster get nodes my-cluster
# oppure: kind get nodes --name my-cluster

# Accedi a un nodo Docker
docker exec -it my-cluster-control-plane bash
```

## Limitazioni

- Richiede Docker in esecuzione
- Le porte 80/443 devono essere libere sull'host (se ingress abilitato)
- Il cluster non sopravvive al riavvio di Docker (i container vengono fermati)
- Supporta una sola versione di Kubernetes per cluster
