# P1 Security Hardening — Tasks

## P1-4: OS Keyring Integration

- [x] Create `internal/keyring/keyring.go` with Provider interface, constants, Status type
- [x] Create `internal/keyring/os_keyring.go` with OSProvider using go-keyring, IsAvailable()
- [x] Create `internal/keyring/keyring_test.go` with mock Provider unit tests
- [x] Create `internal/cli/security/keyring.go` with store/clear/status CLI commands
- [x] Add SourceKeyring to passphrase/acquire.go, update Acquire() priority chain
- [x] Wire OSProvider in bootstrap/bootstrap.go when keyring available
- [x] Add KeyringConfig to config/types.go SecurityConfig
- [x] Add keyring defaults to config/loader.go
- [x] Register keyring command in cli/security/migrate.go
- [x] Add github.com/zalando/go-keyring dependency

## P1-5: Tool Execution Process Isolation

- [x] Create `internal/sandbox/executor.go` with Executor interface, Config, Request/Result types
- [x] Create `internal/sandbox/in_process.go` with InProcessExecutor
- [x] Create `internal/sandbox/subprocess.go` with SubprocessExecutor (JSON protocol, clean env, timeout)
- [x] Create `internal/sandbox/worker.go` with RunWorker() and IsWorkerMode()
- [x] Create `internal/sandbox/executor_test.go` with unit tests
- [x] Add ToolIsolationConfig to config/types.go P2PConfig
- [x] Add tool isolation defaults to config/loader.go
- [x] Add sandboxExec field + SetSandboxExecutor() to protocol/handler.go
- [x] Wire SubprocessExecutor in app/app.go when ToolIsolation.Enabled
- [x] Add --sandbox-worker early check in cmd/lango/main.go

## P1-6: Session Explicit Invalidation

- [x] Add InvalidationReason, InvalidationRecord types to handshake/session.go
- [x] Add Invalidate(), InvalidateAll(), InvalidateByCondition(), InvalidationHistory() to SessionStore
- [x] Add SetInvalidationCallback() and update Validate() for invalidation flag
- [x] Create `internal/p2p/handshake/security_events.go` with SecurityEventHandler
- [x] Create `internal/p2p/handshake/session_test.go` with invalidation tests
- [x] Create `internal/p2p/handshake/security_events_test.go` with event handler tests
- [x] Add SecurityEventTracker interface + SetSecurityEvents() to protocol/handler.go
- [x] Track tool success/failure in handleToolInvoke/handleToolInvokePaid
- [x] Add SetOnChangeCallback() to reputation/store.go
- [x] Wire SecurityEventHandler + reputation callback in app/wiring.go
- [x] Create `internal/cli/p2p/session.go` with list/revoke/revoke-all commands
- [x] Register session command in cli/p2p/p2p.go

## Verification

- [x] `go build ./...` — zero compilation errors
- [x] `go test ./...` — all tests pass
- [x] `go vet ./...` — no vet warnings
