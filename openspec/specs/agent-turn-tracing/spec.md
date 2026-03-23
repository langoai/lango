# agent-turn-tracing Specification

## Purpose
Shared turn-runtime tracing for channels, gateway, and automation. Captures durable per-turn execution events and classified outcomes so operators can diagnose multi-agent failures from structured evidence instead of generic fallback messages.
## Requirements
### Requirement: Shared turn runner owns execution finalization
The system SHALL route channel, gateway, and automation agent execution through a shared turn runner that owns timeout resolution, trace creation, outcome classification, and response finalization.

#### Scenario: Channel and gateway use the same execution core
- **WHEN** a Telegram channel request and a gateway chat request invoke the agent runtime
- **THEN** both paths SHALL execute through the same turn runner abstraction
- **AND** both paths SHALL receive a structured turn result instead of owning independent empty-response/fallback logic

#### Scenario: Automation path reuses the same execution core
- **WHEN** a cron/background/workflow prompt invokes the agent runtime
- **THEN** the automation path SHALL use the same turn runner abstraction
- **AND** the resulting trace SHALL record the automation entrypoint distinctly from channel/gateway entrypoints

### Requirement: Append-only per-turn trace journal
Every agent turn SHALL create an append-only trace identified by a stable trace ID. The trace SHALL record session key, entrypoint, start/end timestamps, user input metadata, delegation events, tool calls, tool results, retries, and final outcome.

#### Scenario: Multi-agent turn records delegation and tool activity
- **WHEN** the orchestrator delegates to `vault` and `vault` calls `payment_balance`
- **THEN** the turn trace SHALL record the delegation event
- **AND** SHALL record the specialist tool call and tool result with agent name, tool name, and call identity

#### Scenario: Trace survives post-turn diagnostics
- **WHEN** a turn completes and later diagnostic tooling inspects the latest trace for the session
- **THEN** the trace SHALL still contain the recorded event sequence and classified outcome

### Requirement: Classified terminal outcomes
Each turn trace SHALL terminate in exactly one classified outcome: `success`, `user_error`, `model_error`, `timeout`, `empty_after_tool_use`, or `loop_detected`. The trace SHALL include a concise root-cause summary for non-success outcomes.

#### Scenario: Repeated identical specialist calls become loop_detected
- **WHEN** the same specialist repeatedly calls the same tool with canonically identical params within one turn
- **THEN** the trace SHALL terminate with outcome `loop_detected`
- **AND** the root-cause summary SHALL identify the offending agent and tool signature

#### Scenario: Tool-only terminal state becomes empty_after_tool_use
- **WHEN** a specialist uses one or more tools successfully but the turn terminates without any visible assistant completion
- **THEN** the trace SHALL terminate with outcome `empty_after_tool_use`
- **AND** the root-cause summary SHALL mention that tool work completed without final synthesis

### Requirement: Trace-backed diagnostics on failure
When a turn ends in `timeout`, `empty_after_tool_use`, or `loop_detected`, the system SHALL emit structured logs that include the trace ID and classified summary, and SHALL make the latest trace retrievable by internal diagnostics tooling.

#### Scenario: Failure log includes trace linkage
- **WHEN** a turn ends in `empty_after_tool_use`
- **THEN** the emitted structured log SHALL include the trace ID and classified summary
- **AND** operators SHALL be able to use that trace ID to inspect the latest recorded sequence for the session

### Requirement: Non-success turns always record a terminal failure event
Every non-success turn SHALL append a `terminal_error` trace event before the trace is finalized, even when the failure happens before the first normal runtime event.

#### Scenario: Pre-event failure still records terminal_error
- **WHEN** a turn fails before any delegation, tool call, tool result, or assistant text event is recorded
- **THEN** the trace SHALL still contain at least one event
- **AND** that event SHALL be `terminal_error`

### Requirement: Bounded detached trace writes
Trace persistence SHALL use a detached context with its own timeout so trace writes survive parent cancellation briefly but never block indefinitely.

#### Scenario: Parent cancellation does not lose trace immediately
- **WHEN** the parent request context is cancelled after a failure
- **THEN** the trace writer SHALL continue using a detached context long enough to attempt persistence
- **AND** the detached context SHALL time out independently after the configured trace-write timeout

### Requirement: Stable trace payload shape with truncation metadata
Trace payload JSON SHALL remain a single stable JSON string. Payload truncation SHALL be represented via explicit metadata, not by wrapping the payload in a different JSON shape.

#### Scenario: Truncated payload marks metadata only
- **WHEN** a trace payload exceeds the configured storage limit
- **THEN** the stored payload SHALL still be a single JSON string
- **AND** the corresponding trace event SHALL set `payload_truncated=true`

### Requirement: Trace stores operator-facing cause fields
The durable trace summary row SHALL store the classified cause alongside the broad outcome.

#### Scenario: Failed trace stores cause class and detail
- **WHEN** a turn finishes with a non-success outcome
- **THEN** the trace row SHALL persist `error_code`, `cause_class`, and `cause_detail`
- **AND** the trace summary SHALL use the operator-facing diagnostic summary rather than the broad user-facing message

