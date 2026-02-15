## ADDED Requirements

### Requirement: Discord approval provider
The Discord channel SHALL provide an approval provider that uses Message Component buttons for tool execution approval.

#### Scenario: Approval message sent
- **WHEN** a sensitive tool approval is requested for a Discord session
- **THEN** the system SHALL send a message with an ActionsRow containing "Approve" (success style) and "Deny" (danger style) buttons to the originating channel

#### Scenario: User approves
- **WHEN** the user clicks the "Approve" button
- **THEN** the interaction SHALL be responded to with an updated message (buttons removed)
- **AND** the tool execution SHALL proceed

#### Scenario: User denies
- **WHEN** the user clicks the "Deny" button
- **THEN** the interaction SHALL be responded to with an updated message (buttons removed)
- **AND** the tool execution SHALL be denied

#### Scenario: Approval timeout
- **WHEN** no button is clicked within the timeout period
- **THEN** the approval request SHALL be denied with a timeout error

### Requirement: Interaction handler registration
The Discord channel SHALL register an InteractionCreate handler at startup to route message component interactions to the approval provider.

#### Scenario: InteractionCreate received
- **WHEN** an InteractionCreate event of type InteractionMessageComponent is received
- **THEN** the event SHALL be routed to the approval provider's HandleInteraction method

### Requirement: Session interface extensions
The Discord Session interface SHALL include `InteractionRespond` and `ChannelMessageEditComplex` methods for approval interaction handling.

#### Scenario: Respond to interaction
- **WHEN** a button interaction needs to be acknowledged
- **THEN** the system SHALL use `InteractionRespond` to update the original message
