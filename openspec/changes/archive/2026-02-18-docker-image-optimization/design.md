## Context

The current Docker image bundles Chromium (~300-400MB) and curl unconditionally, resulting in ~1GB images. Most deployments don't use browser tools, making this waste significant. The health check uses curl, which is the only reason curl is installed.

## Goals / Non-Goals

**Goals:**
- Reduce default Docker image size from ~1GB to ~200MB
- Make Chromium installation conditional via `WITH_BROWSER` build arg
- Replace curl dependency with built-in `lango health` CLI command
- Support remote browser connections via WebSocket for sidecar pattern
- Provide docker-compose profiles for slim, browser, and sidecar modes

**Non-Goals:**
- Switching base image away from debian:bookworm-slim (Alpine has musl issues with CGO)
- Multi-arch image builds
- Kubernetes deployment manifests

## Decisions

### Decision 1: Build arg for conditional Chromium

**Choice**: `ARG WITH_BROWSER=false` with conditional `apt-get install` in a single RUN layer.

**Rationale**: Build args are the standard Docker mechanism for conditional builds. Defaulting to `false` ensures the common case (no browser) gets the small image. A single RUN layer with shell conditional avoids extra layers.

**Alternative considered**: Separate Dockerfiles (Dockerfile.slim, Dockerfile.browser) — rejected because it duplicates the entire file and increases maintenance burden.

### Decision 2: Built-in health command replacing curl

**Choice**: `lango health` CLI command using Go's `net/http` client.

**Rationale**: curl is ~20MB installed. Since we already have a Go binary, using it for health checks eliminates the dependency entirely. The command is simple (HTTP GET to `/health`) and uses Docker's `CMD` array form for HEALTHCHECK.

### Decision 3: Remote browser via WebSocket URL

**Choice**: Config field `RemoteBrowserURL` + `ROD_BROWSER_WS` env var fallback, checked before local launcher.

**Rationale**: go-rod natively supports `ControlURL()` for remote connections. This enables the sidecar pattern where a headless Chrome container (e.g., `chromedp/headless-shell`) runs separately. The env var fallback allows configuration without config file changes (useful in docker-compose).

### Decision 4: Docker Compose profiles

**Choice**: Three profiles — `default` (slim), `browser` (built-in Chromium), `browser-sidecar` (slim + Chrome container).

**Rationale**: Compose profiles allow a single file to support all deployment modes. Users pick their mode with `--profile`. The sidecar pattern keeps the lango image slim while still providing browser functionality.

## Risks / Trade-offs

- **Breaking change for browser users**: Default build no longer includes Chromium → Users must explicitly opt-in with `WITH_BROWSER=true` or use sidecar profile. Mitigation: documented in compose file and README.
- **Sidecar network latency**: Remote browser adds WebSocket hop → Acceptable for browser automation use cases which are inherently slow (page loads, rendering).
- **Docker Compose profile complexity**: Three deployment modes in one file → Clear naming and comments make selection straightforward.
