## ADDED Requirements

### Requirement: Armed starter editing keys return focus to Composer

The standalone workbench SHALL treat composer editing keys as intent to edit the armed starter prompt even if focus has moved away from `Composer`.

#### Scenario: Armed starter backspace returns focus to Composer
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is armed in the composer
- **AND** focus has moved away from `Composer`
- **AND** the operator presses a composer editing key such as `Backspace`
- **THEN** the workbench SHALL return focus to `Composer`
- **AND** SHALL apply the edit to the armed starter prompt
