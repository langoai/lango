## Context

The agent runtime (`internal/adk/agent.go`) uses a hard `context.WithTimeout` (default 5 minutes) for every request. When the deadline fires, the Go ADK iterator silently terminates, and all accumulated text in `strings.Builder` is discarded — returning `""` to the caller. Channels display a generic error with no recovery path. The gateway broadcasts `agent.error` with minimal classification.

Current typing indicators (Slack placeholder, Telegram/Discord `ChatTyping`) give no elapsed-time feedback — users cannot tell if the system is still working or stuck.

## Goals / Non-Goals

**Goals:**
- Preserve partial results on timeout/error instead of discarding accumulated text
- Provide structured error types with codes, user hints, and partial result access
- Show progressive elapsed-time indicators across all channels
- Allow timeouts to auto-extend when the agent is actively producing output
- Maintain full backward compatibility (no interface or config breaking changes)

**Non-Goals:**
- Streaming partial results back to the user during an error (only recovered after failure)
- Per-tool timeout configuration (existing `ToolTimeout` already handles this)
- Retry logic for transient model errors (separate concern)
- UI-specific error rendering (channels receive structured data, UI decides presentation)

## Decisions

### 1. `AgentError` as a structured error type (not sentinel errors)
**Rationale**: Need to carry multiple fields (code, partial, elapsed, cause) through the error chain. A struct type with `Unwrap()` integrates with `errors.Is/As` while allowing rich metadata. Sentinel errors (`var ErrTimeout = ...`) can't carry partial results.

**Alternative**: Return `(string, string, error)` tuple with partial as second return — rejected because it changes all caller signatures and is less composable.

### 2. Duck-typed `UserMessage()` interface in channels
**Rationale**: Channels cannot import `internal/adk` without creating dependency issues. Using a local interface `{ UserMessage() string }` with `errors.As` allows channels to extract user-friendly messages without coupling.

**Alternative**: Shared `internal/errmsg` package — rejected as over-engineering for a 5-line function.

### 3. Posted message + edit for progress (not typing indicators)
**Rationale**: Typing indicators (`ChatTyping`) auto-expire and cannot display elapsed time. Posted placeholder messages can be edited with "Thinking... (30s)" and later replaced with the actual response or error.

**Alternative**: Keep typing indicators — rejected because they provide no timing feedback and expire silently.

### 4. `ExtendableDeadline` via timer reset (not context chaining)
**Rationale**: Go's `context.WithTimeout` creates immutable deadlines. Rather than chaining new contexts (which leak goroutines), we use `context.WithCancel` + `time.AfterFunc` with `Reset()`. A max timeout `AfterFunc` ensures absolute bounds.

**Alternative**: Recreating context on each extension — rejected due to goroutine leak risk and complexity.

## Risks / Trade-offs

- **[Partial result quality]** Partial text may be mid-sentence or incomplete → Mitigation: UI prepends partial with a note explaining it's incomplete
- **[Progress update rate limiting]** Editing messages every 15s could hit API rate limits on high-traffic bots → Mitigation: 15s interval is well within Slack (1/s), Telegram (30/min), Discord (5/s) limits
- **[Auto-extend abuse]** Malicious or runaway prompts could extend indefinitely → Mitigation: Hard `MaxRequestTimeout` cap (default: 3x base)
- **[Timer race in ExtendableDeadline]** Timer may fire between check and reset → Mitigation: Mutex-protected `Extend()` and the cancel is idempotent
