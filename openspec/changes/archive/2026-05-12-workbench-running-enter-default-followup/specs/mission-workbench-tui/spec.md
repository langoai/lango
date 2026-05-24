## MODIFIED Requirements

### Requirement: Running-state Enter can queue the default follow-up when no draft exists

The standalone workbench SHALL let the operator use `Enter` as the zero-input follow-up path while a starter turn is still running.

#### Scenario: Running-state Enter queues the default follow-up
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** no follow-up draft has been staged yet
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL queue and run the default context-aware follow-up prompt as the next turn
