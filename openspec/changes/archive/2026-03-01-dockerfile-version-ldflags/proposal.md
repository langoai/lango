## Why

The Makefile injects Version/BuildTime via `-X main.Version` / `-X main.BuildTime` ldflags, but the Dockerfile's `go build` command only uses `-ldflags="-s -w"`. This causes Docker-built images to always show `lango dev (built unknown)` when running `lango version`, losing traceability for containerized deployments.

## What Changes

- Add `ARG VERSION=dev` and `ARG BUILD_TIME=unknown` build arguments to `Dockerfile`
- Extend `go build -ldflags` to include `-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}`
- Docker builds can now inject version info via `--build-arg VERSION=... --build-arg BUILD_TIME=...`
- Default behavior (no build args) remains unchanged (`dev` / `unknown`)

## Capabilities

### New Capabilities

- `docker-version-injection`: Build-time version and build timestamp injection for Docker images via ARG/ldflags

### Modified Capabilities

## Impact

- **Dockerfile**: Lines 19-21 modified (ARG declarations + go build command)
- **CI/CD**: Docker build commands should be updated to pass `--build-arg VERSION=... --build-arg BUILD_TIME=...`
- **No code changes**: Only Dockerfile modification, no Go source changes required
