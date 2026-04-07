## 1. Backend Registry (sandbox/os)

- [x] 1.1 Add `BackendMode` enum with `String()` method
- [x] 1.2 Add `BackendCandidate` and `BackendInfo` structs
- [x] 1.3 Implement `ParseBackendMode(s) (BackendMode, error)` rejecting unknown values
- [x] 1.4 Implement `SelectBackend(mode, candidates)` with auto/none/explicit semantics
- [x] 1.5 Implement `ListBackends(candidates)` for status display
- [x] 1.6 Implement `PlatformBackendCandidates()` shared helper
- [x] 1.7 Add `NewBwrapStub()` and `NewNativeStub()` placeholder isolators
- [x] 1.8 Implement `aggregateUnavailableReasons()` for auto fallback
- [x] 1.9 Add table-driven tests in `registry_test.go`

## 2. Config + Validation

- [x] 2.1 Add `Backend string` field to `SandboxConfig` in `types_sandbox.go`
- [x] 2.2 Set default `Backend: "auto"` in `loader.go DefaultConfig()`
- [x] 2.3 Add `sandboxos.ParseBackendMode` import to `loader.go`
- [x] 2.4 Add backend validation in `config.Validate()` rejecting unknown values

## 3. Skill Executor Fail-Closed

- [x] 3.1 Add `failClosed bool` field to `skill.Executor`
- [x] 3.2 Add `SetFailClosed(fc bool)` method on `Executor`
- [x] 3.3 Block `executeScript()` when `nil isolator + failClosed=true`
- [x] 3.4 Block `executeScript()` when `Apply error + failClosed=true`
- [x] 3.5 Add `SetFailClosed` delegation method to `skill.Registry`
- [x] 3.6 Add `TestExecuteScript_FailClosed_NilIsolator` and `_ApplyError`

## 4. MCP Connection Fail-Closed

- [x] 4.1 Add `failClosed bool` field to `ServerConnection`
- [x] 4.2 Add `SetFailClosed(fc bool)` method on `ServerConnection`
- [x] 4.3 Add `failClosed bool` field to `ServerManager`
- [x] 4.4 Add `SetFailClosed(fc bool)` to `ServerManager` propagating to current and future connections
- [x] 4.5 Block stdio `createTransport()` when `nil isolator + failClosed=true`
- [x] 4.6 Block stdio `createTransport()` when `Apply error + failClosed=true`
- [x] 4.7 Verify HTTP/SSE transports unaffected
- [x] 4.8 Add fail-closed tests + `TestServerManager_SetFailClosed_Propagates`

## 5. Wiring Integration

- [x] 5.1 Update `initOSSandbox()` to use `SelectBackend()` and short-circuit on `backend=none`
- [x] 5.2 Update `supervisor.New()` to use `SelectBackend()` and short-circuit on `backend=none`
- [x] 5.3 Update `wiring_knowledge.go` to call `registry.SetFailClosed()`
- [x] 5.4 Update `wiring_mcp.go` to call `mgr.SetFailClosed()`

## 6. CLI / TUI Exposure

- [x] 6.1 Update `sandbox status` to use `SelectBackend()` + display Backend label
- [x] 6.2 Add Backend Availability section to `sandbox status` using `ListBackends()`
- [x] 6.3 Display `backend=none` as explicit opt-out, hide fail-closed line
- [x] 6.4 Update `sandbox test` to accept `cfgLoader` and use configured backend
- [x] 6.5 Add `os_sandbox_backend` field to `forms_sandbox.go`
- [x] 6.6 Add `os_sandbox_backend` mapping in `state_update.go`

## 7. Documentation

- [x] 7.1 Add `sandbox.backend` row to `README.md` config table
- [x] 7.2 Add `sandbox.backend` row to `docs/configuration.md`
- [x] 7.3 Document Backend Availability section in `docs/cli/sandbox.md`

## 8. Codex Review Round 1

- [x] 8.1 Route backend selection through `supervisor.go` (P1: exec tool was bypassing backend config)
- [x] 8.2 Move backend validation from wiring to `config.Validate()` (P2: prevent silent fallback)
- [x] 8.3 Replace `http.DefaultTransport` with `stubRoundTripper` in `TestHeaderRoundTripper` (P3: no real I/O)

## 9. Codex Review Round 2

- [x] 9.1 Treat `backend=none` as explicit opt-out across all wiring paths (P1)
- [x] 9.2 Update `lango sandbox test` to honor configured backend (P2)

## 10. Codex Review Round 3

- [x] 10.1 Aggregate candidate reasons in auto fallback noop (P2: preserve actionable diagnostics)
- [x] 10.2 Hide fail-closed line and show explicit opt-out label when `backend=none` in sandbox status (P2)

## 11. Verification

- [x] 11.1 `go build ./...` (macOS)
- [x] 11.2 `GOOS=linux GOARCH=amd64 go build ./...` (Linux cross-compile)
- [x] 11.3 `go test ./...` (all tests pass, including new fail-closed and registry tests)
- [x] 11.4 `golangci-lint run ./...` (0 issues)
