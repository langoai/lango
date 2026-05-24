## MODIFIED Requirements
### Requirement: Mission Control proposed-mission actions fail closed when backing services are absent
The cockpit Mission Control page SHALL return explicit system messages when the operator triggers proposed-mission actions that require unavailable backing services.

#### Scenario: Proposal row help exposes accept and dismiss keys
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** the missions lane is focused
- **THEN** the help bar SHALL label `Enter` as accepting the proposal
- **AND** it SHALL include the `d` dismiss binding

#### Scenario: Ordinary mission row hides inert Enter help
- **WHEN** the focused Mission Control lane has an ordinary mission row selected
- **AND** no starter-flow or composer submit path is active
- **THEN** the help bar SHALL omit the generic `Enter` binding

#### Scenario: True empty cockpit hides inert Enter help
- **WHEN** Mission Control is in the true empty state on the ordinary cockpit surface
- **AND** no starter-flow path is available
- **THEN** the help bar SHALL omit the generic `Enter` binding

#### Scenario: Accept proposed mission fails closed without mission service
- **WHEN** the operator accepts a selected proposed mission
- **AND** no mission service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot accept the proposal into a durable mission row

#### Scenario: Dismiss proposed mission fails closed without proposal service
- **WHEN** the operator dismisses a selected proposed mission
- **AND** no proposal service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot dismiss the proposal
