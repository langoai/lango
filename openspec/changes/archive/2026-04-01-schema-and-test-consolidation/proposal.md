## Why

Post-review quality findings from the exec-safety-hardening feature branch identified several behavior-preserving code quality issues: a `goto` statement in recovery logic, leaked timers in backoff sleep, missing exhaustive switch case, mutable map sharing across goroutine boundaries, unnecessary manual map key sorting, and an over-broad metrics snapshot where a targeted query suffices. These are safe refactorings that improve maintainability and correctness without changing behavior.

## What Changes

- Replace `goto classCheck` in `recovery.go` with early-return guard pattern and convert mutable `defaultRetryLimits` map to a pure `switch` statement
- Extract `sleepWithContext` helper in `coordinating_executor.go` using `time.NewTimer` + `defer timer.Stop()` to prevent timer leaks
- Add exhaustive `"allow"` case to `RecordPolicyDecision` switch in `collector.go`
- Copy `configMetadata` map before closure capture in `wiring.go` to prevent data race
- Remove redundant manual map key sorting in `computeConfigFingerprint` since `json.Marshal` sorts map keys deterministically in Go 1.12+
- Replace full `collector.Snapshot()` with targeted `collector.SessionMetrics(sessionKey)` in `wiring_session_usage.go`

## Capabilities

### New Capabilities

(none — this is a behavior-preserving refactoring)

### Modified Capabilities

(none — no spec-level behavior changes)

## Impact

- `internal/agentrt/recovery.go` — control flow restructuring, mutable global removal
- `internal/agentrt/coordinating_executor.go` — timer leak fix, comment cleanup
- `internal/observability/collector.go` — exhaustive switch
- `internal/app/wiring.go` — map copy for goroutine safety
- `internal/app/modules_provenance.go` — dead code removal
- `internal/app/wiring_session_usage.go` — targeted metrics query
