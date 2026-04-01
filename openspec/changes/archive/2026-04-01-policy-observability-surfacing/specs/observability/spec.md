## MODIFIED Requirements

### Requirement: In-memory metrics collection
The system SHALL provide a thread-safe `MetricsCollector` that aggregates token usage, tool execution, and policy decision metrics in memory. The collector SHALL support per-session, per-agent, per-tool, and per-reason-code breakdowns. The collector SHALL NOT track estimated costs.

#### Scenario: Record token usage
- **WHEN** a `TokenUsageEvent` is published
- **THEN** the collector SHALL update total, per-session, and per-agent token counts

#### Scenario: Record tool execution
- **WHEN** a `ToolExecutedEvent` is published
- **THEN** the collector SHALL update tool count, error count, and average duration

#### Scenario: Record policy decision
- **WHEN** `RecordPolicyDecision(verdict, reason)` is called with verdict="block"
- **THEN** the collector SHALL increment the policy blocks counter and the byReason counter for the given reason

#### Scenario: Snapshot
- **WHEN** `Snapshot()` is called
- **THEN** a point-in-time copy of all metrics SHALL be returned without holding locks
- **AND** the snapshot SHALL include PolicyMetrics with Blocks, Observes, and ByReason fields

#### Scenario: Token usage types exclude cost
- **WHEN** `TokenUsage`, `AgentMetric`, `SessionMetric`, or `TokenUsageSummary` types are used
- **THEN** they SHALL NOT contain an `EstimatedCost` field
