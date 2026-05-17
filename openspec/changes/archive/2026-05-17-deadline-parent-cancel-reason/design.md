## Context

`deadline.New(parent, idleTimeout, maxTimeout)` creates a child context with `context.WithCancel(parent)`. The child context is cancelled automatically when the parent is cancelled, but the `ExtendableDeadline` state machine does not currently observe that parent cancellation to update `reason`, `done`, or timers.

## Goals

- Preserve existing idle and max-timeout behavior.
- Ensure parent cancellation is classified as `ReasonCancelled`.
- Stop timers once parent cancellation wins to avoid later callbacks mutating state.
- Keep the `internal/app` alias path behaviorally equivalent.

## Non-Goals

- Do not introduce fake clocks or a broader timer abstraction in this small bug fix.
- Do not change timeout resolution or gateway/channel wiring.

## Design

Add a parent watcher goroutine inside `deadline.New`:

- Wait on `parent.Done()` and the child context's `Done()` channel.
- If `parent.Done()` fires first, acquire the deadline mutex, mark `done`, set `reason=ReasonCancelled`, stop both timers, and call the child cancel function.
- If the child context is already done due to idle, max timeout, or `Stop()`, exit without changing the reason.

This keeps parent cancellation as a first-writer-wins state transition alongside the existing timer callbacks and `Stop()`.

## Risks

- A watcher goroutine per deadline is acceptable because each deadline already owns timers and exits when the child context is done.
- Races around simultaneous parent cancellation and timer expiry remain first-writer-wins; tests should assert parent cancellation when it clearly happens before timer deadlines.
