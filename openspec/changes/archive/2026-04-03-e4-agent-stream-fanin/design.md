## Context

E1 delivered stream combinators (`Stream[T]`, `Tag[T]`, `Merge`, `Race`, `FanIn`, `Drain`) in `internal/streamx/`. E2 added `ProgressBus` with `Emit`/`Subscribe` pub/sub. D1 established `AgentRun` model. The multi-agent runtime needs a way to merge multiple child agent output streams into a single tagged stream while emitting lifecycle progress events for observability.

## Goals / Non-Goals

**Goals:**
- Compose existing `Merge` combinator with `ProgressBus` for child agent stream fan-in
- Emit per-child lifecycle events (started, completed, failed) via `ProgressBus`
- Support nil bus for contexts where progress tracking is unnecessary
- Keep the API surface minimal: `NewAgentStreamFanIn`, `AddChild`, `MergedStream`

**Non-Goals:**
- Backpressure or rate-limiting on child streams (handled at stream level)
- Retry or reconnection logic for failed children
- Child agent lifecycle management (start/stop) -- this only observes output streams
- Custom event types beyond string (agent text output)

## Decisions

**1. Wrap child streams instead of post-processing merged output**

Each child stream is wrapped before being passed to `Merge`. The wrapper detects stream completion (normal end or error) and emits the appropriate progress event. This preserves per-child identity, since `Merge` error propagation loses source info in error tags.

Alternative considered: Post-process the merged stream to detect per-child completion. Rejected because `Merge` does not tag errors with their source, making it impossible to attribute errors to specific children.

**2. `Stream[string]` for agent text output**

Agent output is modeled as `Stream[string]` -- the simplest representation for text chunks. This avoids premature abstraction while remaining compatible with future typed output via `Stream[AgentOutput]`.

**3. Progress source format: `agent:{parent}:child:{childID}`**

Hierarchical source naming enables prefix-based filtering. Subscribers can filter by parent (`agent:session-1:`) or specific child (`agent:session-1:child:alpha`).

**4. Nil bus is safe (no-op pattern)**

All emit methods check `f.bus == nil` before emitting. This avoids requiring a bus in contexts where progress tracking is not needed, without callers needing sentinel values.

## Risks / Trade-offs

- [Risk] `ProgressStarted` is emitted for all children before any stream events arrive, creating a slight ordering inaccuracy if a child stream produces output faster than the bus subscriber processes the started event. → Mitigation: ProgressBus uses buffered channels (cap 64); in practice, started events will be consumed before stream events complete.
- [Risk] If a consumer breaks out of the merged stream early, `wrapChild` emits `ProgressCompleted` for the interrupted child, which may be misleading. → Mitigation: Acceptable for v1; future versions can distinguish "cancelled" from "completed".
- [Trade-off] `AddChild` is not goroutine-safe. Children must be registered before calling `MergedStream`. → Acceptable: registration happens during orchestration setup, not concurrently.
