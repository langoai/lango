## ADDED Requirements

### Requirement: Provider token usage capture
The system SHALL capture actual token usage data from all LLM providers (OpenAI, Anthropic, Gemini) during streaming responses. Token usage data SHALL be propagated via a `Usage` field on `StreamEvent` and forwarded to the event bus via a `TokenUsageEvent`.

#### Scenario: OpenAI token capture
- **WHEN** an OpenAI streaming response completes with `IncludeUsage: true`
- **THEN** the Done event SHALL contain `Usage` with `InputTokens`, `OutputTokens`, and `TotalTokens` from `response.Usage`

#### Scenario: Anthropic token capture
- **WHEN** an Anthropic streaming response completes
- **THEN** the Done event SHALL contain `Usage` with `InputTokens` and `OutputTokens` from `stream.Message.Usage`

#### Scenario: Gemini token capture
- **WHEN** a Gemini streaming response completes
- **THEN** the Done event SHALL contain `Usage` with `InputTokens`, `OutputTokens`, and `TotalTokens` from `resp.UsageMetadata`

#### Scenario: Backward compatibility
- **WHEN** a consumer processes a `StreamEvent` and does not access the `Usage` field
- **THEN** the `Usage` field SHALL be nil and cause no errors

### Requirement: In-memory metrics collection
The system SHALL provide a thread-safe `MetricsCollector` that aggregates token usage and tool execution metrics in memory. The collector SHALL support per-session, per-agent, and per-tool breakdowns.

#### Scenario: Record token usage
- **WHEN** a `TokenUsageEvent` is published
- **THEN** the collector SHALL update total, per-session, and per-agent token counts

#### Scenario: Record tool execution
- **WHEN** a `ToolExecutedEvent` is published
- **THEN** the collector SHALL update tool count, error count, and average duration

#### Scenario: Snapshot
- **WHEN** `Snapshot()` is called
- **THEN** a point-in-time copy of all metrics SHALL be returned without holding locks

### Requirement: Token cost estimation
The system SHALL estimate USD cost per request using a model pricing table. The calculator SHALL support prefix matching for model name variants.

#### Scenario: Known model cost
- **WHEN** `Calculate("gpt-4o", 1000, 500)` is called
- **THEN** the result SHALL be `(1000 * 2.50 + 500 * 10.00) / 1_000_000`

#### Scenario: Unknown model
- **WHEN** `Calculate("unknown-model", 1000, 500)` is called
- **THEN** the result SHALL be `0`

### Requirement: Health check system
The system SHALL provide a `HealthRegistry` that aggregates health checks from multiple components. The overall status SHALL be the worst status among all components.

#### Scenario: All healthy
- **WHEN** all registered health checkers return `healthy`
- **THEN** `CheckAll` SHALL return overall status `healthy`

#### Scenario: One unhealthy
- **WHEN** any registered health checker returns `unhealthy`
- **THEN** `CheckAll` SHALL return overall status `unhealthy`

### Requirement: Persistent token storage
The system SHALL optionally persist token usage records to an Ent `token_usage` table when `observability.tokens.persistHistory` is true. Records SHALL support retention-based cleanup.

#### Scenario: Save and query
- **WHEN** a token usage record is saved with `persistHistory: true`
- **THEN** the record SHALL be queryable by session, agent, or time range

#### Scenario: Retention cleanup
- **WHEN** `Cleanup(retentionDays)` is called
- **THEN** records older than `retentionDays` SHALL be deleted

### Requirement: Tool execution duration tracking
The system SHALL accurately measure tool execution duration by timing between pre and post hooks. The `ToolExecutedEvent.Duration` field SHALL reflect actual execution time.

#### Scenario: Duration measurement
- **WHEN** a tool executes via the hook chain
- **THEN** `ToolExecutedEvent.Duration` SHALL be the elapsed time between `Pre()` and `Post()` calls

### Requirement: CLI metrics commands
The system SHALL provide `lango metrics` CLI commands that display system metrics by querying the gateway API. Commands SHALL support `--output json|table` format flag.

#### Scenario: Summary command
- **WHEN** `lango metrics` is executed
- **THEN** a system snapshot summary SHALL be displayed with uptime, token totals, and cost

#### Scenario: JSON output
- **WHEN** `lango metrics --output json` is executed
- **THEN** the output SHALL be valid JSON

### Requirement: Gateway metrics API
The system SHALL expose metrics via HTTP endpoints on the gateway: `/metrics`, `/metrics/sessions`, `/metrics/tools`, `/metrics/agents`, `/metrics/cost`, `/metrics/history`, `/health/detailed`.

#### Scenario: Metrics endpoint
- **WHEN** `GET /metrics` is requested
- **THEN** a JSON response SHALL be returned with uptime, token usage totals, and execution counts

#### Scenario: History endpoint
- **WHEN** `GET /metrics/history?days=7` is requested with persistent storage enabled
- **THEN** historical token usage records from the last 7 days SHALL be returned

### Requirement: Audit recording
The system SHALL optionally record tool calls and token usage events to the existing `AuditLog` Ent schema when `observability.audit.enabled` is true.

#### Scenario: Tool call audit
- **WHEN** a tool is executed and audit is enabled
- **THEN** an `AuditLog` entry SHALL be created with action `tool_call`, tool name, duration, and success status

### Requirement: Observability configuration
The system SHALL support configuration under `observability:` with nested `tokens`, `health`, `audit`, and `metrics` sections. Each subsection SHALL have an `enabled` boolean.

#### Scenario: Config gating
- **WHEN** `observability.enabled` is false
- **THEN** no observability components SHALL be initialized
