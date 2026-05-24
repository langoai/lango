## ADDED Requirements

### Requirement: Armed starter submission works from Decisions focus too

The standalone workbench SHALL honor the seeded-starter any-focus submit contract specifically from the `Decisions` lane as well as the other empty-workbench lanes.

#### Scenario: Armed starter submits from Decisions focus
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already armed in the composer
- **AND** focus is on `Decisions`
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL submit the armed starter prompt
