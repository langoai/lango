## Why

google.golang.org/adk v1.0.0 has been released. lango is using v0.6.0 and should upgrade to the GA version to ensure stability and future compatibility. v1.0.0 is fully backward compatible at the Go API level, so only the dependency needs to be updated without code changes.

## What Changes

- Bump `google.golang.org/adk` version from `v0.6.0` → `v1.0.0` in `go.mod`
- Automatic `go.sum` transitive dependency update (`a2a-go v0.3.10`, `go-sdk v1.4.1`, `grpc v1.79.3`, etc.)
- Change `mcptoolset.ConfirmationProvider` → `tool.ConfirmationProvider` type reference in `internal/adk/mcp_spike_test.go` (type moved to `tool` package in v1.0.0)
- No production code changes — all ADK interfaces (`session.Service`, `model.LLM`, `agent.Agent`, `tool.Tool`, `runner.Runner`) are identical or have only additive changes at the source level

## Capabilities

### New Capabilities

(None — this change is a dependency upgrade and does not introduce new features)

### Modified Capabilities

- `adk-architecture`: ADK dependency version changed from v0.6.0 to v1.0.0. Interface contracts are identical, but the available ADK feature surface is expanded (RunOption, AutoCreateSession, HITL tool confirmation, workflow agents, etc.)

## Impact

- **Dependencies**: `google.golang.org/adk v0.6.0` → `v1.0.0`, approximately 5 transitive dependency minor bumps
- **Code**: Only 1 spike test file modified (`internal/adk/mcp_spike_test.go`), 0 lines of production code changed
- **Tests**: Full test suite passing confirmed (`go test ./...` all pass)
- **Build**: `go build ./...` + `go vet ./...` passing confirmed
- **Risk**: LOW — source compatibility verified via module cache diff
