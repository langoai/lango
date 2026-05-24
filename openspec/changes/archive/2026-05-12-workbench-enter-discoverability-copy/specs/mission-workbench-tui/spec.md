## ADDED Requirements

### Requirement: Ready-profile workbench copy advertises the Enter quick-start path

The standalone workbench SHALL explicitly advertise `Enter` as the default quick-start seed on the empty ready-profile first screen.

#### Scenario: Ready-profile copy mentions Enter and numeric hotkeys
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **THEN** the empty-state body SHALL mention `Enter` for the default starter prompt
- **AND** SHALL continue mentioning `1`, `2`, and `3` for explicit starter selection
- **AND** the footer or empty composer hint SHALL surface the same `Enter` quick-start path
