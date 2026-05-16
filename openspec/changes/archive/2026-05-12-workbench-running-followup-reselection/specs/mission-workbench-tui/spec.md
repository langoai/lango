## ADDED Requirements

### Requirement: Running follow-up drafts remain replaceable by starter hotkeys

The standalone workbench SHALL let the operator replace a staged follow-up draft with `1`, `2`, or `3` while the current starter turn is still running.

#### Scenario: Running follow-up hotkey replacement
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up draft
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the staged follow-up SHALL switch to the corresponding starter prompt
- **AND** the digit SHALL NOT be appended as free text
