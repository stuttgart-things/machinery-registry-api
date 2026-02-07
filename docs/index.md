# machinery-registry-api

Lightweight, read-only REST API that serves the claim inventory from `claims/registry.yaml` in the harvester repo.

## Overview

The `machinery-registry-api` is Phase 2 of the claim-machinery platform roadmap. It periodically syncs the claim registry from a GitHub repository and exposes it via a REST API. The architecture mirrors `claim-machinery-api` (Cobra CLI + Gorilla Mux + middleware stack).

## Architecture

```
                    +-------------------+
                    |   GitHub Repo     |
                    | (harvester)       |
                    | claims/           |
                    |   registry.yaml   |
                    +--------+----------+
                             |
                      HTTP poll (raw URL)
                             |
                    +--------v----------+
                    |   Git Syncer      |
                    | (background poll) |
                    +--------+----------+
                             |
                      RWMutex snapshot
                             |
                    +--------v----------+
                    |   HTTP API Server |
                    |   (Gorilla Mux)   |
                    +-------------------+
                    | GET /health       |
                    | GET /version      |
                    | GET /api/v1/claims|
                    | GET /api/v1/      |
                    |     claims/{name} |
                    +-------------------+
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/version` | Build version info |
| `GET` | `/` | Service index |
| `GET` | `/api/v1/claims` | List all claims (supports query filters) |
| `GET` | `/api/v1/claims/{name}` | Get a single claim by name |
| `GET` | `/openapi.yaml` | OpenAPI 3.0 spec |
| `GET` | `/docs` | Redoc API documentation viewer |

### Query Filters for `/api/v1/claims`

| Parameter | Description |
|-----------|-------------|
| `category` | Filter by claim category (e.g., `cli`) |
| `template` | Filter by template name (e.g., `volumeclaim`) |
| `status` | Filter by status (e.g., `active`) |
| `source` | Filter by source (e.g., `cli`) |

### Response Format

```json
{
  "apiVersion": "claim-registry.io/v1alpha1",
  "kind": "ClaimList",
  "items": [
    {
      "name": "hacky",
      "template": "volumeclaim",
      "category": "cli",
      "namespace": "default",
      "createdAt": "2026-02-05T10:58:33Z",
      "createdBy": "patrick",
      "source": "cli",
      "repository": "stuttgart-things/harvester",
      "path": "claims/cli/hacky.yaml",
      "status": "active"
    }
  ]
}
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
| `DEBUG` | `false` | Enable debug logging |

## Getting Started

### Prerequisites

- Go 1.25.6+
- [Task](https://taskfile.dev/) (optional)

### Installation

```bash
git clone https://github.com/stuttgart-things/machinery-registry-api.git
cd machinery-registry-api
go mod tidy
```

### Running

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

## Project Structure

```
machinery-registry-api/
├── main.go                          # Entry point -> cmd.Execute()
├── cmd/
│   ├── root.go                      # Cobra root command, persistent flags
│   ├── server.go                    # Server command: config, sync, lifecycle
│   ├── version.go                   # Version subcommand
│   └── logo.go                      # ASCII logo
├── internal/
│   ├── api/
│   │   ├── server.go                # Server struct, routes, middleware
│   │   ├── handlers.go              # listClaims, getClaim handlers
│   │   ├── middleware.go            # CORS, requestID, logging, errorHandler
│   │   └── handlers_test.go         # HTTP handler tests
│   ├── registry/
│   │   ├── types.go                 # ClaimRegistry, ClaimEntry structs
│   │   ├── registry.go              # Parse YAML, filter/find helpers
│   │   └── registry_test.go         # Unit tests
│   ├── sync/
│   │   ├── syncer.go                # Git sync via GitHub raw URL
│   │   └── syncer_test.go           # Sync tests with httptest
│   └── version/
│       └── version.go               # Build-time vars
├── docs/
│   ├── index.md                     # This documentation
│   └── openapi.yaml                 # OpenAPI 3.0 spec
├── .goreleaser.yaml                 # Multi-platform release
├── Taskfile.yaml                    # Build/test/lint/run tasks
├── catalog-info.yaml                # Backstage component
├── go.mod / go.sum
└── .ko.yaml, .gitignore, LICENSE
```

## Development

### Available Tasks

| Task | Description |
|------|-------------|
| `task build` | Build the binary |
| `task run` | Run the application |
| `task test` | Run tests |
| `task lint` | Run linter |
| `task fmt` | Format code |
| `task tidy` | Tidy go modules |

## Contributing

1. Fork the repository
1. Create a feature branch (`git checkout -b feat/amazing-feature`)
1. Commit your changes (`git commit -m 'Add amazing feature'`)
1. Push to the branch (`git push origin feat/amazing-feature`)
1. Open a Pull Request
