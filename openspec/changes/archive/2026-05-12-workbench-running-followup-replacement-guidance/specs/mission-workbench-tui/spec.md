## ADDED Requirements

### Requirement: Running-state follow-up guidance preserves starter replacement

The standalone workbench SHALL treat starter replacement as part of the running follow-up loop while a starter turn is still in flight.

#### Scenario: Running follow-up accepts starter replacement
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up prompt
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the staged follow-up SHALL switch to the corresponding starter prompt
