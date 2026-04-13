## MODIFIED Requirements

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
