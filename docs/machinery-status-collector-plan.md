# machinery-status-collector — Implementation Plan

## Overview

The `machinery-status-collector` is Phase 3 of the claim-machinery platform roadmap. It collects Crossplane claim status from multiple Kubernetes clusters and batches updates into pull requests against the central registry repository.

## Architecture

```
  +------------------+     +------------------+     +------------------+
  |  Cluster A       |     |  Cluster B       |     |  Cluster N       |
  |  (informer mode) |     |  (informer mode) |     |  (informer mode) |
  |  K8s Informer    |     |  K8s Informer    |     |  K8s Informer    |
  +--------+---------+     +--------+---------+     +--------+---------+
           |                        |                        |
           |   POST /api/v1/status  |                        |
           +------------------------+------------------------+
                                    |
                           +--------v----------+
                           |  Collector Server  |
                           |  (server mode)     |
                           +---------+----------+
                           | API     | Store    |
                           | Server  | (in-mem) |
                           +---------+----------+
                                    |
                              Reconciler
                              (periodic)
                                    |
                           +--------v----------+
                           |   GitHub API       |
                           |   (REST)           |
                           +--------+----------+
                                    |
                              Create PR
                                    |
                           +--------v----------+
                           |  Registry Repo     |
                           |  (harvester)       |
                           |  claims/           |
                           |    registry.yaml   |
                           +--------------------+
```

## Components

| Component | Mode | Description |
|-----------|------|-------------|
| **Collector Server** | `server` | Central HTTP API that receives status updates, stores them in memory, and periodically reconciles into PRs |
| **Cluster Agent** | `informer` | Kubernetes informer that watches Crossplane claims and POSTs status to the collector |
| **Reconciler** | (internal) | Periodic loop that batches dirty status entries into a single PR against the registry repo |
| **GitHub Client** | (internal) | REST API client for fetching registry, creating branches, committing, and opening PRs |

## Implementation Steps

### #1 — Cobra CLI scaffold + version/logo commands
**Issue:** [#6](https://github.com/stuttgart-things/machinery-status-collector/issues/6)

**Files:** `main.go`, `cmd/root.go`, `cmd/version.go`, `cmd/logo.go`

- Minimal `main.go` calling `cmd.Execute()`
- Root command defaults to server subcommand
- Version command with ldflags (`version`, `date`, `commit`)
- ASCII logo command

---

### #2 — Registry types and YAML parsing
**Issue:** [#7](https://github.com/stuttgart-things/machinery-status-collector/issues/7)

**Files:** `internal/registry/types.go`, `registry.go`, `registry_test.go`

- `ClaimEntry` struct with `StatusMessage`, `LastCheckedAt`
- `RegistryFile` struct (map of cluster → []ClaimEntry)
- Parse, update, and serialize YAML
- Round-trip tests

---

### #3 — In-memory status store
**Issue:** [#8](https://github.com/stuttgart-things/machinery-status-collector/issues/8)

**Files:** `internal/collector/store.go`, `store_test.go`

- Thread-safe map keyed by `cluster/claimRef`
- Dirty tracking (set on write, cleared on flush)
- Concurrent access tests with race detector

---

### #4 — Central API server (HTTP handlers + middleware)
**Issue:** [#9](https://github.com/stuttgart-things/machinery-status-collector/issues/9)

**Files:** `internal/api/server.go`, `handlers.go`, `middleware.go`, `handlers_test.go`

- `POST /api/v1/status` — accept status from cluster agents
- `GET /api/v1/status` — list all collected statuses
- `GET /api/v1/status/{cluster}` — filter by cluster
- `GET /healthz`, `GET /version`
- Middleware: logging, request ID, recovery

---

### #5 — GitHub client for batch PR creation
**Issue:** [#10](https://github.com/stuttgart-things/machinery-status-collector/issues/10)

**Files:** `internal/git/github.go`, `github_test.go`

- Fetch file content + SHA from GitHub API
- Create branch, commit updated file, open PR
- List open PRs for duplicate detection
- Tests with `httptest` mock server

---

### #6 — Reconciler (periodic batch PR logic)
**Issue:** [#11](https://github.com/stuttgart-things/machinery-status-collector/issues/11)

**Files:** `internal/collector/reconciler.go`, `reconciler_test.go`

- Ticker-based reconciliation loop
- Skip if store is clean
- Fetch registry → apply updates → create PR
- Duplicate PR avoidance
- Mark store flushed on success

---

### #7 — Server Cobra subcommand
**Issue:** [#12](https://github.com/stuttgart-things/machinery-status-collector/issues/12)

**Files:** `cmd/server.go`

- Wire store + reconciler + API server
- Environment variable configuration:

| Env Var | Default | Description |
|---------|---------|-------------|
| `COLLECTOR_PORT` | `8080` | HTTP listen port |
| `COLLECTOR_RECONCILE_INTERVAL` | `5m` | Reconcile interval |
| `GITHUB_TOKEN` | (required) | GitHub API token |
| `REGISTRY_REPO_OWNER` | (required) | GitHub org/user |
| `REGISTRY_REPO_NAME` | (required) | GitHub repo name |
| `REGISTRY_FILE_PATH` | (required) | Path to registry YAML |
| `REGISTRY_BASE_BRANCH` | `main` | Base branch |

- Graceful shutdown on SIGINT/SIGTERM

---

### #8 — Kubernetes informer (cluster agent)
**Issue:** [#13](https://github.com/stuttgart-things/machinery-status-collector/issues/13)

**Files:** `internal/informer/watcher.go`, `status.go`, `watcher_test.go`

- Dynamic informer for configurable Crossplane claim GVR
- Extract status from `.status.conditions` (Ready condition)
- POST status JSON to collector API
- Handle missing/empty status gracefully

---

### #9 — Informer Cobra subcommand
**Issue:** [#14](https://github.com/stuttgart-things/machinery-status-collector/issues/14)

**Files:** `cmd/informer.go`

- Environment variable configuration:

| Env Var | Default | Description |
|---------|---------|-------------|
| `CLUSTER_NAME` | (required) | Current cluster name |
| `COLLECTOR_URL` | (required) | Central collector API URL |
| `CLAIM_GROUP` | (required) | Crossplane claim API group |
| `CLAIM_VERSION` | `v1alpha1` | Crossplane claim API version |
| `CLAIM_RESOURCE` | (required) | Crossplane claim resource name |
| `CLAIM_NAMESPACE` | (all) | Namespace to watch |

- In-cluster and kubeconfig authentication support
- Graceful shutdown on SIGINT/SIGTERM

---

### #10 — Build and release configuration
**Issue:** [#15](https://github.com/stuttgart-things/machinery-status-collector/issues/15)

**Files:** `Taskfile.yaml`, `.goreleaser.yaml`, `.ko.yaml`, `catalog-info.yaml`

- Taskfile: build, test, lint, run-server, run-informer, tag
- GoReleaser: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
- ko: container image with chainguard static base
- Backstage catalog entry

---

### #11 — OpenAPI spec and documentation
**Issue:** [#16](https://github.com/stuttgart-things/machinery-status-collector/issues/16)

**Files:** `docs/openapi.yaml`, `docs/index.md`, `mkdocs.yml`

- OpenAPI 3.0.3 spec for all endpoints
- Architecture overview with component diagram
- Environment variable reference for both modes
- MkDocs Material theme configuration

## Project Structure

```
machinery-status-collector/
├── main.go                          # Entry point → cmd.Execute()
├── cmd/
│   ├── root.go                      # Cobra root command
│   ├── server.go                    # Server subcommand (collector mode)
│   ├── informer.go                  # Informer subcommand (cluster agent mode)
│   ├── version.go                   # Version subcommand
│   └── logo.go                      # ASCII logo
├── internal/
│   ├── api/
│   │   ├── server.go                # HTTP server, routes, middleware
│   │   ├── handlers.go              # POST/GET status, health, version
│   │   ├── middleware.go            # Logging, request ID, recovery
│   │   └── handlers_test.go        # Handler tests
│   ├── collector/
│   │   ├── store.go                 # Thread-safe in-memory status store
│   │   ├── store_test.go           # Store tests (incl. race)
│   │   ├── reconciler.go           # Periodic batch PR reconciler
│   │   └── reconciler_test.go      # Reconciler tests
│   ├── registry/
│   │   ├── types.go                 # ClaimEntry, RegistryFile structs
│   │   ├── registry.go             # Parse/update/serialize YAML
│   │   └── registry_test.go        # Registry tests
│   ├── git/
│   │   ├── github.go               # GitHub REST API client
│   │   └── github_test.go          # GitHub client tests
│   └── informer/
│       ├── watcher.go               # K8s dynamic informer for claims
│       ├── status.go                # Status extraction from unstructured
│       └── watcher_test.go          # Watcher tests
├── docs/
│   ├── openapi.yaml                 # OpenAPI 3.0 spec
│   └── index.md                     # Documentation
├── Taskfile.yaml                    # Build/test/run tasks
├── .goreleaser.yaml                 # Multi-platform release
├── .ko.yaml                         # Container image build
├── catalog-info.yaml                # Backstage component
├── mkdocs.yml                       # MkDocs config
├── go.mod / go.sum
└── .gitignore, LICENSE
```

## Data Flow

1. **Cluster agents** (informer mode) watch Crossplane claims via Kubernetes dynamic informers
2. On claim add/update, the agent extracts status from `.status.conditions` and POSTs to the collector
3. **Collector server** stores status updates in a thread-safe in-memory map
4. **Reconciler** fires on a configurable interval, checks if the store is dirty
5. If dirty: fetches current `registry.yaml` from GitHub, applies all status updates, creates a branch + commit + PR
6. Store is marked flushed; next cycle skips if no new updates arrived

## Dependencies

- [cobra](https://github.com/spf13/cobra) — CLI framework
- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) — YAML parsing
- [k8s.io/client-go](https://pkg.go.dev/k8s.io/client-go) — Kubernetes client (informer mode)
- [k8s.io/apimachinery](https://pkg.go.dev/k8s.io/apimachinery) — Unstructured types
