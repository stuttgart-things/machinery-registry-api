"""
ConfigMap for machinery-registry-api
"""

import schema
import labels

config = labels.config

configMap = {
    apiVersion: "v1"
    kind: "ConfigMap"
    metadata: {
        name: config.name
        namespace: config.namespace
        labels: labels.commonLabels
    }
    data: {
        REGISTRY_REPO: config.registryRepo
        REGISTRY_PATH: config.registryPath
        REGISTRY_BRANCH: config.registryBranch
        SYNC_INTERVAL: config.syncInterval
        PORT: config.port
        LOG_FORMAT: config.logFormat
        **config.extraEnvVars
    }
}
