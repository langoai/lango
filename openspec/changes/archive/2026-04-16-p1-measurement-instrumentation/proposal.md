## Why

P1 units (shared task coordination, child session reset policy, workspace cleanup telemetry) require data before implementation decisions can be made. Currently the relevant code paths either silently discard errors, emit events that nobody aggregates, or lack timing instrumentation. This change adds minimal logging/metrics at three points so data accumulates naturally during normal operation. No behavioral changes — observe only.

## What Changes

- **Unit 4 prep (team task duplication)**: Add a metrics observer that subscribes to `TeamTaskDelegatedEvent` and `TeamTaskCompletedEvent` via EventBus. Log worker count, success/fail ratio, and avg duration per delegation. Accumulate in-memory counters exposed via existing health/diagnostics.
- **Unit 5 prep (child session lifecycle)**: Extend the existing `childHook` in `wiring.go:703` to log fork/merge/discard with timestamps and session keys at Info level (currently Debug only for errors). This enables post-hoc analysis of session lifetimes from logs.
- **Unit 9 prep (workspace cleanup)**: Replace `_ =` error swallowing in `workspace.go:107-108` cleanup function with `log.Warnw` so cleanup failures become visible in logs.

## Capabilities

### New Capabilities
_(none)_

### Modified Capabilities
_(no spec-level behavior changes — logging only)_

## Impact

- `internal/app/wiring.go` — childHook logging enhancement (3 lines)
- `internal/runledger/workspace.go` — cleanup error logging (2 lines)
- `internal/app/` — new small observer file for team task metrics bridge
