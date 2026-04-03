## 1. Dependency Upgrade

- [x] 1.1 Change `google.golang.org/adk` version from `v0.6.0` to `v1.0.0` in `go.mod`
- [x] 1.2 Run `go mod tidy` to resolve transitive dependency updates
- [x] 1.3 Verify `go build ./...` passes with zero production code changes

## 2. Spike Test Fix

- [x] 2.1 Update `internal/adk/mcp_spike_test.go` — replace `mcptoolset.ConfirmationProvider` with `tool.ConfirmationProvider` and add `"google.golang.org/adk/tool"` import
- [x] 2.2 Update version comment references from `v0.6.0` to `v1.0.0` in spike test header

## 3. Verification

- [x] 3.1 Run `go vet ./...` — must pass with no errors
- [x] 3.2 Run `go test ./internal/adk/... ./internal/orchestration/... ./internal/a2a/... ./internal/agentrt/... ./internal/turnrunner/... -count=1` — all ADK adapter tests must pass
- [x] 3.3 Run `go test ./... -count=1` — full test suite must pass
