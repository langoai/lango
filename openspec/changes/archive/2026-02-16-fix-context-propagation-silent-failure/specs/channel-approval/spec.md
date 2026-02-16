## MODIFIED Requirements

### Requirement: TTY approval fallback behavior
The `TTYProvider.RequestApproval` SHALL return `(false, error)` when stdin is not a terminal, with an error message containing "not a terminal". This replaces the previous behavior of returning `(false, nil)` which was indistinguishable from an explicit user denial.

#### Scenario: Non-terminal environment returns error
- **WHEN** `TTYProvider.RequestApproval` is called and stdin is not a terminal
- **THEN** it SHALL return `false` and a non-nil error containing "not a terminal"

#### Scenario: Terminal environment prompts user
- **WHEN** `TTYProvider.RequestApproval` is called and stdin is a terminal
- **THEN** it SHALL prompt on stderr and read the user's response from stdin

## ADDED Requirements

### Requirement: Safe type assertions in approval providers
All channel approval providers (Discord, Telegram, Slack) SHALL use comma-ok pattern when asserting types from `sync.Map` loads. If the type assertion fails, the provider SHALL log a warning and return without panicking.

#### Scenario: Discord approval handles unexpected type
- **WHEN** `HandleInteraction` loads a value from `pending` sync.Map and the type assertion to `chan bool` fails
- **THEN** it SHALL log a warning with the request ID and return without sending to the channel

#### Scenario: Telegram approval handles unexpected type
- **WHEN** `HandleCallback` loads a value from `pending` sync.Map and the type assertion to `chan bool` fails
- **THEN** it SHALL log a warning with the request ID and return without sending to the channel

#### Scenario: Slack approval handles unexpected type
- **WHEN** `HandleInteractive` loads a value from `pending` sync.Map and the type assertion to `*approvalPending` fails
- **THEN** it SHALL log a warning with the request ID and return without sending to the channel

### Requirement: Audit log error logging
Tool handlers that call `store.SaveAuditLog` SHALL log a warning via `logger().Warnw` if the audit log write fails, rather than discarding the error with `_ =`. The tool handler SHALL NOT return this error to the caller (log and degrade gracefully).

#### Scenario: Audit log write failure is logged
- **WHEN** `store.SaveAuditLog` returns a non-nil error during `save_knowledge` tool execution
- **THEN** a warning log SHALL be emitted with the action name and error details
- **AND** the tool SHALL still return success to the caller
