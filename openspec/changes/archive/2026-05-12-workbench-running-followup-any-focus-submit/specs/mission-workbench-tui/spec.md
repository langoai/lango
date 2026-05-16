## ADDED Requirements

### Requirement: Running follow-up drafts submit from any empty-workbench focus lane

The standalone workbench SHALL let the operator submit a staged follow-up draft with `Enter` even if focus has moved away from `Composer`, as long as the workbench is still in the empty ready-profile running state.

#### Scenario: Staged follow-up submits from Decisions focus
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up draft
- **AND** focus is on `Decisions`
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL queue and run the staged follow-up as the next turn
