## ADDED Requirements

### Requirement: Running follow-up starter replacement changes the next turn that runs

The standalone workbench SHALL treat starter replacement during a running follow-up loop as a change to the actual next turn that will execute.

#### Scenario: Replaced follow-up starter becomes the next executed prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up prompt
- **AND** the operator replaces that follow-up with `1`, `2`, or `3`
- **AND** the operator presses `Enter`
- **THEN** the replacement starter prompt SHALL be the next turn that runs
