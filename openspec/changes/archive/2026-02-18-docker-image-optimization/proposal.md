## Why

The current Docker image is ~1GB, primarily because Chromium (~300-400MB) and curl are always bundled regardless of whether browser features are needed. Since browser tools are optional for most deployments, the image should default to a slim build without Chromium, with opt-in browser support via build arg or sidecar pattern.

## What Changes

- Add `.dockerignore` to exclude unnecessary files from build context (`.git`, `.claude`, `openspec/`, etc.)
- Optimize Dockerfile: add `WITH_BROWSER` build arg (default `false`), use `--no-install-recommends`, remove `curl` dependency, merge RUN layers
- Add `lango health` CLI command to replace `curl` in Docker HEALTHCHECK
- Support remote browser WebSocket connections in `browser.Tool` (config field + `ROD_BROWSER_WS` env fallback)
- Update `docker-compose.yml` with profiles for slim, built-in browser, and Chrome sidecar deployment modes

## Capabilities

### New Capabilities
- `cli-health-check`: CLI health check command (`lango health`) that performs HTTP health check against the gateway, replacing external `curl` dependency in Docker containers
- `remote-browser`: Support for connecting to remote browser instances via WebSocket URL, enabling Chrome sidecar deployment pattern

### Modified Capabilities
- `docker-deployment`: Chromium is now conditional via `WITH_BROWSER` build arg (default false), curl removed, layers merged, compose profiles added for slim/browser/sidecar modes
- `tool-browser`: Browser tool gains `RemoteBrowserURL` config field and `ROD_BROWSER_WS` env var fallback for remote browser connections

## Impact

- **Files**: `Dockerfile`, `.dockerignore` (new), `cmd/lango/main.go`, `internal/tools/browser/browser.go`, `internal/config/types.go`, `internal/app/app.go`, `docker-compose.yml`
- **Image size**: ~1GB → ~200MB (slim), ~550MB (with browser)
- **Dependencies**: `curl` removed from runtime image; health check now uses built-in Go HTTP client
- **Backwards compatibility**: Default `docker compose up` behavior changes (no longer includes Chromium). Users needing browser must use `--build-arg WITH_BROWSER=true` or sidecar profile.
