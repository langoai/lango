## Purpose

The Channel Approval capability provides a unified interface for routing tool execution approval requests through channel-native interactive components. It defines the core `Provider` interface, composite routing logic, and fallback mechanisms (Gateway WebSocket, TTY terminal).

## Requirements

### Requirement: Approval Provider interface
The system SHALL define a `Provider` interface with `RequestApproval(ctx, req) (ApprovalResponse, error)` and `CanHandle(sessionKey) bool` methods for handling tool execution approval requests. `ApprovalResponse` SHALL carry `Approved bool` and `AlwaysAllow bool` fields.

#### Scenario: Provider implementation
- **WHEN** a new approval channel is added
- **THEN** it SHALL implement the `Provider` interface returning `ApprovalResponse`
- **AND** `CanHandle` SHALL return true only for session keys it can handle

#### Scenario: Approve response
- **WHEN** a user approves a request
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: false}`

#### Scenario: Always Allow response
- **WHEN** a user clicks "Always Allow"
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: true}`

#### Scenario: Deny response
- **WHEN** a user denies a request
- **THEN** the provider SHALL return `ApprovalResponse{Approved: false, AlwaysAllow: false}`

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
The `TTYProvider.RequestApproval` SHALL return `(ApprovalResponse{}, error)` when stdin is not a terminal. When stdin is a terminal, it SHALL prompt with `[y/a/N]` where `a` means "always allow".

#### Scenario: Non-terminal environment returns error
- **WHEN** `TTYProvider.RequestApproval` is called and stdin is not a terminal
- **THEN** it SHALL return an empty `ApprovalResponse` and a non-nil error containing "not a terminal"

#### Scenario: Terminal user types 'a'
- **WHEN** the user enters "a" or "always" at the TTY prompt
- **THEN** it SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: true}`

#### Scenario: Terminal user types 'y'
- **WHEN** the user enters "y" or "yes" at the TTY prompt
- **THEN** it SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: false}`

### Requirement: Gateway approval provider
The system SHALL provide a `GatewayProvider` that delegates approval to connected companion apps via WebSocket. The `GatewayApprover` interface SHALL return `(ApprovalResponse, error)`.

#### Scenario: Companions connected
- **WHEN** a companion app is connected
- **THEN** `CanHandle` SHALL return true
- **AND** the approval request SHALL be forwarded to companions

#### Scenario: No companions connected
- **WHEN** no companion app is connected
- **THEN** `CanHandle` SHALL return false

#### Scenario: Companion sends alwaysAllow
- **WHEN** a companion responds with `{"approved": true, "alwaysAllow": true}`
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: true}`

#### Scenario: Companion omits alwaysAllow (backward compatible)
- **WHEN** a companion responds with `{"approved": true}` without `alwaysAllow`
- **THEN** the provider SHALL return `ApprovalResponse{Approved: true, AlwaysAllow: false}`

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

### Requirement: Approval callback unblocks agent before UI update
`HandleCallback` SHALL send the approval result to the waiting agent's channel BEFORE calling `editApprovalMessage` to update the Telegram message. This ensures the agent pipeline is not blocked by Telegram API latency.

#### Scenario: Approval granted
- **WHEN** a user clicks "Approve" on an inline keyboard
- **THEN** the approval result (true) is sent to the agent's channel immediately, and THEN the Telegram message is edited to show "Approved" status

#### Scenario: Approval denied
- **WHEN** a user clicks "Deny" on an inline keyboard
- **THEN** the denial result (false) is sent to the agent's channel immediately, and THEN the Telegram message is edited to show "Denied" status

#### Scenario: Multiple consecutive approvals
- **WHEN** 4 tools require consecutive approval and all are approved
- **THEN** the agent processes each approval without cumulative Telegram API latency between them

### Requirement: Telegram Always Allow button
The Telegram approval message SHALL include a second row with an "Always Allow" button using the callback data prefix `always:`.

#### Scenario: Always Allow button layout
- **WHEN** an approval message is sent in Telegram
- **THEN** it SHALL have Row 1 with Approve and Deny buttons, and Row 2 with an Always Allow button

#### Scenario: Always Allow callback
- **WHEN** a user clicks "Always Allow" in Telegram
- **THEN** the response channel SHALL receive `ApprovalResponse{Approved: true, AlwaysAllow: true}`
- **AND** the message SHALL be edited to show "Always Allowed"

### Requirement: Discord Always Allow button
The Discord approval message SHALL include a second ActionsRow with a Secondary-style "Always Allow" button using the custom ID prefix `always:`.

#### Scenario: Always Allow button layout
- **WHEN** an approval message is sent in Discord
- **THEN** it SHALL have ActionsRow 1 with Approve and Deny, and ActionsRow 2 with Always Allow

#### Scenario: Always Allow interaction
- **WHEN** a user clicks "Always Allow" in Discord
- **THEN** the response channel SHALL receive `ApprovalResponse{Approved: true, AlwaysAllow: true}`
- **AND** the interaction response SHALL show "Always Allowed"

### Requirement: Slack Always Allow button
The Slack approval message SHALL include an "Always Allow" button in the action block with the action ID prefix `always:`.

#### Scenario: Always Allow button in action block
- **WHEN** an approval message is sent in Slack
- **THEN** the action block SHALL contain Approve, Deny, and Always Allow buttons

#### Scenario: Always Allow interactive action
- **WHEN** a user clicks "Always Allow" in Slack
- **THEN** the response channel SHALL receive `ApprovalResponse{Approved: true, AlwaysAllow: true}`
- **AND** the message SHALL be updated to show "Always Allowed"

### Requirement: Audit log error logging
Tool handlers that call `store.SaveAuditLog` SHALL log a warning via `logger().Warnw` if the audit log write fails, rather than discarding the error with `_ =`. The tool handler SHALL NOT return this error to the caller (log and degrade gracefully).

#### Scenario: Audit log write failure is logged
- **WHEN** `store.SaveAuditLog` returns a non-nil error during `save_knowledge` tool execution
- **THEN** a warning log SHALL be emitted with the action name and error details
- **AND** the tool SHALL still return success to the caller
