[package]
name = "deploy-machinery-registry-api"
version = "0.1.0"
description = "KCL module for deploying machinery-registry-api on Kubernetes"

[dependencies]
k8s = "1.31"

[profile]
entries = [
    "main.k"
]
