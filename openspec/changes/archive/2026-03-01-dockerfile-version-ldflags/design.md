## Context

The Makefile already injects `main.Version` and `main.BuildTime` via ldflags during local builds. The Dockerfile's `go build` uses only `-ldflags="-s -w"`, omitting version injection. Docker images always report `dev (built unknown)`.

## Goals / Non-Goals

**Goals:**
- Inject version and build time into Docker-built binaries via `ARG` + ldflags
- Maintain backward compatibility (default values match current behavior)

**Non-Goals:**
- Automating CI/CD pipeline changes to pass build args
- Changing the Go source code or version display format
- Adding additional metadata (e.g., commit SHA, Go version)

## Decisions

1. **Use Docker `ARG` for build-time injection** — Standard Docker mechanism for parameterized builds. Alternatives like `.env` files or multi-stage variable passing add unnecessary complexity for two simple strings.

2. **Default values `dev` / `unknown`** — Matches the Go source defaults in `cmd/lango/main.go`, ensuring identical behavior when no build args are provided.

3. **Place ARGs in builder stage only** — ARGs are scoped to the build stage where they're used, not leaked to the runtime image.

## Risks / Trade-offs

- [Minimal risk] Build args are visible in `docker history` — Version/BuildTime are not secrets, so this is acceptable.
