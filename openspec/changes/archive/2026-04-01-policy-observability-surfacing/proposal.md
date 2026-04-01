## Why

PolicyDecisionEvent is published to the event bus when the exec policy evaluator blocks or observes a command, but no subsystem currently subscribes to it. Operators have no way to inspect policy decisions through audit logs, metrics, or turn traces. This makes it impossible to diagnose false positives, tune policy rules, or detect policy-evading patterns.

## What Changes

- Subscribe to `PolicyDecisionEvent` in the audit recorder and persist decisions as audit log entries with action `policy_decision`
- Add `EventPolicyDecision` constant to turntrace events for trace timeline integration
- Add policy decision counters and reason-code aggregates to the metrics collector
- Add `PolicyMetrics` to the `SystemSnapshot` type for snapshot queries
- Add `/metrics/policy` HTTP endpoint for operator access
- Add `lango metrics policy` CLI subcommand for table/JSON rendering
- Wire `PolicyDecisionEvent` subscription in `initObservability` to feed the metrics collector

## Capabilities

### New Capabilities
- `policy-observability`: Subscribe to PolicyDecisionEvent across audit, metrics, and turntrace subsystems to make policy decisions observable via logs, counters, and CLI

### Modified Capabilities
- `observability`: Add PolicyMetrics to SystemSnapshot and RecordPolicyDecision to MetricsCollector

## Impact

- `internal/ent/schema/audit_log.go` — new enum value `policy_decision`
- `internal/observability/audit/recorder.go` — new handler and subscription
- `internal/turntrace/events.go` — new constant
- `internal/observability/collector.go` — new counters and method
- `internal/observability/types.go` — new struct in SystemSnapshot
- `internal/app/wiring_observability.go` — new event subscription
- `internal/app/routes_observability.go` — new HTTP endpoint
- `internal/cli/metrics/policy.go` — new CLI subcommand (new file)
- `internal/cli/metrics/metrics.go` — register new subcommand
- No breaking changes, no dependency additions
