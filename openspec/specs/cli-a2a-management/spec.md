# CLI A2A Management

## Purpose
Provides CLI commands for managing Agent-to-Agent (A2A) protocol configuration, including viewing the local agent card and checking remote agent cards.

## Requirements

### Requirement: A2A card command
The system SHALL provide a `lango a2a card [--json]` command that displays the local agent's A2A agent card including name, description, capabilities, and endpoint URL. The command SHALL use cfgLoader to read the A2A configuration.

#### Scenario: A2A enabled
- **WHEN** user runs `lango a2a card` with a2a.enabled set to true
- **THEN** system displays agent name, description, URL, capabilities, and supported protocols

#### Scenario: A2A disabled
- **WHEN** user runs `lango a2a card` with a2a.enabled set to false
- **THEN** system displays "A2A protocol is not enabled"

#### Scenario: A2A card in JSON format
- **WHEN** user runs `lango a2a card --json`
- **THEN** system outputs a JSON object matching the A2A agent card schema

### Requirement: A2A check command
The system SHALL provide a `lango a2a check <url> [--json]` command that fetches a remote agent's A2A agent card from the given URL and displays its contents. The command SHALL validate the card structure and report any issues.

#### Scenario: Valid remote card
- **WHEN** user runs `lango a2a check https://agent.example.com`
- **THEN** system fetches the agent card from the URL and displays name, capabilities, and protocol version

#### Scenario: Unreachable URL
- **WHEN** user runs `lango a2a check https://unreachable.example.com`
- **THEN** system returns error indicating the remote agent is unreachable

#### Scenario: Invalid card format
- **WHEN** user runs `lango a2a check <url>` and the response is not a valid agent card
- **THEN** system returns error indicating the card format is invalid

### Requirement: A2A command group entry
The system SHALL provide a `lango a2a` command group that shows help text listing card and check subcommands.

#### Scenario: Help text
- **WHEN** user runs `lango a2a`
- **THEN** system displays help listing card and check subcommands
