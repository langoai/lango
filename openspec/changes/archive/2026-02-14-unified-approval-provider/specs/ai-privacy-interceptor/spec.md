## MODIFIED Requirements

### Requirement: Approval Workflow
The system SHALL block execution of "sensitive" tools (configured list) and require explicit approval before proceeding. The system SHALL use fail-closed semantics: without explicit approval, execution is denied. Approval requests SHALL be routed through a `CompositeProvider` that selects the appropriate approval channel based on the originating session key.

#### Scenario: Channel-specific approval (Telegram)
- **WHEN** the agent attempts to call a sensitive tool
- **AND** the request originates from a Telegram session (session key starts with "telegram:")
- **THEN** the approval request SHALL be routed to the Telegram approval provider

#### Scenario: Channel-specific approval (Discord)
- **WHEN** the agent attempts to call a sensitive tool
- **AND** the request originates from a Discord session (session key starts with "discord:")
- **THEN** the approval request SHALL be routed to the Discord approval provider

#### Scenario: Channel-specific approval (Slack)
- **WHEN** the agent attempts to call a sensitive tool
- **AND** the request originates from a Slack session (session key starts with "slack:")
- **THEN** the approval request SHALL be routed to the Slack approval provider

#### Scenario: Companion approval granted
- **WHEN** the agent attempts to call a sensitive tool
- **AND** no channel-specific provider matches
- **AND** a companion is connected
- **AND** the companion approves the request
- **THEN** the tool execution SHALL proceed

#### Scenario: Companion approval denied
- **WHEN** the agent attempts to call a sensitive tool
- **AND** no channel-specific provider matches
- **AND** a companion is connected
- **AND** the companion denies the request
- **THEN** the tool execution SHALL be denied with an error

#### Scenario: Companion approval error
- **WHEN** the agent attempts to call a sensitive tool
- **AND** no channel-specific provider matches
- **AND** a companion is connected
- **AND** the approval request encounters an error
- **THEN** the tool execution SHALL be denied (fail-closed)

#### Scenario: TTY fallback approval
- **WHEN** the agent attempts to call a sensitive tool
- **AND** no channel-specific provider matches
- **AND** no companion is connected
- **AND** stdin is a terminal (TTY)
- **THEN** the system SHALL prompt the user via stderr with "Allow? [y/N]"
- **AND** proceed only if user responds "y" or "yes"

#### Scenario: No approval source available
- **WHEN** the agent attempts to call a sensitive tool
- **AND** no channel-specific provider matches
- **AND** no companion is connected
- **AND** stdin is not a terminal
- **THEN** the tool execution SHALL be denied

## ADDED Requirements

### Requirement: Approval timeout configuration
The system SHALL support an `ApprovalTimeoutSec` configuration field in `InterceptorConfig` that controls how long to wait for an approval response before timing out.

#### Scenario: Default timeout
- **WHEN** `ApprovalTimeoutSec` is not set or is 0
- **THEN** the system SHALL use a default timeout of 30 seconds

#### Scenario: Custom timeout
- **WHEN** `ApprovalTimeoutSec` is set to a positive value
- **THEN** the system SHALL use that value as the timeout in seconds
