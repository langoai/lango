## ADDED Requirements

### Requirement: Team workflow tools keep actionable wrapper parameter guards

Team workflow tools SHALL reject missing required wrapper inputs with actionable parameter errors before workflow orchestration begins.

#### Scenario: Team workflow tools reject missing required inputs
- **WHEN** `team_form`, `team_form_with_budget`, or `team_complete_milestone` is invoked without one of its declared required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not proceed into downstream coordination, escrow, or budget operations
