## ADDED Requirements

### Requirement: Telegram bot connection
The system SHALL connect to Telegram using the Bot API with a provided bot token.

#### Scenario: Successful bot connection
- **WHEN** the application starts with a valid TELEGRAM_BOT_TOKEN
- **THEN** the system SHALL establish a connection to Telegram servers

#### Scenario: Invalid bot token
- **WHEN** the bot token is invalid or revoked
- **THEN** the system SHALL log an error and retry with exponential backoff

### Requirement: Message reception
The system SHALL receive and process incoming messages from Telegram chats.

#### Scenario: Direct message received
- **WHEN** a user sends a direct message to the bot
- **THEN** the message SHALL be forwarded to the agent with sender context

#### Scenario: Group message with mention
- **WHEN** a user mentions the bot in a group chat
- **THEN** the message SHALL be processed if mention-gating allows

### Requirement: Message sending
The system SHALL send agent responses back to Telegram chats.

#### Scenario: Send text response
- **WHEN** the agent generates a text response
- **THEN** the response SHALL be sent to the originating chat

#### Scenario: Long message chunking
- **WHEN** a response exceeds Telegram's 4096 character limit
- **THEN** the response SHALL be split into multiple messages

### Requirement: Media handling
The system SHALL process media attachments from Telegram messages.

#### Scenario: Image attachment
- **WHEN** a user sends an image
- **THEN** the system SHALL download the image and provide it to the agent

#### Scenario: Document attachment
- **WHEN** a user sends a document
- **THEN** the system SHALL download and make it available for processing

### Requirement: Allowlist filtering
The system SHALL filter incoming messages based on configured allowlists.

#### Scenario: Allowed sender
- **WHEN** a message is from an allowed user/group
- **THEN** the message SHALL be processed normally

#### Scenario: Unknown sender
- **WHEN** a message is from an unknown sender and pairing is enabled
- **THEN** a pairing code SHALL be sent to the sender

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
