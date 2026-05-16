## ADDED Requirements

### Requirement: Seeded starter prompts advertise the submit step

The standalone workbench SHALL distinguish between the starter-seeding step and the starter-submission step in its quick-start copy.

#### Scenario: Seeded starter shows submit-focused footer hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt has already been seeded into the composer
- **THEN** the footer hint SHALL indicate that `Enter` submits the starter prompt
- **AND** it SHALL stop showing the pre-seed `Enter default starter` hint for that state
