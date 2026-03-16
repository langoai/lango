## Why

`lango serve` cannot be terminated with Ctrl+C when the agent is mid-run (especially during tool-call approval waits). The root cause is that `handleChatMessage()` creates per-request contexts from `context.Background()`, so `Gateway.Shutdown()` has no way to cancel in-flight agent runs. Users must `kill -9` the process.

## What Changes

- Add `shutdownCtx`/`shutdownCancel` fields to `gateway.Server` to provide a cancellable parent context for all in-flight requests.
- Replace `context.Background()` with `s.shutdownCtx` in `handleChatMessage()` context creation.
- Call `shutdownCancel()` at the start of `Shutdown()` so all in-flight request contexts are immediately cancelled.
- Add regression tests verifying shutdown cancels child contexts and pending approval waits.

## Capabilities

### New Capabilities

### Modified Capabilities

- `gateway`: Shutdown now cancels all in-flight request contexts before closing WebSocket connections.

## Impact

- `internal/gateway/server.go` — Server struct, `New()`, `handleChatMessage()`, `Shutdown()`
- `internal/gateway/server_test.go` — New shutdown cancellation tests
- No API changes, no dependency changes, no breaking changes
