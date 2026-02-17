"""
Schema definition for machinery-registry-api deployment configuration
"""

schema MachineryRegistryAPI:
    """Configuration for machinery-registry-api Kubernetes deployment"""

    # Metadata
    name: str = "machinery-registry-api"
    namespace: str = "default"

    # Image
    image: str = "ghcr.io/stuttgart-things/machinery-registry-api:latest"
    imagePullPolicy: str = "IfNotPresent"

    # Replicas and resources
    replicas: int = 1
    cpuRequest: str = "100m"
    cpuLimit: str = "500m"
    memoryRequest: str = "128Mi"
    memoryLimit: str = "256Mi"

    # Service
    serviceType: str = "ClusterIP"
    servicePort: int = 8090
    containerPort: int = 8090

    # Ingress
    ingressEnabled: bool = False
    ingressClassName: str = "nginx"
    ingressHost: str = "machinery-registry-api.example.com"
    ingressTlsEnabled: bool = False
    ingressTlsSecretName: str = "machinery-registry-api-tls"
    ingressAnnotations: {str:str} = {}

    # HTTPRoute (Gateway API)
    httpRouteEnabled: bool = False
    httpRouteParentRefName: str = ""
    httpRouteParentRefNamespace: str = ""
    httpRouteHostname: str = ""
    httpRouteAnnotations: {str:str} = {}

    # Registry config (required)
    registryRepo: str = ""  # GitHub repo slug (e.g. stuttgart-things/harvester)
    registryPath: str = "claims/registry.yaml"
    registryBranch: str = "main"
    syncInterval: str = "60s"

    # Application config
    port: str = "8090"
    logFormat: str = "text"

    # Extra environment variables
    extraEnvVars: {str:str} = {}

    # Secrets (will be base64 encoded)
    secrets: {str:str} = {}  # e.g. {"GITHUB_TOKEN": "ghp_..."}

    # ServiceAccount
    serviceAccountAnnotations: {str:str} = {}

    # Labels and annotations
    labels: {str:str} = {}
    annotations: {str:str} = {}

    check:
        replicas >= 0, "replicas must be >= 0"
        servicePort >= 1 and servicePort <= 65535, "servicePort must be valid"
        containerPort >= 1 and containerPort <= 65535, "containerPort must be valid"
        len(registryRepo) > 0, "registryRepo must be set (e.g. stuttgart-things/harvester)"
