# machinery-registry-api Kubernetes Deployment

KCL module for deploying machinery-registry-api on Kubernetes.

## Render Manifests

```bash
# Default configuration (outputs YAML)
kcl run main.k -D config.registryRepo="stuttgart-things/harvester"

# Output as JSON
kcl run main.k -D config.registryRepo="stuttgart-things/harvester" --format json
```

## Override Variables

Use `-D` flag to override configuration at render time:

```bash
# Override namespace and replicas
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.namespace=production \
  -D config.replicas=3

# Override image
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.image="ghcr.io/stuttgart-things/machinery-registry-api:v1.0.0"

# Custom registry config
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.registryPath="claims/registry.yaml" \
  -D config.registryBranch="main" \
  -D config.syncInterval="30s"

# Enable ingress with custom host
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.ingressEnabled=True \
  -D config.ingressHost="registry-api.example.com"

# Enable Gateway API HTTPRoute (alternative to ingress)
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.httpRouteEnabled=True \
  -D config.httpRouteParentRefName="main-gateway" \
  -D config.httpRouteHostname="registry-api.example.com"

# HTTPRoute with gateway in different namespace
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.httpRouteEnabled=True \
  -D config.httpRouteParentRefName="main-gateway" \
  -D config.httpRouteParentRefNamespace="gateway-system" \
  -D config.httpRouteHostname="registry-api.example.com"
```

## Apply to Cluster

```bash
# Render and apply directly (pipe through yq to split manifests array)
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.namespace=machinery-registry \
  | yq '.manifests | (.[] | splitDoc)' - \
  | kubectl apply -f -

# Dry-run (server-side validation)
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  | yq '.manifests | (.[] | splitDoc)' - \
  | kubectl apply --dry-run=server -f -

# Delete resources
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  | yq '.manifests | (.[] | splitDoc)' - \
  | kubectl delete -f -
```

## Available Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `config.name` | string | `machinery-registry-api` | Application name |
| `config.namespace` | string | `default` | Kubernetes namespace |
| `config.image` | string | `ghcr.io/stuttgart-things/machinery-registry-api:latest` | Container image |
| `config.imagePullPolicy` | string | `IfNotPresent` | Image pull policy |
| `config.replicas` | int | `1` | Number of replicas |
| `config.cpuRequest` | string | `100m` | CPU request |
| `config.cpuLimit` | string | `500m` | CPU limit |
| `config.memoryRequest` | string | `128Mi` | Memory request |
| `config.memoryLimit` | string | `256Mi` | Memory limit |
| `config.serviceType` | string | `ClusterIP` | Service type |
| `config.servicePort` | int | `8090` | Service port |
| `config.containerPort` | int | `8090` | Container port |
| `config.ingressEnabled` | bool | `False` | Enable ingress |
| `config.ingressClassName` | string | `nginx` | Ingress class |
| `config.ingressHost` | string | `machinery-registry-api.example.com` | Ingress hostname |
| `config.ingressTlsEnabled` | bool | `False` | Enable TLS |
| `config.ingressTlsSecretName` | string | `machinery-registry-api-tls` | TLS secret name |
| `config.ingressAnnotations` | {str:str} | `{}` | Ingress annotations |
| `config.httpRouteEnabled` | bool | `False` | Enable Gateway API HTTPRoute |
| `config.httpRouteParentRefName` | string | `` | Gateway name (required when httpRouteEnabled) |
| `config.httpRouteParentRefNamespace` | string | `` | Gateway namespace (optional) |
| `config.httpRouteHostname` | string | `` | HTTPRoute hostname (defaults to ingressHost) |
| `config.httpRouteAnnotations` | {str:str} | `{}` | HTTPRoute annotations |
| `config.registryRepo` | string | `` | GitHub repo slug (required, e.g. `stuttgart-things/harvester`) |
| `config.registryPath` | string | `claims/registry.yaml` | Path to registry file in the repo |
| `config.registryBranch` | string | `main` | Git branch to fetch from |
| `config.syncInterval` | string | `60s` | Polling interval for GitHub sync |
| `config.port` | string | `8090` | Application port (PORT env var) |
| `config.logFormat` | string | `text` | Log format (`text` or `json`) |
| `config.extraEnvVars` | {str:str} | `{}` | Extra environment variables for ConfigMap |
| `config.secrets` | {str:str} | `{}` | Secret key-value pairs, e.g. `{"GITHUB_TOKEN": "ghp_..."}` |
| `config.serviceAccountAnnotations` | {str:str} | `{}` | ServiceAccount annotations |
| `config.labels` | {str:str} | `{}` | Additional labels for resources |
| `config.annotations` | {str:str} | `{}` | Additional annotations for resources |

## Files

| File | Description |
|------|-------------|
| `schema.k` | Configuration schema |
| `labels.k` | Common labels |
| `serviceaccount.k` | ServiceAccount resource |
| `configmap.k` | ConfigMap resource |
| `secret.k` | Secret resource |
| `deploy.k` | Deployment resource |
| `service.k` | Service resource |
| `ingress.k` | Ingress resource |
| `httproute.k` | HTTPRoute resource (Gateway API) |
| `main.k` | Entry point |

## Gateway API Example

The deployment supports [Gateway API](https://gateway-api.sigs.k8s.io/) HTTPRoute as an alternative to Ingress. This requires a Gateway resource deployed on the cluster (e.g. with Cilium).

### Example Gateway (Cilium)

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: whatever-gateway
  namespace: default
spec:
  gatewayClassName: cilium
  listeners:
    - name: https
      port: 443
      protocol: HTTPS
      hostname: "*.whatever.sthings-vsphere.labul.sva.de"
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            name: wildcard-whatever-tls
      allowedRoutes:
        namespaces:
          from: All
    - name: http
      port: 80
      protocol: HTTP
      hostname: "*.whatever.sthings-vsphere.labul.sva.de"
      allowedRoutes:
        namespaces:
          from: All
```

> **Note:** Set `allowedRoutes.namespaces.from: All` on the Gateway listeners to allow HTTPRoutes from other namespaces. With `from: Same` (default), only routes in the Gateway's own namespace are accepted.

### Deploy with HTTPRoute

```bash
# Render and apply with Gateway API HTTPRoute
kcl run main.k \
  -D config.registryRepo="stuttgart-things/harvester" \
  -D config.namespace=machinery-registry \
  -D config.httpRouteEnabled=True \
  -D config.httpRouteParentRefName="whatever-gateway" \
  -D config.httpRouteParentRefNamespace="default" \
  -D config.httpRouteHostname="registry-api.whatever.sthings-vsphere.labul.sva.de" \
  | yq '.manifests | (.[] | splitDoc)' - \
  | kubectl apply -f -
```

### Verify HTTPRoute

```bash
# Check HTTPRoute status (should show Accepted: True)
kubectl -n machinery-registry get httproute machinery-registry-api

# Check HTTPRoute details
kubectl -n machinery-registry get httproute machinery-registry-api -o yaml | yq '.status.parents[0].conditions'

# Test endpoint
curl http://registry-api.whatever.sthings-vsphere.labul.sva.de/health
curl http://registry-api.whatever.sthings-vsphere.labul.sva.de/api/v1/claims
```
