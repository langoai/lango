---
title: Build & Test
---

# Build & Test

## Prerequisites

Lango requires **CGO** for sqlite3 dependencies. Ensure CGO is enabled:

```bash
export CGO_ENABLED=1
```

You also need a C compiler (`gcc` or `clang`) installed on your system.

### Build Tags

| Tags | Features |
|------|----------|
| `fts5` | Full-text search (default) |
| `fts5,vec` | Full-text search + semantic vector search (requires sqlite-vec) |

The `fts5` tag is sufficient for most use cases. Add `vec` only when you need embedding-based semantic search (RAG). sqlite-vec is an **optional** dependency; without the `vec` tag the binary compiles and runs without it.

```bash
# FTS5-only build (default, no sqlite-vec required)
CGO_ENABLED=1 go build -tags fts5 ./cmd/lango

# Full build with semantic vector search
CGO_ENABLED=1 go build -tags "fts5,vec" ./cmd/lango
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build binary for current platform |
| `build-linux` | Cross-compile for Linux amd64 |
| `build-darwin` | Cross-compile for macOS arm64 |
| `build-all` | Build for all platforms |
| `install` | Install binary to `$GOPATH/bin` |
| `dev` | Build and run server locally |
| `run` | Run server from existing binary |
| `test` | Run tests with race detector and coverage |
| `test-short` | Run short tests only |
| `bench` | Run benchmarks |
| `coverage` | Generate HTML coverage report |
| `fmt` | Format code |
| `fmt-check` | Check code formatting (CI) |
| `vet` | Run `go vet` |
| `lint` | Run `golangci-lint` |
| `generate` | Run `go generate` (Ent code) |
| `ci` | Full local CI pipeline (`fmt-check` -> `vet` -> `lint` -> `test`) |
| `deps` | Download and tidy dependencies |
| `docker-build` | Build Docker image |
| `docker-push` | Push to registry |
| `docker-up` | Start Docker Compose services |
| `docker-down` | Stop Docker Compose services |
| `docker-logs` | View Docker Compose logs |
| `health` | Check running server health |
| `clean` | Remove build artifacts |

## Development Workflow

The typical workflow for local development:

```bash
# 1. Download and tidy dependencies
make deps

# 2. Generate Ent ORM code
make generate

# 3. Build the binary
make build

# 4. Run tests
make test

# 5. Lint
make lint
```

Or run the full CI pipeline locally:

```bash
make ci
```

### Code Generation

Lango uses [Ent ORM](https://entgo.io/) for database schema management. After modifying any schema in the `ent/schema/` directory, regenerate the ORM code:

```bash
make generate
```

### Running Locally

Build and start the server in one step:

```bash
make dev
```

Or run from an existing binary:

```bash
make run
```

### Testing

Run the full test suite with race detection and coverage:

```bash
make test
```

Run only short tests (skip integration tests):

```bash
make test-short
```

Generate an HTML coverage report:

```bash
make coverage
```

Run benchmarks:

```bash
make bench
```

## Related

- [Installation](../getting-started/installation.md) -- Build from source
- [Docker](../deployment/docker.md) -- Container-based builds
- [Architecture](../architecture/index.md) -- System design and project structure
