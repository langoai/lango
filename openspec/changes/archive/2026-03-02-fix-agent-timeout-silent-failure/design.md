## Context

When an agent request times out, users receive no notification and typing indicators persist indefinitely. The root cause is in the Google ADK's genai SDK: `iterateResponseStream` checks `rs.r.Err()` (scanner error) but never checks `ctx.Err()` on stream termination. This causes the iterator to silently complete on context deadline, making `RunAndCollect`/`RunStreaming` return `("", nil)` instead of an error.

This affects all communication paths: Telegram, Discord, Slack channels, and the Gateway WebSocket UI.

## Goals / Non-Goals

**Goals:**
- Detect agent timeout when ADK's iterator silently terminates
- Deliver user-visible error messages on timeout across all channels and Gateway
- Proactively warn Gateway UI users when timeout is approaching (80%)
- Differentiate `agent.error` from `agent.done` events in Gateway WebSocket protocol

**Non-Goals:**
- Fixing the upstream ADK/genai SDK bug (we apply a workaround)
- Adding timeout configuration UI
- Implementing request cancellation by users

## Decisions

### Decision 1: Post-iteration `ctx.Err()` check (workaround pattern)
**Choice**: Add `ctx.Err()` check after the iterator `for range` loop completes in both `runAndCollectOnce` and `RunStreaming`.

**Rationale**: This is the narrowest possible fix that catches the ADK bug at the boundary where our code consumes the iterator. It works regardless of which layer in the ADK/genai stack swallowed the error. The check is a no-op for normal completions (`ctx.Err() == nil`).

**Alternative considered**: Wrapping the ADK iterator with a context-aware wrapper. Rejected because it would add unnecessary complexity — the post-loop check achieves the same result with a single `if` statement.

### Decision 2: Separate `agent.error` event (not error field in `agent.done`)
**Choice**: Introduce a new `agent.error` WebSocket event and send it instead of `agent.done` on failure.

**Rationale**: UI clients can treat both events as "stop thinking" signals while handling them differently for display. This is backward-compatible — existing clients that don't handle `agent.error` will simply not show the error, but the thinking indicator will stop (since `agent.done` is not sent, clients will eventually timeout their own thinking state). New clients can show actionable error messages.

**Alternative considered**: Adding `error` field to `agent.done` payload. Rejected because it changes the semantics of an existing event and requires all existing clients to update their handling.

### Decision 3: 80% timeout warning via `agent.warning` event
**Choice**: Fire a `time.AfterFunc` at 80% of the request timeout that broadcasts `agent.warning`.

**Rationale**: Mirrors the existing pattern in `app/channels.go:runAgent` (which already logs at 80%). Giving the UI a heads-up allows showing "taking longer than expected" before the hard timeout hits.

## Risks / Trade-offs

- **[Risk] ADK fixes the silent termination bug** → Our `ctx.Err()` check becomes a harmless no-op (the iterator would yield the error first, and `ctx.Err()` after normal completion returns `nil`). No migration needed.
- **[Risk] Partial response discarded on timeout** → If the agent streamed partial chunks before timeout, the error replaces the partial response. This is acceptable — a partial response without completion is misleading.
- **[Trade-off] `agent.warning` timer precision** → The 80% threshold is approximate (fires relative to wall-clock, not actual processing time). Acceptable for UX purposes.
