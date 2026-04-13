## ADDED Requirements

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
