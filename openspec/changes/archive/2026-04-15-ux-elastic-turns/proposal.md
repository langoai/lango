## Why

The TUI chat currently blocks user input during streaming (`input.go:63` Blur), offers no automatic recovery from transient provider errors (rate limits, connection drops, stale streams), and has no emergency compaction when the context window fills up. Users experience a rigid, fragile agent that stops on errors, hangs on stale connections, and crashes on context overflow — the opposite of the "zero-config extensible UX" vision.

## What Changes

- **Interruptible streaming**: Users can type and submit new input while the agent is streaming. The current turn is cancelled and the new input starts immediately via a `pendingRedirectInput` queue consumed in the `DoneMsg` handler.
- **Self-healing retry loop**: The `Runner` becomes the single retry owner, creating a new context per attempt. Transient provider errors (rate limit, connection, transient) trigger automatic retry with jittered backoff (max 3 attempts). Recovery events are emitted from the Runner-level attempt loop.
- **Stale stream detection**: A watchdog timer in `wrapChunkCallback()` detects when no streaming chunk arrives for 30s (configurable). Stale detection cancels the current attempt context, and the retry loop starts a fresh attempt.
- **Inline emergency compaction**: `ContextAwareModelAdapter` gains a `SessionCompactor` interface to compress session messages when measured token total exceeds 90% of the model window. `budgets.Degraded` is NOT a compaction trigger (it means base prompt itself is too large). Compaction is entirely the context model's responsibility — the Runner never decides `CompressAndRetry`.

## Capabilities

### New Capabilities
- `interruptible-streaming`: Streaming interruption with redirect via `pendingRedirectInput` queue pattern in the TUI chat model
- `turn-retry-loop`: Runner-owned retry orchestration with `RecoveryAction` mapping, per-attempt context creation, and recovery event emission
- `stale-stream-detection`: Chunk-based watchdog timer in the turn runner that cancels stale streaming attempts
- `inline-emergency-compaction`: Session compactor injection into `ContextAwareModelAdapter` for synchronous context overflow recovery

### Modified Capabilities
- `interactive-tui-chat`: `inputAcceptsText()` adds `stateStreaming`, `handleStreamingKey()` handles Enter for redirect, `DoneMsg` handler gains short-circuit for pending redirect
- `agent-error-handling`: `RecoveryAction` type and `recoveryActionFor()` mapping function added to error classification
- `agent-turn-tracing`: Turn trace accumulates attempt/recovery events within a single trace (retry does not create separate traces)
- `context-budget`: `budgets.Degraded` explicitly excluded as compaction trigger; Degraded means base prompt exceeds model window (config issue, not session issue)

## Impact

- **Code**: `internal/cli/chat/chat.go`, `internal/cli/chat/input.go`, `internal/turnrunner/runner.go`, `internal/adk/errors.go`, `internal/adk/agent.go`, `internal/adk/context_model.go`, `internal/app/wiring.go`
- **Interfaces**: New `SessionCompactor` interface in `internal/adk/`; `Runner.Run()` signature unchanged but internal retry loop added
- **Dependencies**: No new external dependencies
- **Risk**: `input.go` state machine change (medium-low); `Runner` retry loop reuses existing `classifyResult` path (low); `ContextAwareModelAdapter` compactor injection follows existing `WithMemory` pattern (low)
