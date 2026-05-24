## MODIFIED Requirements
### Requirement: Mission Control proposed-mission actions fail closed when backing services are absent
The cockpit Mission Control page SHALL return explicit system messages when the operator triggers proposed-mission actions that require unavailable backing services.

#### Scenario: Proposal row help exposes accept and dismiss keys
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** the missions lane is focused
- **THEN** the help bar SHALL label `Enter` as accepting the proposal
- **AND** it SHALL include the `d` dismiss binding

#### Scenario: Proposal row help preserves navigation when another row exists
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** the missions lane is focused
- **AND** another mission row exists
- **THEN** the help bar SHALL continue exposing the `↑/↓` navigation bindings
- **AND** SHALL keep the `Enter` accept and `d` dismiss bindings visible at the same time
