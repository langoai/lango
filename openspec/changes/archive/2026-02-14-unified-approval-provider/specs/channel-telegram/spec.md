## ADDED Requirements

### Requirement: Telegram approval provider
The Telegram channel SHALL provide an approval provider that uses InlineKeyboard buttons for tool execution approval.

#### Scenario: Approval message sent
- **WHEN** a sensitive tool approval is requested for a Telegram session
- **THEN** the system SHALL send a message with InlineKeyboard containing "Approve" and "Deny" buttons to the originating chat

#### Scenario: User approves
- **WHEN** the user clicks the "Approve" button
- **THEN** the callback query SHALL be answered
- **AND** the original message SHALL be edited to show approval status
- **AND** the tool execution SHALL proceed

#### Scenario: User denies
- **WHEN** the user clicks the "Deny" button
- **THEN** the callback query SHALL be answered
- **AND** the original message SHALL be edited to show denial status
- **AND** the tool execution SHALL be denied

#### Scenario: Approval timeout
- **WHEN** no button is clicked within the timeout period
- **THEN** the approval request SHALL be denied with a timeout error

#### Scenario: Context cancellation
- **WHEN** the request context is cancelled before a response
- **THEN** the approval request SHALL return the context error

### Requirement: Callback query routing
The Telegram channel event loop SHALL route CallbackQuery updates to the approval provider before processing regular messages.

#### Scenario: CallbackQuery received
- **WHEN** an update contains a CallbackQuery
- **THEN** the update SHALL be routed to the approval provider's HandleCallback method
- **AND** regular message processing SHALL be skipped for that update

### Requirement: BotAPI Request method
The BotAPI interface SHALL include a `Request` method for operations that return an APIResponse (e.g., AnswerCallbackQuery).

#### Scenario: Answer callback query
- **WHEN** a callback query needs to be acknowledged
- **THEN** the system SHALL use the `Request` method to send an AnswerCallbackQuery
