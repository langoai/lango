## 1. Core Implementation

- [x] 1.1 Modify `SessionServiceAdapter.Get()` in `internal/adk/session_service.go` to auto-create session when store returns "session not found" error
- [x] 1.2 Modify `SessionServiceAdapter.Get()` to auto-create session when store returns nil session
- [x] 1.3 Add `strings` import to `internal/adk/session_service.go`

## 2. Tests

- [x] 2.1 Replace `TestSessionServiceAdapter_GetNotFound` with `TestSessionServiceAdapter_GetAutoCreate` in `internal/adk/state_test.go`
- [x] 2.2 Verify auto-created session has correct ID
- [x] 2.3 Verify auto-created session exists in store after creation

## 3. Verification

- [x] 3.1 Run `go build ./...` to confirm compilation
- [x] 3.2 Run `go test ./internal/adk/...` to confirm all tests pass
- [x] 3.3 Run `go test ./...` to confirm no regressions
