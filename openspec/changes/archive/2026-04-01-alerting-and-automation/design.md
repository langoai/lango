## Context

The Lango observability layer currently captures policy decisions, token usage, and tool executions via EventBus subscriptions and persists them to an audit log. However, no automated system monitors these signals for anomalies. Operators must manually query metrics or audit logs to discover issues. The existing infrastructure (EventBus, audit recorder, chi router, cobra CLI) provides all building blocks needed for an alerting system.

## Goals / Non-Goals

**Goals:**
- Detect operational anomalies (excessive policy blocks, recovery retries, circuit breaker trips) automatically
- Publish alerts through the existing EventBus as AlertEvent
- Persist alerts to the audit log for querying
- Expose alerts via HTTP `/alerts` endpoint and CLI `lango alerts` commands
- Support configurable thresholds with sensible defaults (disabled by default)

**Non-Goals:**
- External notification systems (email, PagerDuty, webhooks) — future work
- Real-time push notifications via WebSocket
- Alert acknowledgment or silence workflows
- Complex alert routing rules or escalation policies

## Decisions

### 1. Alert delivery: 3-tier EventBus → Audit → CLI

Alerts flow through the existing EventBus as `AlertEvent`, get persisted by the audit recorder (action="alert"), and are queryable via `/alerts` endpoint + CLI. This reuses existing infrastructure without new dependencies.

**Alternative considered**: Dedicated alert store with separate schema. Rejected because the audit log already has the right structure (action enum, session key, details JSON, timestamps) and adding a separate store adds maintenance overhead for no benefit.

### 2. Sliding window threshold detection in dispatcher

The alerting dispatcher subscribes to `PolicyDecisionEvent` and maintains in-memory sliding windows (5-minute intervals) to detect threshold breaches. When a threshold is exceeded, it publishes an `AlertEvent` and deduplicates to avoid alert storms (one alert per type per window).

**Alternative considered**: Database-backed window counting. Rejected because in-memory windows are simpler, faster, and alerts are ephemeral signals — the persisted audit log provides the durable record.

### 3. Config-driven with disabled-by-default

`AlertingConfig` is added to the root Config struct, defaulting to `Enabled: false`. Thresholds are configurable via `mapstructure` tags. This follows the pattern used by observability, sandbox, and provenance configs.

### 4. Audit action as string literal until ent regeneration

The ent schema gains `"alert"` in the action enum, but `go generate` is not run during this change. The recorder uses `auditlog.Action("alert")` as a string cast, matching the pattern used for `policy_decision` during its initial implementation. Post-flight `go generate` makes it a proper enum constant.

## Risks / Trade-offs

- **[Risk] In-memory window lost on restart** → Acceptable for alerting purposes; audit log retains history. Windows rebuild naturally from incoming events.
- **[Risk] Alert storm if thresholds too low** → Deduplication (one alert per type per 5min window) prevents flooding. Config allows tuning thresholds.
- **[Risk] Ent enum not generated** → Using `auditlog.Action("alert")` cast until `go generate` runs in post-flight. Build compiles and tests pass without generation.
