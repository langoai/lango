## Context

When `lango serve` is running and the agent is mid-execution (especially waiting for tool approval), pressing Ctrl+C triggers `application.Stop()` → `registry.StopAll()` → `Gateway.Shutdown()`. However, `handleChatMessage()` creates per-request contexts from `context.Background()`, which means `Shutdown()` has no mechanism to cancel in-flight agent runs. The agent continues retrying tool calls (30s approval timeout loops) until the process is forcefully killed.

The signal handler → lifecycle registry → Gateway.Shutdown() chain is already correctly wired. The only gap is the missing link between shutdown and request contexts.

## Goals / Non-Goals

**Goals:**
- Ctrl+C immediately cancels all in-flight agent runs during shutdown
- Pending `RequestApproval` waits return `context.Canceled` instead of looping on `ErrApprovalTimeout`
- Process terminates cleanly within the existing lifecycle timeout (10s)
- Zero behavioral change for normal (non-shutdown) request processing

**Non-Goals:**
- Hot-restart support (Server instances are not reused after shutdown)
- Graceful drain with completion window (all in-flight requests are cancelled immediately)
- Changes to the deadline package or signal handler

## Decisions

### Decision 1: Server-scoped shutdown context as parent for all request contexts

Add `shutdownCtx context.Context` and `shutdownCancel context.CancelFunc` to the `Server` struct, initialized in `New()` via `context.WithCancel(context.Background())`. All per-request contexts in `handleChatMessage()` use `s.shutdownCtx` as parent instead of `context.Background()`.

**Rationale**: This is the minimal change that connects the shutdown signal to in-flight requests. The existing `deadline.New()` and `context.WithTimeout()` already propagate parent cancellation to children. No changes needed in the deadline package, ADK agent, or signal handler.

**Alternative considered**: Tracking individual request cancellation functions in a slice and iterating on shutdown. Rejected — more complex, race-prone, and unnecessary since a parent context achieves the same propagation automatically.

### Decision 2: Call shutdownCancel() before closing WebSocket connections

In `Shutdown()`, call `s.shutdownCancel()` as the first operation, before closing WebSocket clients and stopping the HTTP server. This ensures agent runs observe context cancellation before their streaming connections are torn down, producing clean error paths rather than broken-pipe panics.

## Risks / Trade-offs

- **Server not reusable after shutdown**: `shutdownCtx` is cancelled permanently. This matches the current lifecycle (server is created fresh on each `lango serve`). If hot-restart is ever needed, `shutdownCtx` would need to be re-created. → Acceptable for current design.
- **Immediate cancellation vs. graceful drain**: All in-flight requests are cancelled instantly with no completion grace period. → Acceptable because the agent already handles `ctx.Err()` and produces partial results where possible.
