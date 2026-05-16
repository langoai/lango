## ADDED Requirements

### Requirement: Ready-profile workbench Enter key seeds the default starter prompt

The standalone workbench SHALL let the operator use `Enter` as the default quick-start key on an empty ready-profile first screen.

#### Scenario: Enter seeds the first starter prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the composer is empty
- **AND** the operator presses `Enter`
- **THEN** the first starter prompt SHALL be loaded into the composer
- **AND** the operator SHALL remain in control of whether to press `Enter` again to submit it

#### Scenario: Incomplete profile does not seed a starter prompt on Enter
- **WHEN** bare `lango` renders an empty Mission Control workbench state with an incomplete profile
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL keep the setup-first guidance instead of seeding a starter prompt
