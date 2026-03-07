## Purpose

Observability system for tracking token usage, tool execution, health checks, and audit logging across all LLM providers.

## Requirements

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
The system SHALL provide a thread-safe `MetricsCollector` that aggregates token usage and tool execution metrics in memory. The collector SHALL support per-session, per-agent, and per-tool breakdowns. The collector SHALL NOT track estimated costs.

#### Scenario: Record token usage
- **WHEN** a `TokenUsageEvent` is published
- **THEN** the collector SHALL update total, per-session, and per-agent token counts

#### Scenario: Record tool execution
- **WHEN** a `ToolExecutedEvent` is published
- **THEN** the collector SHALL update tool count, error count, and average duration

#### Scenario: Snapshot
- **WHEN** `Snapshot()` is called
- **THEN** a point-in-time copy of all metrics SHALL be returned without holding locks

#### Scenario: Token usage types exclude cost
- **WHEN** `TokenUsage`, `AgentMetric`, `SessionMetric`, or `TokenUsageSummary` types are used
- **THEN** they SHALL NOT contain an `EstimatedCost` field

### Requirement: Health check system
The system SHALL provide a `HealthRegistry` that aggregates health checks from multiple components. The overall status SHALL be the worst status among all components.

#### Scenario: All healthy
- **WHEN** all registered health checkers return `healthy`
- **THEN** `CheckAll` SHALL return overall status `healthy`

#### Scenario: One unhealthy
- **WHEN** any registered health checker returns `unhealthy`
- **THEN** `CheckAll` SHALL return overall status `unhealthy`

### Requirement: Persistent token storage
The system SHALL persist token usage records via Ent without an `estimated_cost` column. Records SHALL support retention-based cleanup.

#### Scenario: Save and query
- **WHEN** a token usage record is saved with `persistHistory: true`
- **THEN** the record SHALL be queryable by session, agent, or time range

#### Scenario: Retention cleanup
- **WHEN** `Cleanup(retentionDays)` is called
- **THEN** records older than `retentionDays` SHALL be deleted

#### Scenario: Save token usage
- **WHEN** a token usage record is saved
- **THEN** the record SHALL include provider, model, session key, agent name, input/output/total/cache tokens, and timestamp, and SHALL NOT include estimated cost

#### Scenario: Aggregate results
- **WHEN** aggregate stats are computed
- **THEN** the result SHALL include total input, total output, total tokens, and record count, and SHALL NOT include total cost

### Requirement: Tool execution duration tracking
The system SHALL accurately measure tool execution duration by timing between pre and post hooks. The `ToolExecutedEvent.Duration` field SHALL reflect actual execution time.

#### Scenario: Duration measurement
- **WHEN** a tool executes via the hook chain
- **THEN** `ToolExecutedEvent.Duration` SHALL be the elapsed time between `Pre()` and `Post()` calls

### Requirement: CLI metrics commands
The system SHALL provide `lango metrics` CLI commands that display system metrics by querying the gateway API. Commands SHALL support `--output json|table` format flag. The CLI SHALL NOT display cost columns or cost subcommands.

#### Scenario: Summary command
- **WHEN** `lango metrics` is executed
- **THEN** the output SHALL display uptime, total input tokens, total output tokens, and tool executions, and SHALL NOT display estimated cost

#### Scenario: JSON output
- **WHEN** `lango metrics --output json` is executed
- **THEN** the output SHALL be valid JSON

#### Scenario: Cost subcommand removed
- **WHEN** `lango metrics cost` is run
- **THEN** the command SHALL NOT be recognized

#### Scenario: Sessions table
- **WHEN** `lango metrics sessions` is run
- **THEN** the table SHALL include SESSION, INPUT, OUTPUT, TOTAL, REQUESTS columns and SHALL NOT include a COST column

#### Scenario: Agents table
- **WHEN** `lango metrics agents` is run
- **THEN** the table SHALL include AGENT, INPUT, OUTPUT, TOOL CALLS columns and SHALL NOT include a COST column

#### Scenario: History table
- **WHEN** `lango metrics history` is run
- **THEN** the table SHALL include TIME, PROVIDER, MODEL, INPUT, OUTPUT columns and SHALL NOT include a COST column

### Requirement: Gateway metrics API
The system SHALL expose metrics via HTTP endpoints on the gateway: `/metrics`, `/metrics/sessions`, `/metrics/tools`, `/metrics/agents`, `/metrics/history`, `/health/detailed`. The API SHALL NOT include cost estimation fields or a `/metrics/cost` endpoint.

#### Scenario: Metrics endpoint
- **WHEN** `GET /metrics` is requested
- **THEN** a JSON response SHALL be returned with uptime, token usage totals (without cost), and execution counts

#### Scenario: Sessions endpoint
- **WHEN** `GET /metrics/sessions` is called
- **THEN** each session object SHALL include token counts and request count, and SHALL NOT include `estimatedCost`

#### Scenario: Agents endpoint
- **WHEN** `GET /metrics/agents` is called
- **THEN** each agent object SHALL include token counts and tool calls, and SHALL NOT include `estimatedCost`

#### Scenario: History endpoint
- **WHEN** `GET /metrics/history?days=7` is requested with persistent storage enabled
- **THEN** historical token usage records from the last 7 days SHALL be returned without cost fields

#### Scenario: Cost endpoint removed
- **WHEN** `GET /metrics/cost` is called
- **THEN** the server SHALL return 404

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
