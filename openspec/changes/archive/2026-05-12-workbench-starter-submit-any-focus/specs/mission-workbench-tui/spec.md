## ADDED Requirements

### Requirement: Seeded starter prompts submit from any empty-workbench focus lane

The standalone workbench SHALL let the operator submit an armed starter prompt with `Enter` even if focus has moved away from `Composer`, as long as the workbench is still in the empty ready-profile seeded state.

#### Scenario: Seeded starter submits outside composer focus
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is armed in the composer
- **AND** focus has moved to a different lane
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL submit the armed starter prompt
- **AND** seeded-state guidance SHALL no longer require tabbing back to `Composer` first
