## ADDED Requirements

### Requirement: Typed event constants
The system SHALL define `type EventType = string` (alias for backward compatibility) and typed constants for all trace event types: `EventToolCall`, `EventToolResult`, `EventDelegation`, `EventDelegationReturn`, `EventText`, `EventTerminalError`, `EventBudgetWarning`, `EventRecoveryAttempt`.

#### Scenario: Constants replace string literals
- **WHEN** the turn runner records a trace event
- **THEN** it SHALL use typed `EventType` constants instead of raw string literals

### Requirement: Delegation graph computation
The system SHALL provide a `BuildDelegationGraph([]Trace, []Event) DelegationGraph` pure function that computes a directed graph of agent-to-agent handoffs from trace events. `DelegationGraph` SHALL contain `Edges []DelegationEdge` and `Agents map[string]AgentNode` with per-agent delegation counts.

#### Scenario: Graph from delegation events
- **WHEN** trace events include delegation from "orchestrator" to "operator" and back
- **THEN** `BuildDelegationGraph` SHALL return edges `[{From: orchestrator, To: operator}, {From: operator, To: orchestrator}]`

### Requirement: Agent metrics computation
The system SHALL provide a `ComputeAgentMetrics([]Trace, []Event) map[string]*AgentMetricsSummary` pure function that derives per-agent performance statistics including turn count, success/failure rates, tool call count, delegation counts, and duration percentiles (p50, p95, p99). Trace attribution SHALL prefer the first delegation target for delegated turns. For non-delegated turns, attribution SHALL come from trace/event agent evidence rather than transport entrypoints or the operator's current runtime configuration.

#### Scenario: Metrics from traces
- **WHEN** 10 traces exist with 3 failures for agent "navigator"
- **THEN** `ComputeAgentMetrics` SHALL return `navigator.FailureCount == 3` and `navigator.SuccessRate == 0.7`

#### Scenario: Non-delegated turn uses event author
- **WHEN** a trace has no delegation events and its first attributable event author is `lango-agent`
- **THEN** `ComputeAgentMetrics` SHALL attribute that turn to `lango-agent`
- **AND** it SHALL NOT attribute the turn to transport names such as `tui` or `gateway`

#### Scenario: Historical metrics ignore current config mode
- **WHEN** stored traces were created in single-agent mode and the current runtime is now multi-agent
- **THEN** `ComputeAgentMetrics` SHALL still attribute non-delegated historical turns from trace/event evidence
- **AND** it SHALL NOT relabel those turns solely from the current config mode

### Requirement: Trace retention cleaner
The system SHALL provide a `RetentionCleaner` lifecycle component that periodically purges traces older than `observability.traceStore.maxAge` (default: 30 days) and keeps total count below `observability.traceStore.maxTraces` (default: 10000). Failed traces SHALL be retained `failedTraceMultiplier` times longer (default: 2x).

#### Scenario: Purge old successful traces
- **WHEN** cleanup interval fires and traces older than `maxAge` exist with outcome `success`
- **THEN** the cleaner SHALL delete those traces and their associated events

#### Scenario: Retain failed traces longer
- **WHEN** cleanup interval fires and failed traces older than `maxAge` but younger than `maxAge * failedTraceMultiplier` exist
- **THEN** the cleaner SHALL NOT delete those traces
