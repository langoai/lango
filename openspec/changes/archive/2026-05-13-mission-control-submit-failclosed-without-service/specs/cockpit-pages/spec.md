## ADDED Requirements

### Requirement: Mission Control top-level submit fails closed without mission service
The cockpit Mission Control page SHALL preserve its durable-first submit contract by failing closed when no mission service is configured for an ordinary top-level composer submit.

#### Scenario: Ordinary submit returns explicit system message without mission service
- **WHEN** the Mission Control composer contains a non-slash request
- **AND** no mission service is configured
- **THEN** submit handling SHALL return an explicit system message explaining that Mission Control cannot start a durable mission
- **AND** it SHALL NOT fall through to ordinary shared-chat execution

#### Scenario: Slash submit still passes through without mission service
- **WHEN** the Mission Control composer contains a slash command
- **AND** no mission service is configured
- **THEN** the command SHALL still pass through to the shared chat path
