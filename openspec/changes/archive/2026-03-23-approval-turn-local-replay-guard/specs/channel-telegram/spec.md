## MODIFIED Requirements

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
- **THEN** the approval request SHALL be denied with an error wrapping `approval.ErrTimeout`

#### Scenario: Context cancellation
- **WHEN** the request context is cancelled before a response
- **THEN** the approval request SHALL return the context error

#### Scenario: Approval outcome logs include provider metadata
- **WHEN** the Telegram provider processes approval request, callback, approval, denial, or expiry
- **THEN** it SHALL emit structured logs including provider=`telegram`, request ID, tool, session, and outcome
