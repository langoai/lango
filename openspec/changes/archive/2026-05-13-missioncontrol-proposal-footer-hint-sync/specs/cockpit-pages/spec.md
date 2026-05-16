## MODIFIED Requirements

### Requirement: Mission Control proposed-mission actions fail closed when backing services are absent
The cockpit Mission Control page SHALL return explicit system messages when the operator triggers proposed-mission actions that require unavailable backing services.

#### Scenario: Proposal-focused footer hint advertises accept and dismiss actions
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** the missions lane is focused
- **THEN** the footer hint SHALL mention `Enter` for accepting the proposal
- **AND** SHALL mention `d` for dismissing the proposal
