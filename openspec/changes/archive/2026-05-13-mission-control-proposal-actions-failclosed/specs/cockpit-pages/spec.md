## ADDED Requirements

### Requirement: Mission Control proposed-mission actions fail closed when backing services are absent
The cockpit Mission Control page SHALL return explicit system messages when the operator triggers proposed-mission actions that require unavailable backing services.

#### Scenario: Accept proposed mission fails closed without mission service
- **WHEN** the operator accepts a selected proposed mission
- **AND** no mission service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot accept the proposal into a durable mission row

#### Scenario: Dismiss proposed mission fails closed without proposal service
- **WHEN** the operator dismisses a selected proposed mission
- **AND** no proposal service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot dismiss the proposal
