## 1. Shared Deadline Package

- [x] 1.1 Create `internal/deadline/deadline.go` with ExtendableDeadline (New, Extend, Stop, Reason)
- [x] 1.2 Create `internal/deadline/deadline_test.go` with tests for idle, max_timeout, cancelled, extend, extend-after-done
- [x] 1.3 Convert `internal/app/deadline.go` to thin wrapper with type alias and NewExtendableDeadline delegating to deadline.New

## 2. Config and Error Codes

- [x] 2.1 Add `IdleTimeout time.Duration` field to `AgentConfig` in `internal/config/types.go`
- [x] 2.2 Add `ErrIdleTimeout ErrorCode = "E006"` to `internal/adk/errors.go`
- [x] 2.3 Add `ErrIdleTimeout` case to `UserMessage()` with inactivity-specific messaging

## 3. Session Store AnnotateTimeout

- [x] 3.1 Add `AnnotateTimeout(key, partial string) error` to `session.Store` interface
- [x] 3.2 Implement `AnnotateTimeout` in `EntStore` (synthetic assistant message)
- [x] 3.3 Update all mock implementations (testutil, child_test, middleware_test, state_test)

## 4. Channel Handler Idle Timeout

- [x] 4.1 Add `resolveTimeouts()` helper to `internal/app/channels.go` with 4-way config precedence
- [x] 4.2 Refactor `runAgent()` to use `resolveTimeouts()` and `deadline.New()` when idle > 0
- [x] 4.3 Wire `AnnotateTimeout` on timeout errors and partial-result recovery
- [x] 4.4 Return structured `*adk.AgentError` with `ErrIdleTimeout` or `ErrTimeout` based on `Reason()`

## 5. Gateway Idle Timeout

- [x] 5.1 Add `IdleTimeout` and `MaxTimeout` fields to `gateway.Config`
- [x] 5.2 Refactor `handleChatMessage()` to use `deadline.New()` when IdleTimeout > 0
- [x] 5.3 Pass `WithOnActivity` to `RunStreaming` via runOpts
- [x] 5.4 Add `AnnotateTimeout` call on timeout with reason classification
- [x] 5.5 Update `initGateway()` in `wiring.go` to pass IdleTimeout/MaxTimeout

## 6. Tests

- [x] 6.1 Create `internal/app/channels_test.go` with resolveTimeouts backward-compatibility tests
- [x] 6.2 Verify all existing deadline tests pass via backward-compat alias
- [x] 6.3 Full build (`go build ./...`) and test suite (`go test ./...`) pass with zero failures
