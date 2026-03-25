## Context

The first structured control plane rollout added a `CoordinatingExecutor` wrapper, but review found three execution-path bugs:

1. `WithOnEvent` is a setter, so the wrapper can overwrite `TurnRunner`'s trace recorder hook.
2. Circuit breaker outcomes are recorded from mutable executor-level state, which can be stale across retries or concurrent turns.
3. Non-delegated trace metrics can be attributed from today's runtime mode instead of trace/event evidence.

This follow-up keeps the v1 architecture intact. It fixes correctness bugs without expanding scope into v2 items such as authoritative external budget enforcement or per-agent direct execution.

## Goals / Non-Goals

**Goals:**
- Preserve previously installed ADK event hooks when the control-plane wrapper adds its own observer.
- Isolate mutable delegation and budget state per turn attempt so concurrent requests cannot contaminate each other.
- Record circuit breaker outcomes for the delegated specialist that actually ran during the current attempt.
- Attribute non-delegated trace metrics from trace/event evidence.

**Non-Goals:**
- Refactor `agent.go` to delegate authoritative budget or recovery control.
- Change routing ownership away from the root orchestrator.
- Add new CLI commands, config keys, or lifecycle components.

## Decisions

### Use `ChainOnEvent` instead of replacing `onEvent`

`RunOption` currently stores one `onEvent` callback. The wrapper therefore needs a helper that preserves the existing callback and appends another one. This keeps `TurnRunner` tracing intact while still letting the control-plane observe delegations.

Alternative considered:
- Replace `WithOnEvent` call sites manually.
  Rejected because every wrapper would need to reimplement chaining logic and the bug could recur.

### Keep mutable execution state in a per-run container

Delegation target and mirrored budget counters are execution state, not executor configuration. They are moved into per-run state objects created for each `RunStreamingDetailed` call.

Alternative considered:
- Protect shared executor state with mutexes only.
  Rejected because mutexes prevent data races but do not prevent one turn from resetting or overwriting another turn's state.

### Derive non-delegated metrics from event authors

Metrics attribution now uses:
1. first delegation target, if any
2. otherwise first non-empty agent author seen in trace events
3. otherwise no attribution

Alternative considered:
- Use current runtime config (`lango-agent` vs `lango-orchestrator`) as fallback.
  Rejected because historical traces can outlive config changes, producing incorrect attribution.

## Risks / Trade-offs

- [Risk] Traces with only terminal error events may have no attributable agent.
  Mitigation: skip attribution instead of guessing from transport or current config.
- [Risk] Per-run budget state clone could diverge from shared alert wiring.
  Mitigation: copy only immutable thresholds and alert handler into each run-local tracker.
- [Risk] Circuit breaker still remains observational in v1.
  Mitigation: keep specs explicit that routing authority stays with the root orchestrator.

## Migration Plan

1. Update specs for control-plane and diagnostics correctness.
2. Patch `ChainOnEvent`, run-local state, and metrics attribution.
3. Add regression tests for hook preservation, correct breaker attribution, and non-delegated metrics attribution.
4. Run `go build ./...` and `go test ./...`.

Rollback is straightforward: revert the follow-up patch set. No data migration is involved.

## Open Questions

None for v1 follow-up scope.
