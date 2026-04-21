---
title: Build & Test
---

# Build & Test

## Prerequisites

The default runtime does **not** require CGO. A C compiler is only needed if you explicitly build optional legacy `vec` integrations.

### Build Tags

| Tag | Purpose | Required? |
|-----|---------|-----------|
| `vec` | Legacy sqlite-vec semantic vector search integration | Optional |
| `kms_aws` | AWS KMS signer provider | Optional |
| `kms_gcp` | GCP Cloud KMS signer provider | Optional |
| `kms_azure` | Azure Key Vault signer provider | Optional |
| `kms_pkcs11` | PKCS#11 / HSM signer provider | Optional |
| `kms_all` | All KMS providers above | Optional |
| `integration` | Include integration tests | Optional |

FTS5 is part of the default runtime. Add `vec` only if you explicitly want the legacy sqlite-vec integration. KMS tags pull in cloud-specific SDKs and are only needed when using HSM or cloud key management for P2P signing. Without any `kms_*` tag, stub providers are compiled in and KMS features are unavailable.

```bash
# Default build (FTS5 included)
go build ./cmd/lango

# Optional legacy build with sqlite-vec integration
CGO_ENABLED=1 go build -tags "vec" ./cmd/lango

# Build with AWS KMS support
go build -tags "kms_aws" ./cmd/lango

# Build with all KMS providers + vector search
CGO_ENABLED=1 go build -tags "vec,kms_all" ./cmd/lango

# Run integration tests
go test -tags "integration" ./...
```

The Makefile uses the default runtime for normal `build` and `test` targets.

## Documentation

For the in-progress docs migration, the local Zensical build path is:

```bash
.venv/bin/zensical build
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| **Build & Install** | |
| `build` | Build binary for current platform |
| `build-linux` | Cross-compile for Linux amd64 |
| `build-darwin` | Cross-compile for macOS arm64 |
| `build-all` | Build for all platforms |
| `install` | Install binary to `$GOPATH/bin` |
| **Development** | |
| `dev` | Build and run server locally |
| `run` | Run server from existing binary |
| **Testing** | |
| `test` | Run tests with race detector and coverage |
| `test-short` | Run short tests only |
| `test-p2p` | Run P2P and wallet spending tests |
| `test-security` | Run security, sandbox, and keyring tests |
| `test-graph` | Run graph store and GraphRAG tests |
| `test-mcp` | Run MCP plugin integration tests |
| `test-economy` | Run economy layer tests (budget, escrow, pricing) |
| `bench` | Run benchmarks |
| `coverage` | Generate HTML coverage report |
| **Code Quality** | |
| `fmt` | Format code |
| `fmt-check` | Check code formatting (CI) |
| `vet` | Run `go vet` |
| `lint` | Run `golangci-lint` (auto-installs if missing) |
| `generate` | Run `go generate` (Ent code) |
| `check-abi` | Verify ABI bindings match Solidity sources |
| `ci` | Full local CI pipeline (`fmt-check` -> `vet` -> `lint` -> `test`) |
| **Dependencies** | |
| `deps` | Download and tidy dependencies |
| **Code Signing** | |
| `codesign` | Sign macOS binary with Apple Developer ID (requires `APPLE_IDENTITY`) |
| **Sandbox** | |
| `sandbox-image` | Build sandbox Docker image for P2P tool isolation |
| **Docker** | |
| `docker-build` | Build Docker image |
| `docker-push` | Push to registry (requires `REGISTRY`) |
| `docker-up` | Start Docker Compose services |
| `docker-down` | Stop Docker Compose services |
| `docker-logs` | View Docker Compose logs |
| **Release** | |
| `release-dry` | Test GoReleaser build locally (current platform, snapshot) |
| `release-check` | Validate `.goreleaser.yaml` configuration |
| **Utility** | |
| `health` | Check running server health |
| `clean` | Remove build artifacts and coverage reports |
| `help` | Show available targets |

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

Run domain-scoped tests to iterate on a specific subsystem:

```bash
make test-p2p        # P2P networking + wallet spending
make test-security   # Security, sandbox, keyring
make test-graph      # Graph store + GraphRAG
make test-mcp        # MCP plugin integration
make test-economy    # Economy layer (budget, escrow, pricing)
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
