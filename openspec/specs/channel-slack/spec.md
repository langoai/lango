## ADDED Requirements

### Requirement: Slack app connection
The system SHALL connect to Slack using Socket Mode with app and bot tokens.

#### Scenario: Successful connection
- **WHEN** the application starts with valid SLACK_BOT_TOKEN and SLACK_APP_TOKEN
- **THEN** the system SHALL establish a Socket Mode connection

#### Scenario: Token refresh
- **WHEN** the bot token expires
- **THEN** the system SHALL handle OAuth refresh if configured

### Requirement: Event handling
The system SHALL process incoming Slack events using the Events API.

#### Scenario: App mention event
- **WHEN** a user mentions the bot in a channel
- **THEN** the event SHALL be forwarded to the agent

#### Scenario: Direct message event
- **WHEN** a user sends a DM to the bot
- **THEN** the message SHALL be processed by the agent

### Requirement: Message sending
The system SHALL send agent responses back to Slack channels.

#### Scenario: Send to channel
- **WHEN** the agent generates a response to a channel message
- **THEN** the response SHALL be posted to that channel

#### Scenario: Thread reply
- **WHEN** the original message was in a thread
- **THEN** the response SHALL be posted as a thread reply

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
