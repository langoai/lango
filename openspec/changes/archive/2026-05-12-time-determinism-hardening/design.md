## Context

The repository baseline failed in three places before any broader production-readiness work could continue. Two failures came from tests that compared TTL- and age-based behavior against historical fixed dates while production code used the real current clock. The third failure came from `ParallelReadOnlyExecutor`, where a completed failing handler could return a measured duration of zero on fast clocks.

## Goals / Non-Goals

**Goals:**
- Remove calendar-time coupling from clock-sensitive regression tests.
- Keep Mission Control loop threshold assertions deterministic by explicitly pinning the projector clock in tests that depend on age windows.
- Ensure completed parallel read-only tool invocations always expose a positive duration, even on error paths.
- Restore a green `go build ./...` and `go test ./...` baseline.

**Non-Goals:**
- Redesign Mission Control loop ordering or proposal TTL semantics.
- Change public CLI, TUI, or API behavior.
- Expand this slice into broader architecture or UX refactoring.

## Decisions

### D1: Use relative or injected clocks in regression tests
Tests that assert TTL or age-window behavior now derive timestamps from `time.Now()` or inject `projector.nowFn` explicitly. This keeps the behavior under test the same while removing accidental dependence on the wall-clock date when the suite runs.

### D2: Keep proposal lifecycle behavior unchanged
The proposal module test now uses current-time-relative fixtures instead of historical dates. This avoids turning a test fragility cleanup into a proposal semantics change.

### D3: Clamp completed durations to a positive minimum
`ParallelReadOnlyExecutor` now normalizes any non-positive measured duration to `time.Nanosecond` after handler completion. This preserves ordering and observability semantics while guaranteeing that completed results never look unmeasured.

## Risks / Trade-offs

- **[Synthetic minimum duration]** → A sub-nanosecond or zero raw measurement is normalized to `1ns`. Mitigation: this only affects the degenerate case and is preferable to reporting a completed invocation as having no duration at all.
- **[Test-only hardening can hide product issues]** → Clock injection could be misused to paper over runtime bugs. Mitigation: this slice only hardens tests whose production behavior was already correct and leaves runtime logic unchanged except for duration normalization.
