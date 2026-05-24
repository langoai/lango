## ADDED Requirements

### Requirement: Seeded starter prompts switch body copy to submit guidance

The standalone workbench SHALL replace the ready-profile quick-start body copy with submit guidance once a starter prompt is armed in the composer.

#### Scenario: Seeded starter body shows submit guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt has been seeded into the composer
- **THEN** the empty-state body SHALL explain that `Enter` runs the starter prompt or that the operator may edit it before sending
- **AND** it SHALL stop showing the pre-seed quick-start line for that state
