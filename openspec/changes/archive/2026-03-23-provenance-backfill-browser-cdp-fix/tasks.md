## 1. Provenance Session Backfill

- [x] 1.1 Add `GetNode()` method to `SessionTree` in `internal/provenance/session_tree.go`
- [x] 1.2 Make `rootObserver` closure idempotent in `internal/app/wiring.go:591-600` (GetNode check before RegisterSession)
- [x] 1.3 Call `rootSessionObserver` from `SessionServiceAdapter.Get()` when returning existing session (`internal/adk/session_service.go:141-145`)
- [x] 1.4 Verify `go test ./internal/adk/... -count=1` passes

## 2. Browser CDP Navigate Retry

- [x] 2.1 Add CDP target error recovery block in `internal/toolchain/mw_browser.go` scoped to `browser_navigate` only
- [x] 2.2 Verify retry uses session reset (`sm.Close()`) before retrying
- [x] 2.3 Verify `browser_action` is NOT retried on CDP error
- [x] 2.4 Verify `go test ./internal/toolchain/... -count=1` passes

## 3. Integration

- [x] 3.1 `go build ./...` passes
- [x] 3.2 `go test ./...` full suite passes
