## Context

`lango serve` starts the gateway and channel components under a shared lifecycle registry, then stops them synchronously during shutdown. The current stop path cancels the app context first, but several components still use unbounded `Wait()`-style shutdown logic or ignore the provided shutdown deadline entirely. As a result, one blocked component can stall the whole process and keep the terminal attached until the user force-kills it.

The change is cross-cutting: it touches CLI signal handling, lifecycle coordination, channel stop contracts, and automation managers. It also needs stronger observability so the next shutdown regression can be localized quickly from logs.

## Goals / Non-Goals

**Goals:**
- Bound graceful shutdown by the existing 10-second deadline for `lango serve`
- Prevent one blocked lifecycle component from hanging global shutdown forever
- Support a second interrupt as an explicit force-exit path
- Make channel and manager shutdown paths context-aware where they currently wait indefinitely
- Improve shutdown logs so stop progress and timeout points are visible

**Non-Goals:**
- Redesign every stop API in the codebase to be context-aware in one pass
- Add a new shutdown configuration knob
- Guarantee cleanup completion after the user triggers forced exit
- Change gateway request cancellation semantics beyond the already-implemented in-flight shutdown behavior

## Decisions

### 1. Enforce deadline at the lifecycle registry boundary

`lifecycle.Registry.StopAll()` will wrap each `Component.Stop(ctx)` call in a goroutine and race it against `ctx.Done()`. This prevents a single stop handler from blocking the rest of the shutdown sequence forever.

Rationale:
- The registry is the single place that already serializes global shutdown
- It gives immediate protection to legacy components that still ignore context
- It lets us keep reverse stop ordering while preserving forward progress during timeout scenarios

Alternative considered:
- Update every component first and keep `StopAll()` synchronous. Rejected because it leaves the current shutdown bug exposed until all downstream stop paths are fixed.

### 2. Make serve signal handling explicitly two-stage

The `serve` command will treat the first signal as a graceful shutdown trigger and the second as an immediate forced exit with status `130`. The exit path will go through a small helper so the behavior is unit-testable.

Rationale:
- Users need a reliable escape hatch when graceful shutdown is stuck or timed out
- This mirrors standard CLI expectations for repeated `Ctrl+C`

Alternative considered:
- Continue retrying graceful shutdown on every signal. Rejected because it preserves the current “terminal stays attached” failure mode.

### 3. Narrow context-aware API changes to serve-connected blockers

Only the internal `app.Channel`, `background.Manager`, and `workflow.Engine` shutdown contracts will be changed in this iteration. Other legacy stop paths remain protected by the lifecycle timeout wrapper.

Rationale:
- These are directly on the `lango serve` shutdown path
- They currently use indefinite wait behavior or have stop ordering issues
- The narrower change reduces churn while still fixing the user-visible bug

Alternative considered:
- Context-enable every stop path under `internal/`. Rejected as too wide for a targeted shutdown hardening change.

### 4. Fix Telegram stop ordering before waiting

Telegram shutdown will stop receiving updates before waiting on its worker goroutines, and the stop signal channel will be guarded with `sync.Once`.

Rationale:
- The current implementation can wait before interrupting the source that feeds the update loop
- Repeated stop calls must remain safe

## Risks / Trade-offs

- [Risk] A timed-out component may still continue its own cleanup in the background after the registry moves on
  - Mitigation: keep app context cancellation first, add component-level timeout logs, and expose forced exit on second interrupt
- [Risk] Context-aware stop API changes may require test fixture updates across channels and automation managers
  - Mitigation: update internal interfaces only and keep changes localized to serve-connected call sites
- [Risk] Force exit can skip non-critical cleanup work
  - Mitigation: only trigger forced exit on the second interrupt, after a graceful shutdown attempt has already started
- [Trade-off] Registry timeout wrapping can produce overlapping stop execution if a component ignores cancellation
  - Accepted: bounded process exit is more important than strict stop serialization when the component is already non-cooperative
