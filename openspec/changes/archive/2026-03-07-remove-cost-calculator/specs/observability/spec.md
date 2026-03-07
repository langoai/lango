## REMOVED Requirements

### Requirement: Token cost estimation
**Reason**: No LLM provider offers a model pricing API. Hardcoded price tables become inaccurate on every model release, and inaccurate cost estimates are worse than no estimates.
**Migration**: Use token counts directly. Cost estimation should be done externally by users who can reference current provider pricing pages.

## MODIFIED Requirements

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

### Requirement: Observability HTTP API
The system SHALL expose token usage and tool execution metrics via HTTP endpoints. The API SHALL NOT include cost estimation fields.

#### Scenario: Metrics summary endpoint
- **WHEN** `GET /metrics` is called
- **THEN** the response SHALL include `tokenUsage` with `inputTokens`, `outputTokens`, `totalTokens`, `cacheTokens` and SHALL NOT include `estimatedCost`

#### Scenario: Sessions endpoint
- **WHEN** `GET /metrics/sessions` is called
- **THEN** each session object SHALL include token counts and request count, and SHALL NOT include `estimatedCost`

#### Scenario: Agents endpoint
- **WHEN** `GET /metrics/agents` is called
- **THEN** each agent object SHALL include token counts and tool calls, and SHALL NOT include `estimatedCost`

#### Scenario: History endpoint
- **WHEN** `GET /metrics/history` is called
- **THEN** each record SHALL include provider, model, token counts, and timestamp, and SHALL NOT include `estimatedCost`

#### Scenario: Cost endpoint removed
- **WHEN** `GET /metrics/cost` is called
- **THEN** the server SHALL return 404

### Requirement: CLI metrics commands
The system SHALL provide CLI commands for viewing token usage metrics. The CLI SHALL NOT display cost columns or cost subcommands.

#### Scenario: Metrics summary
- **WHEN** `lango metrics` is run
- **THEN** the output SHALL display uptime, total input tokens, total output tokens, and tool executions, and SHALL NOT display estimated cost

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

### Requirement: Persistent token storage
The system SHALL persist token usage records via Ent without an `estimated_cost` column.

#### Scenario: Save token usage
- **WHEN** a token usage record is saved
- **THEN** the record SHALL include provider, model, session key, agent name, input/output/total/cache tokens, and timestamp, and SHALL NOT include estimated cost

#### Scenario: Aggregate results
- **WHEN** aggregate stats are computed
- **THEN** the result SHALL include total input, total output, total tokens, and record count, and SHALL NOT include total cost
