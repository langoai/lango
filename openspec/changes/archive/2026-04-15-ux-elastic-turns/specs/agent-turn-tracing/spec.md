## MODIFIED Requirements

### Requirement: Append-only per-turn trace journal
Every agent turn SHALL create an append-only trace identified by a stable trace ID. The trace SHALL record session key, entrypoint, start/end timestamps, user input metadata, delegation events, tool calls, tool results, retries, and final outcome. When the Runner retries a turn due to transient failures, all retry attempts and recovery events SHALL accumulate within the same trace instance. Each retry attempt SHALL NOT create a separate trace.

#### Scenario: Multi-agent turn records delegation and tool activity
- **WHEN** the orchestrator delegates to `vault` and `vault` calls `payment_balance`
- **THEN** the turn trace SHALL record the delegation event
- **AND** SHALL record the specialist tool call and tool result with agent name, tool name, and call identity

#### Scenario: Trace survives post-turn diagnostics
- **WHEN** a turn completes and later diagnostic tooling inspects the latest trace for the session
- **THEN** the trace SHALL still contain the recorded event sequence and classified outcome

#### Scenario: Retry attempts accumulate in single trace
- **WHEN** the Runner retries a turn after a transient provider error
- **THEN** the recovery event SHALL be recorded in the same trace as the original attempt
- **AND** a new trace SHALL NOT be created for the retry attempt
