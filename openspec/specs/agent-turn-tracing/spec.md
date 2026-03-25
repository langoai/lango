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
Trace persistence SHALL use detached contexts with their own timeout so trace writes survive parent cancellation briefly but never block indefinitely. Each create, append, and finish operation SHALL use a fresh detached timeout instead of reusing a single run-scoped timeout context.

#### Scenario: Long turn still records terminal trace state
- **WHEN** a turn runs longer than the configured trace-write timeout
- **THEN** later append and finish operations SHALL still receive a fresh detached timeout context
- **AND** trace persistence SHALL continue attempting to record the terminal outcome independently of earlier trace writes

#### Scenario: Parent cancellation does not lose trace immediately
- **WHEN** the parent request context is cancelled after a failure
- **THEN** the trace writer SHALL continue using a detached context long enough to attempt persistence
- **AND** the detached context SHALL time out independently after the configured trace-write timeout

### Requirement: Recovery attempts are recorded in turn traces
Structured recovery attempts SHALL be recorded as trace events with enough metadata to identify reroute-vs-retry behavior during diagnosis.

#### Scenario: Specialist reroute recovery is traced
- **WHEN** structured orchestration retries a failed specialist turn with a reroute hint
- **THEN** the trace SHALL append a `recovery_attempt` event
- **AND** the event payload SHALL include the recovery action and failed specialist name

#### Scenario: Generic retry recovery is traced
- **WHEN** structured orchestration retries a turn without a failed specialist identity
- **THEN** the trace SHALL append a `recovery_attempt` event
- **AND** the event payload SHALL distinguish the generic retry from reroute recovery

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

## MODIFIED Requirements

### Requirement: Turn trace Store interface
The turn trace Store interface SHALL include the following additional methods beyond the existing `CreateTrace`, `AppendEvent`, `FinishTrace`, `RecentFailures`, and `IsolationLeakCount`:

- `EventsForTrace(ctx context.Context, traceID string) ([]Event, error)` — returns all events for a trace, ordered by seq
- `TracesForSession(ctx context.Context, sessionKey string) ([]Trace, error)` — returns all traces for a session, ordered by started_at
- `PurgeTraces(ctx context.Context, traceIDs []string) error` — deletes traces and their associated events
- `TraceCount(ctx context.Context) (int, error)` — returns total trace count
- `OldTraces(ctx context.Context, cutoff time.Time, onlySuccess bool, limit int) ([]string, error)` — returns trace IDs older than cutoff
- `RecentByOutcome(ctx context.Context, outcome Outcome, since time.Time, limit int) ([]Trace, error)` — returns traces matching outcome within time window

All methods SHALL be implemented in `EntStore`. All methods SHALL be nil-safe (return nil/0 when store is nil).

#### Scenario: Query events for trace
- **WHEN** `EventsForTrace` is called with a valid trace ID
- **THEN** it SHALL return all events ordered by sequence number

#### Scenario: Query traces by outcome and time window
- **WHEN** `RecentByOutcome` is called with `OutcomeLoopDetected` and `since` 24 hours ago
- **THEN** it SHALL return only traces with that outcome created after the cutoff

#### Scenario: Purge cascades to events
- **WHEN** `PurgeTraces` is called with trace IDs
- **THEN** both the trace rows and their associated event rows SHALL be deleted

#### Scenario: Nil store returns safely
- **WHEN** any method is called on a nil `EntStore`
- **THEN** it SHALL return nil/0/empty without error
