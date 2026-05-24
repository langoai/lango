## Purpose

Slack channel adapter for the Lango agent. Connects to Slack via Socket Mode, handles Events API events (mentions and DMs), Block Kit formatting, approval workflows, and thinking placeholder messages.
## Requirements
### Requirement: Slack app connection
The system SHALL connect to Slack using Socket Mode with app and bot tokens.

#### Scenario: Successful connection
- **WHEN** the application starts with valid SLACK_BOT_TOKEN and SLACK_APP_TOKEN
- **THEN** the system SHALL establish a Socket Mode connection

#### Scenario: Token refresh
- **WHEN** the bot token expires
- **THEN** the system SHALL handle OAuth refresh if configured

### Requirement: Event handling
The system SHALL process incoming Slack events using the Events API. Message handler invocations SHALL NOT block the event loop, allowing concurrent processing of interactive events.

#### Scenario: App mention event
- **WHEN** a user mentions the bot in a channel
- **THEN** the event SHALL be forwarded to the agent

#### Scenario: Direct message event
- **WHEN** a user sends a DM to the bot
- **THEN** the message SHALL be processed by the agent

#### Scenario: Concurrent message and interactive event processing
- **WHEN** a message handler is blocking (e.g., waiting for tool approval)
- **THEN** the event loop SHALL remain free to process interactive events (button clicks, approval callbacks)

#### Scenario: Graceful shutdown with active handlers
- **WHEN** the channel is stopped while message handlers are running
- **THEN** the system SHALL wait for all active handlers to complete before shutting down

### Requirement: Message sending
The Slack channel SHALL auto-convert standard Markdown to Slack mrkdwn format in the Send() method before posting messages. The conversion SHALL apply to the text content passed via MsgOptionText.

#### Scenario: Auto-format standard Markdown
- **WHEN** Send() is called with an OutgoingMessage
- **THEN** the system converts the text via FormatMrkdwn() before creating MsgOptionText

#### Scenario: Thread reply preserved
- **WHEN** Send() is called with ThreadTS set
- **THEN** the system applies mrkdwn conversion and replies in the specified thread

#### Scenario: Block Kit content unaffected
- **WHEN** Send() is called with Blocks specified
- **THEN** the system converts the text field but Block Kit content is sent as-is

### Requirement: Block Kit formatting
The system SHALL format rich responses using Slack Block Kit.

#### Scenario: Code block formatting
- **WHEN** a response contains code
- **THEN** the code SHALL be formatted using Block Kit code blocks

#### Scenario: Action buttons
- **WHEN** an interactive response is needed
- **THEN** Block Kit buttons SHALL be included in the message

### Requirement: Workspace configuration
The system SHALL support multi-workspace installation.

#### Scenario: Workspace-specific settings
- **WHEN** the bot is installed in multiple workspaces
- **THEN** each workspace SHALL use its own configuration

### Requirement: Slack approval provider
The Slack channel SHALL provide an approval provider that uses Block Kit action buttons for tool execution approval.

#### Scenario: Approval message sent
- **WHEN** a sensitive tool approval is requested for a Slack session
- **THEN** the system SHALL post a message with Block Kit action block containing "Approve" (primary style) and "Deny" (danger style) buttons to the originating channel

#### Scenario: User approves
- **WHEN** the user clicks the "Approve" button
- **THEN** the original message SHALL be updated to show approval status (buttons removed)
- **AND** the tool execution SHALL proceed

#### Scenario: User denies
- **WHEN** the user clicks the "Deny" button
- **THEN** the original message SHALL be updated to show denial status (buttons removed)
- **AND** the tool execution SHALL be denied

#### Scenario: Approval timeout
- **WHEN** no button is clicked within the timeout period
- **THEN** the approval request SHALL be denied with a timeout error

### Requirement: Approval message editing on timeout and cancellation
The Slack approval provider SHALL update the approval message to display "Expired" status and remove action buttons when the approval times out or the context is cancelled.

#### Scenario: Timeout removes buttons
- **WHEN** an approval request times out without user response
- **THEN** the system SHALL call UpdateMessage with text "🔐 Tool approval — ⏱ Expired" and empty MsgOptionBlocks to remove action buttons

#### Scenario: Context cancellation removes buttons
- **WHEN** the approval request context is cancelled
- **THEN** the system SHALL call UpdateMessage with expired text and empty blocks

### Requirement: TOCTOU-safe interactive callback handling
The Slack approval provider SHALL use a single `LoadAndDelete` call as the first operation in HandleInteractive to atomically claim the pending request, preventing the race condition between Load and a concurrent timeout Delete.

#### Scenario: First action click succeeds
- **WHEN** a user clicks an approval button for a pending request
- **THEN** the system SHALL atomically load and delete the pending entry, update the message with approval status, and deliver the result

#### Scenario: Duplicate action click is silently ignored
- **WHEN** a second action arrives for an already-processed request
- **THEN** the system SHALL return immediately without updating the message or delivering a result

### Requirement: Button removal on approval or denial
The Slack approval provider SHALL pass empty `MsgOptionBlocks()` when updating the approval message after user action to remove action buttons.

#### Scenario: Approved message has no buttons
- **WHEN** a user approves a tool request
- **THEN** the updated message SHALL contain no action blocks

### Requirement: Interactive event handling
The Slack channel event loop SHALL handle `EventTypeInteractive` socket mode events and route block_actions to the approval provider.

#### Scenario: Interactive event received
- **WHEN** an EventTypeInteractive event is received with type block_actions
- **THEN** each action SHALL be routed to the approval provider's HandleInteractive method

### Requirement: Client interface extension
The Slack Client interface SHALL include an `UpdateMessage` method for editing approval messages after a response.

#### Scenario: Update approval message
- **WHEN** an approval response is received
- **THEN** the system SHALL use `UpdateMessage` to edit the original message

### Requirement: Thinking placeholder during processing
The Slack channel SHALL post a placeholder message ("Thinking...") when a user message is received, then replace it with the actual response when the handler completes.

#### Scenario: Placeholder posted and replaced
- **WHEN** a user sends a message to the Slack bot
- **THEN** the bot SHALL post a "_Thinking..._" placeholder message to the channel
- **AND** SHALL replace the placeholder with the formatted response via `UpdateMessage`

#### Scenario: Placeholder post failure
- **WHEN** the placeholder `PostMessage` call fails
- **THEN** the bot SHALL fall back to sending the response as a new message

#### Scenario: Placeholder update failure
- **WHEN** the `UpdateMessage` call to replace the placeholder fails
- **THEN** the bot SHALL send the response as a new message instead

#### Scenario: Handler error with placeholder
- **WHEN** the message handler returns an error after a placeholder was posted
- **THEN** the bot SHALL update the placeholder with the error message

### Requirement: Public StartTyping with placeholder pattern
The Slack channel SHALL expose a public `StartTyping(channelID string) func()` method that posts a `_Processing..._` placeholder message. The returned stop function SHALL delete the placeholder message.

#### Scenario: Successful placeholder lifecycle
- **WHEN** `StartTyping` is called and the stop function is subsequently called
- **THEN** the placeholder message SHALL be posted on start and deleted on stop

#### Scenario: Post failure returns no-op
- **WHEN** posting the placeholder message fails
- **THEN** the error SHALL be logged at Warn level and a no-op stop function SHALL be returned

#### Scenario: Stop function is idempotent
- **WHEN** the returned stop function is called multiple times
- **THEN** no panic SHALL occur (protected by `sync.Once`)

### Requirement: Slack Client interface includes DeleteMessage
The Slack `Client` interface SHALL include a `DeleteMessage(channelID, messageTimestamp string) (string, string, error)` method for placeholder cleanup.

#### Scenario: Mock clients implement DeleteMessage
- **WHEN** test mock clients implement the `Client` interface
- **THEN** they SHALL include a `DeleteMessage` method

### Requirement: Slack HTTP client timeout
The Slack channel SHALL use a bounded default HTTP client for Slack REST/API operations when no custom HTTP client is provided. Custom HTTP clients supplied through `Config.HTTPClient` SHALL be preserved unchanged.

#### Scenario: Default Slack HTTP client is bounded
- **WHEN** the Slack channel is created without `Config.HTTPClient`
- **THEN** the Slack SDK client SHALL use an HTTP client with a finite timeout
- **AND** the timeout SHALL be at least 10 seconds

#### Scenario: Custom Slack HTTP client is preserved
- **WHEN** the Slack channel is created with `Config.HTTPClient`
- **THEN** the provided HTTP client SHALL be used unchanged

