## Context

The exec-safety-hardening feature branch introduced several new modules across `internal/agentrt`, `internal/observability`, and `internal/app`. Post-review identified six behavior-preserving code quality issues that should be cleaned up before merging. All changes are safe refactorings with no functional impact.

## Goals / Non-Goals

**Goals:**
- Eliminate `goto` in recovery logic for readability
- Remove mutable package-level map (`defaultRetryLimits`) in favor of a pure function
- Fix timer leak in backoff sleep by using `time.NewTimer` + `defer timer.Stop()`
- Add exhaustive switch case for `"allow"` verdict in metrics collector
- Prevent potential data race on `configMetadata` map by copying before closure capture
- Remove dead code (manual map key sorting superseded by `json.Marshal` Go 1.12+)
- Replace over-broad `Snapshot()` with targeted `SessionMetrics()` query

**Non-Goals:**
- Changing any observable behavior or test expectations
- Adding new features or capabilities
- Modifying public API surfaces

## Decisions

### 1. Recovery `goto` removal — early-return guard with computed effective limit

Compute the effective retry limit (class-specific or global) before the exhaustion check. This removes the `goto` by making the check a single linear flow: compute limit, check exhaustion, proceed with class-level logic.

Alternative: nested if/else — rejected because it increases nesting depth without improving clarity.

### 2. `defaultRetryLimits` map to switch statement

Replace the mutable `map[CauseClass]int` with a `switch` in `retryLimitForClass`. The map is never modified at runtime, but a `switch` is inherently immutable, zero-allocation, and easier to reason about in concurrent code.

### 3. `sleepWithContext` helper with `time.NewTimer`

Extract a `sleepWithContext(ctx, d)` helper that uses `time.NewTimer` + `defer timer.Stop()` instead of `time.After`. `time.After` leaks the timer until it fires; in retry loops with context cancellation, this can accumulate leaked timers.

### 4. Map copy before closure capture

In `buildProvenanceAgentOptions`, `cachedMetadata` is captured by the `rootObserver` closure. The map is later mutated (hook_registry key updated). Copy the map before closure capture to prevent a data race between the mutator and the closure consumer.

### 5. Remove manual key sorting in `computeConfigFingerprint`

`json.Marshal` on `map[string]bool` produces deterministic key ordering in Go 1.12+. The manual sort-and-rebuild is dead code.

### 6. Targeted `SessionMetrics()` instead of `Snapshot()`

`wireSessionUsage` calls `collector.Snapshot()` which deep-copies all sessions, agents, tools, and policy metrics just to read one session's counters. `SessionMetrics(sessionKey)` copies only the single session metric needed.

## Risks / Trade-offs

- [Risk] Recovery logic restructuring could subtly change retry behavior → Mitigation: existing tests cover all branches; no test expectation changes allowed.
- [Risk] `sleepWithContext` changes timer semantics → Mitigation: behavior is identical (select on ctx.Done vs timer); only resource cleanup differs.
