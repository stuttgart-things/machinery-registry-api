# machinery-registry-api

Lightweight, read-only REST API that serves the claim inventory from a GitHub-hosted `registry.yaml`.

Part of the [claim-machinery](https://github.com/stuttgart-things) platform (Phase 2). Periodically syncs the claim registry from a GitHub repository and exposes it via a filtered REST API.

## Architecture

```
GitHub Repo (harvester)          machinery-registry-api
┌────────────────────┐          ┌─────────────────────┐
│ claims/             │  HTTP   │  Git Syncer          │
│   registry.yaml    │◄────────│  (background poll)   │
└────────────────────┘  poll    │         │            │
                                │    RWMutex snapshot  │
                                │         │            │
                                │  HTTP API Server     │
                                │  (Gorilla Mux)       │
                                └─────────────────────┘
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/version` | Build version info |
| `GET` | `/api/v1/claims` | List all claims (with query filters) |
| `GET` | `/api/v1/claims/{name}` | Get a single claim by name |
| `GET` | `/openapi.yaml` | OpenAPI 3.0 spec |
| `GET` | `/docs` | Redoc API documentation |

### Query Filters

`/api/v1/claims` supports filtering via query parameters:

```
/api/v1/claims?category=cli&template=volumeclaim&status=active&source=cli
```

## Configuration

| Env Var | Default | Description |
|---------|---------|-------------|
| `REGISTRY_REPO` | (required) | GitHub repo slug, e.g. `stuttgart-things/harvester` |
| `REGISTRY_PATH` | `claims/registry.yaml` | Path to registry file in repo |
| `REGISTRY_BRANCH` | `main` | Git branch |
| `SYNC_INTERVAL` | `60s` | Polling interval |
| `PORT` | `8080` | HTTP server port |
| `GITHUB_TOKEN` | (optional) | For private repos |
| `LOG_FORMAT` | `text` | `json` for structured logging |

## Getting Started

### Prerequisites

- Go 1.25.6+
- [Task](https://taskfile.dev/) (optional)

### Run

```bash
# Using go directly
REGISTRY_REPO=stuttgart-things/harvester go run .

# Using Task
REGISTRY_REPO=stuttgart-things/harvester task run
```

### Verify

```bash
curl localhost:8080/health
curl localhost:8080/version
curl localhost:8080/api/v1/claims
curl localhost:8080/api/v1/claims/hacky
curl "localhost:8080/api/v1/claims?category=cli&template=volumeclaim"
```

### Build

```bash
go build -o bin/machinery-registry-api .
# or
task build
```

### Test

```bash
go test ./...
# or
task test
```

## License

See [LICENSE](LICENSE) for details.
