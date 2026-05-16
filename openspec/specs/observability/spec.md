## Purpose

Capability spec for observability. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Audit recorder handles AlertEvent
The audit recorder SHALL subscribe to AlertEvent via SubscribeTyped and persist each alert to the audit log with action="alert", actor="system", target=alert type, and details containing severity, message, and alert-specific metadata.

#### Scenario: AlertEvent persisted to audit log
- **WHEN** an AlertEvent is published on the EventBus
- **THEN** the audit recorder creates an audit log entry with action="alert", actor="system", target set to the alert type, and details containing severity and message

### Requirement: Alerts HTTP route registered
The `/alerts` HTTP route SHALL be registered alongside existing observability routes when the observability system is enabled and a database client is available.

#### Scenario: Alerts endpoint available
- **WHEN** observability is enabled and the application starts
- **THEN** the GET `/alerts` endpoint is registered on the chi router



### Requirement: Session map capacity limit
The `MetricsCollector` MUST support a `MaxSessions` field (default: 10,000) that caps the number of tracked sessions. When the cap is reached and a new session is inserted, the least-recently-updated session MUST be evicted.

#### Scenario: Eviction at capacity
- **WHEN** `MaxSessions` is 10,000 and the 10,001st session records token usage
- **THEN** the oldest session (by `LastUpdated`) is removed before the new session is inserted

#### Scenario: Eviction selects oldest
- **GIVEN** sessions A (updated 1 min ago) and B (updated 5 min ago) at capacity
- **WHEN** a new session C records usage
- **THEN** session B is evicted (oldest `LastUpdated`)

#### Scenario: Cap disabled
- **WHEN** `MaxSessions` is 0 or negative
- **THEN** no eviction occurs and sessions grow unbounded (backward compatible)

### Requirement: Session metric timestamp
Each `SessionMetric` MUST include a `LastUpdated time.Time` field that is set to `time.Now()` on every `RecordTokenUsage` call for that session.

#### Scenario: RecordTokenUsage refreshes LastUpdated
- **WHEN** token usage is recorded for an existing or new session
- **THEN** that session metric MUST update `LastUpdated` to the current time

### Requirement: Prometheus metrics endpoint
When `observability.metrics.format` is `"prometheus"`, the system MUST register a `/metrics/prometheus` HTTP endpoint serving metrics in Prometheus text exposition format. The existing `/metrics` JSON endpoint MUST remain unchanged.

#### Scenario: Prometheus format enabled
- **WHEN** `observability.metrics.format` is `"prometheus"`
- **THEN** `/metrics/prometheus` SHALL serve `promhttp.Handler()` output
- **AND** `/metrics` SHALL continue serving JSON

#### Scenario: Prometheus format disabled
- **WHEN** `observability.metrics.format` is `"json"` or empty
- **THEN** `/metrics/prometheus` SHALL NOT be registered

### Requirement: CLI system metrics summary
The CLI SHALL provide `lango metrics` that fetches the `/metrics` JSON endpoint and renders a system metrics snapshot summary as table (default) or JSON.

#### Scenario: Metrics summary table output
- **WHEN** `lango metrics` is run without --output flag
- **THEN** it SHALL display uptime, total input/output token counts, and tool execution count

#### Scenario: Metrics summary JSON output
- **WHEN** `lango metrics --output json` is run
- **THEN** it SHALL output raw JSON from the `/metrics` endpoint

#### Scenario: Metrics summary rejects an unknown output format before fetch
- **WHEN** `lango metrics --output yaml` is run
- **THEN** it SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT contact the gateway

### Requirement: CLI system metrics output routing
`lango metrics` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Metrics summary table writes to command output
- **WHEN** `lango metrics` is run without --output flag
- **THEN** the command writes the table summary to the Cobra command output stream

#### Scenario: Metrics summary JSON writes to command output
- **WHEN** `lango metrics --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream

### Requirement: CLI metrics breakdown subcommands
The CLI SHALL provide `lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history` that fetch their respective `/metrics/*` endpoints and render table (default) or JSON output.

#### Scenario: Sessions output
- **WHEN** `lango metrics sessions` is run
- **THEN** it SHALL display per-session token usage or an empty-state message when no session data exists

#### Scenario: Tools output
- **WHEN** `lango metrics tools` is run
- **THEN** it SHALL display per-tool execution statistics or an empty-state message when no tool data exists

#### Scenario: Agents output
- **WHEN** `lango metrics agents` is run
- **THEN** it SHALL display per-agent token usage or an empty-state message when no agent data exists

#### Scenario: History output
- **WHEN** `lango metrics history --days 7` is run
- **THEN** it SHALL display historical token usage records and aggregate totals or an empty-state message when no history exists

### Requirement: CLI metrics breakdown output routing
`lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Sessions JSON writes to command output
- **WHEN** `lango metrics sessions --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Tools empty-state writes to command output
- **WHEN** `lango metrics tools` returns no data
- **THEN** the command writes the empty-state message to the Cobra command output stream

#### Scenario: Agents empty-state writes to command output
- **WHEN** `lango metrics agents` returns no data
- **THEN** the command writes the empty-state message to the Cobra command output stream

#### Scenario: History table writes to command output
- **WHEN** `lango metrics history --days 3` is run
- **THEN** the command writes the history summary and table to the Cobra command output stream

### Requirement: Prometheus metric instruments
The `PrometheusExporter` MUST register: `lango_token_usage_total` (counter, labels: type), `lango_tool_executions_total` (counter, labels: tool, success), `lango_tool_duration_seconds` (histogram, labels: tool), `lango_policy_decisions_total` (counter, labels: verdict), `lango_tracked_sessions` (gauge). All counters MUST be updated via EventBus event subscriptions.

#### Scenario: Tool execution recorded
- **WHEN** a `ToolExecutedEvent` is published
- **THEN** `lango_tool_executions_total` SHALL increment with the tool name and success label

#### Scenario: Tracked sessions updated
- **WHEN** a `TokenUsageEvent` is published
- **THEN** `lango_tracked_sessions` gauge SHALL reflect the current collector session count

### Requirement: OpenTelemetry tracing
When `observability.tracing.enabled` is true, the system MUST initialize an OpenTelemetry `TracerProvider` with the configured exporter (`"stdout"` or `"none"`). The provider MUST be shut down during `App.Stop()` to flush pending spans.

#### Scenario: Stdout exporter
- **WHEN** `observability.tracing.exporter` is `"stdout"`
- **THEN** spans SHALL be written to stderr in OTLP JSON format

#### Scenario: Tracer shutdown flushes spans
- **WHEN** `App.Stop()` is called
- **THEN** `TracerProvider.Shutdown()` SHALL be called to flush batched spans

#### Scenario: Unsupported exporter rejected
- **WHEN** `observability.tracing.exporter` is an unknown value
- **THEN** `InitTracer` SHALL return an error

### Requirement: Alerts route uses storage alert reader
The observability alerts route MUST query alert history through a storage-provided alert reader instead of a raw ent client.

#### Scenario: Alerts endpoint reads through storage facade
- **WHEN** the `/alerts` route is requested
- **THEN** the route queries alert records through the storage facade alert reader
- **AND** it does not issue ad hoc ent queries from the route layer
