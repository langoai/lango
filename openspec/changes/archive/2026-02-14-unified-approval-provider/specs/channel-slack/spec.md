## ADDED Requirements

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
