## ADDED Requirements

### Requirement: Seeded starter guidance reflects whether Composer still has focus

The standalone workbench SHALL tailor seeded-starter guidance to the current focus lane so it only tells the operator to press keys that will actually work next.

#### Scenario: Seeded starter outside composer explains the focus step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is armed in the composer
- **AND** focus is no longer on the composer lane
- **THEN** the guidance SHALL instruct the operator to return to `Composer` before pressing `Enter`
