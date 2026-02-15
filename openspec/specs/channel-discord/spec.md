## ADDED Requirements

### Requirement: Discord bot connection
The system SHALL connect to Discord using the Bot API with a provided bot token.

#### Scenario: Successful bot connection
- **WHEN** the application starts with a valid DISCORD_BOT_TOKEN
- **THEN** the system SHALL establish a gateway connection to Discord

#### Scenario: Gateway reconnection
- **WHEN** the Discord gateway connection drops
- **THEN** the system SHALL automatically reconnect with session resumption

### Requirement: Message reception
The system SHALL receive and process incoming messages from Discord channels.

#### Scenario: Direct message received
- **WHEN** a user sends a DM to the bot
- **THEN** the message SHALL be forwarded to the agent

#### Scenario: Channel message with mention
- **WHEN** a user mentions the bot in a text channel
- **THEN** the message SHALL be processed if the channel is configured

### Requirement: Slash command support
The system SHALL register and handle Discord slash commands.

#### Scenario: Slash command registration
- **WHEN** the bot connects to Discord
- **THEN** configured slash commands SHALL be registered with the Discord API

#### Scenario: Slash command invocation
- **WHEN** a user invokes a slash command
- **THEN** the system SHALL execute the corresponding handler

### Requirement: Message sending
The system SHALL send agent responses back to Discord channels.

#### Scenario: Send text response
- **WHEN** the agent generates a response
- **THEN** the response SHALL be sent to the originating channel

#### Scenario: Discord markdown formatting
- **WHEN** a response contains markdown
- **THEN** the response SHALL use Discord-compatible markdown syntax

### Requirement: Guild configuration
The system SHALL support per-guild configuration for Discord servers.

#### Scenario: Guild-specific settings
- **WHEN** a guild has custom configuration
- **THEN** the bot SHALL use guild-specific settings for that server

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
