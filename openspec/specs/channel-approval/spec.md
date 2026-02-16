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
The system SHALL provide a `CompositeProvider` that routes approval requests to the first registered provider whose `CanHandle` returns true for the given session key.

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

### Requirement: Thread-safe provider registration
The system SHALL allow providers to be registered concurrently without data races.

#### Scenario: Concurrent registration
- **WHEN** multiple providers are registered from different goroutines
- **THEN** all registrations SHALL complete without data races

### Requirement: TTY approval fallback
The system SHALL provide a `TTYProvider` that prompts the user via terminal stdin for approval as a last-resort fallback.

#### Scenario: TTY prompt
- **WHEN** TTY fallback is invoked
- **AND** stdin is a terminal
- **THEN** the system SHALL print a prompt to stderr and read y/N from stdin

#### Scenario: Non-terminal stdin
- **WHEN** TTY fallback is invoked
- **AND** stdin is not a terminal
- **THEN** the request SHALL be denied (return false, nil)

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
