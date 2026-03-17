## 1. Core Implementation

- [x] 1.1 Add `shutdownCtx context.Context` and `shutdownCancel context.CancelFunc` fields to Server struct
- [x] 1.2 Initialize `shutdownCtx`/`shutdownCancel` via `context.WithCancel(context.Background())` in `New()`
- [x] 1.3 Replace `context.Background()` with `s.shutdownCtx` in `handleChatMessage()` for both deadline and timeout paths
- [x] 1.4 Call `s.shutdownCancel()` as the first operation in `Shutdown()`

## 2. Tests

- [x] 2.1 Add `TestShutdown_CancelsInflightRequestContexts` — verify child context derived from `shutdownCtx` is cancelled when `shutdownCancel()` is called
- [x] 2.2 Add `TestShutdown_CancelsApprovalWait` — verify `RequestApproval()` returns `context.Canceled` (not `ErrApprovalTimeout`) when `shutdownCancel()` is called during wait

## 3. Verification

- [x] 3.1 Run `go build ./...` — confirm no build errors
- [x] 3.2 Run `go test ./internal/gateway/...` — confirm all tests pass
- [x] 3.3 Run `go test ./...` — confirm no regressions
