## ADDED Requirements

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
- **AND** `lango_tool_duration_seconds` SHALL observe the duration

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
