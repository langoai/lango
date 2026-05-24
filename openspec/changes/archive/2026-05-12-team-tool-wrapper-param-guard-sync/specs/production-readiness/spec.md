## ADDED Requirements

### Requirement: Team workflow wrapper guards stay actionable

Team workflow tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: Team workflow tools reject missing required inputs
- **WHEN** `team_form`, `team_form_with_budget`, or `team_complete_milestone` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream workflow execution begins
