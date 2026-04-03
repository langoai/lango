## ADDED Requirements

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


## ADDED Requirements

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

### Requirement: Prometheus metrics endpoint
When `observability.metrics.format` is `"prometheus"`, the system MUST register a `/metrics/prometheus` HTTP endpoint serving metrics in Prometheus text exposition format. The existing `/metrics` JSON endpoint MUST remain unchanged.

#### Scenario: Prometheus format enabled
- **WHEN** `observability.metrics.format` is `"prometheus"`
- **THEN** `/metrics/prometheus` SHALL serve `promhttp.Handler()` output
- **AND** `/metrics` SHALL continue serving JSON

#### Scenario: Prometheus format disabled
- **WHEN** `observability.metrics.format` is `"json"` or empty
- **THEN** `/metrics/prometheus` SHALL NOT be registered

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
