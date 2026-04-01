## Why

The observability layer captures policy decisions, recovery events, and budget signals, but there is no system that monitors these signals for anomalies and generates operational alerts. Operators must manually query metrics or audit logs to detect issues like excessive policy blocks, recovery retries, or circuit breaker trips. An alerting system closes this gap by providing threshold-based monitoring with deduplication and delivery through the existing EventBus/audit/CLI infrastructure.

## What Changes

- New `AlertEvent` event type published through EventBus when alert conditions are met
- New alerting dispatcher that monitors policy/recovery/budget events using sliding window thresholds
- Audit recorder extended to persist alert events (action="alert") to the audit log
- New `/alerts` HTTP endpoint to query persisted alerts from the audit database
- New `lango alerts list` and `lango alerts summary` CLI commands
- New `AlertingConfig` in the config system with threshold defaults

## Capabilities

### New Capabilities
- `alerting`: Threshold-based operational alerting system that monitors policy/recovery/budget signals and generates alerts through EventBus, audit persistence, and CLI

### Modified Capabilities
- `observability`: Extended audit recorder to handle AlertEvent and new `/alerts` route
- `eventbus`: New AlertEvent struct and event name constant

## Impact

- **Code**: `internal/alerting/` (new), `internal/eventbus/events.go`, `internal/observability/audit/recorder.go`, `internal/app/wiring_observability.go`, `internal/app/routes_observability.go`, `internal/config/types.go`, `internal/config/loader.go`, `internal/cli/alerts/` (new), `cmd/lango/main.go`
- **APIs**: New `/alerts` HTTP endpoint
- **Dependencies**: No new external dependencies — uses existing EventBus, Ent audit schema, and chi router
- **Ent schema**: `internal/ent/schema/audit_log.go` gains `"alert"` in action enum (requires `go generate` post-flight, not during this change)
