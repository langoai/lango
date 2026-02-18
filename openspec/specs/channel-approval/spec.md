## Purpose

The Channel Approval capability provides a unified interface for routing tool execution approval requests through channel-native interactive components. It defines the core `Provider` interface, composite routing logic, and fallback mechanisms (Gateway WebSocket, TTY terminal).

## Requirements

### Requirement: Approval Provider interface
The system SHALL define a `Provider` interface with `RequestApproval(ctx, req) (bool, error)` and `CanHandle(sessionKey) bool` methods for handling tool execution approval requests.

#### Scenario: Provider implementation
- **WHEN** a new approval channel is added
- **THEN** it SHALL implement the `Provider` interface
- **AND** `CanHandle` SHALL return true only for session keys it can handle

### Requirement: Composite approval routing
The system SHALL provide a `CompositeProvider` that routes approval requests to the first registered provider whose `CanHandle` returns true for the given session key. The session key used for routing MAY be overridden by an explicit approval target set in the context, which takes precedence over the standard session key.

#### Scenario: Route to matching provider
- **WHEN** an approval request has session key "telegram:123:456"
- **AND** a Telegram provider is registered
- **THEN** the request SHALL be routed to the Telegram provider

#### Scenario: Multiple providers registered
- **WHEN** multiple providers are registered
- **THEN** the first provider whose `CanHandle` returns true SHALL handle the request

#### Scenario: No matching provider with TTY fallback
- **WHEN** no registered provider can handle the session key
- **AND** a TTY fallback is configured
- **THEN** the request SHALL be routed to the TTY fallback

#### Scenario: No matching provider without fallback (fail-closed)
- **WHEN** no registered provider can handle the session key
- **AND** no TTY fallback is configured
- **THEN** the request SHALL be denied (return false)
- **AND** an error SHALL be returned with the message `no approval provider for session "<sessionKey>"`

#### Scenario: Approval target overrides session key
- **WHEN** an approval request has session key "cron:MyJob:123"
- **AND** the context has approval target "telegram:123456789"
- **THEN** the request SHALL use "telegram:123456789" for provider matching
- **AND** the request SHALL be routed to the Telegram provider

### Requirement: Thread-safe provider registration
The system SHALL allow providers to be registered concurrently without data races.

#### Scenario: Concurrent registration
- **WHEN** multiple providers are registered from different goroutines
- **THEN** all registrations SHALL complete without data races

### Requirement: TTY approval fallback behavior
The `TTYProvider.RequestApproval` SHALL return `(false, error)` when stdin is not a terminal, with an error message containing "not a terminal". This replaces the previous behavior of returning `(false, nil)` which was indistinguishable from an explicit user denial.

#### Scenario: Non-terminal environment returns error
- **WHEN** `TTYProvider.RequestApproval` is called and stdin is not a terminal
- **THEN** it SHALL return `false` and a non-nil error containing "not a terminal"

#### Scenario: Terminal environment prompts user
- **WHEN** `TTYProvider.RequestApproval` is called and stdin is a terminal
- **THEN** it SHALL prompt on stderr and read the user's response from stdin

### Requirement: Gateway approval provider
The system SHALL provide a `GatewayProvider` that delegates approval to connected companion apps via WebSocket.

#### Scenario: Companions connected
- **WHEN** a companion app is connected
- **THEN** `CanHandle` SHALL return true
- **AND** the approval request SHALL be forwarded to companions

#### Scenario: No companions connected
- **WHEN** no companion app is connected
- **THEN** `CanHandle` SHALL return false

### Requirement: Approval request context
Each approval request SHALL carry an ID, tool name, session key, parameters, a human-readable Summary string, and creation timestamp.

#### Scenario: Request fields
- **WHEN** an approval request is created
- **THEN** it SHALL contain a unique ID, the tool name, the originating session key, tool parameters, a Summary string, and a timestamp

#### Scenario: Summary populated
- **WHEN** a tool approval request is created via wrapWithApproval
- **THEN** the Summary field SHALL be populated by buildApprovalSummary with a human-readable description of the operation

#### Scenario: Empty summary backward compatibility
- **WHEN** an approval request has an empty Summary
- **THEN** providers SHALL display the existing tool-name-only message

### Requirement: Approval summary rendering
All approval providers SHALL include the Summary field in their approval messages when it is non-empty.

#### Scenario: Gateway provider summary
- **WHEN** a GatewayProvider receives a request with Summary "Execute: ls -la"
- **THEN** the message sent to companions SHALL include the Summary text

#### Scenario: TTY provider summary
- **WHEN** a TTYProvider receives a request with Summary "Delete: /tmp/test"
- **THEN** the terminal prompt SHALL display the Summary on a separate line before the y/N prompt

#### Scenario: Headless provider summary
- **WHEN** a HeadlessProvider receives a request with Summary
- **THEN** the audit log entry SHALL include a "summary" field

#### Scenario: Telegram provider summary
- **WHEN** a Telegram ApprovalProvider receives a request with Summary
- **THEN** the InlineKeyboard message SHALL include the Summary text below the tool name

#### Scenario: Discord provider summary
- **WHEN** a Discord ApprovalProvider receives a request with Summary
- **THEN** the button message Content SHALL include the Summary in a code block

#### Scenario: Slack provider summary
- **WHEN** a Slack ApprovalProvider receives a request with Summary
- **THEN** the Block Kit section text SHALL include the Summary in a code block

### Requirement: Approval summary builder
The system SHALL provide a `buildApprovalSummary(toolName, params)` function that generates human-readable descriptions of tool invocations.

#### Scenario: Exec tool summary
- **WHEN** buildApprovalSummary is called with toolName "exec" and params containing command "curl https://api.example.com"
- **THEN** it SHALL return "Execute: curl https://api.example.com"

#### Scenario: File write summary
- **WHEN** buildApprovalSummary is called with toolName "fs_write" and params containing path "/tmp/test.txt" and content of 100 bytes
- **THEN** it SHALL return "Write to /tmp/test.txt (100 bytes)"

#### Scenario: Unknown tool summary
- **WHEN** buildApprovalSummary is called with an unrecognized toolName
- **THEN** it SHALL return "Tool: <toolName>"

#### Scenario: Long command truncation
- **WHEN** a command string exceeds 200 characters
- **THEN** it SHALL be truncated to 200 characters with "..." appended

### Requirement: Session key context propagation
The `runAgent` function SHALL inject the session key into the context via `WithSessionKey` before passing it to the agent pipeline, ensuring downstream components (approval providers, learning engine) can access the session key via `SessionKeyFromContext`.

#### Scenario: Channel message triggers agent with session key
- **WHEN** a Telegram/Discord/Slack handler calls `runAgent(ctx, sessionKey, input)`
- **THEN** `runAgent` SHALL call `WithSessionKey(ctx, sessionKey)` before invoking the agent
- **AND** `SessionKeyFromContext` SHALL return the session key within the agent pipeline

#### Scenario: Session key reaches approval provider
- **WHEN** a tool requiring approval is invoked from a channel message
- **THEN** the `ApprovalRequest.SessionKey` field SHALL contain the channel session key (e.g., "telegram:123:456")
- **AND** `CompositeProvider` SHALL route to the matching channel provider

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
