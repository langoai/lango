## ADDED Requirements

### Requirement: Running-state follow-up guidance preserves direct editing

The standalone workbench SHALL treat staged follow-up editing as part of the running-state interaction loop, not as a separate hidden behavior.

#### Scenario: Running-state follow-up editing returns to Composer
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already running
- **AND** the operator has staged a follow-up prompt
- **AND** the operator presses an editing key while focus is away from `Composer`
- **THEN** the workbench SHALL return focus to `Composer`
- **AND** SHALL apply the edit to the staged follow-up prompt
