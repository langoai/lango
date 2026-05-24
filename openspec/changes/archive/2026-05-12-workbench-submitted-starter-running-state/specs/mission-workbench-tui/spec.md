## ADDED Requirements

### Requirement: Submitted starter prompts switch the empty workbench into a running-state hint

The standalone workbench SHALL replace its starter-oriented empty-state guidance with a running-state hint while a submitted starter prompt is still in flight.

#### Scenario: Submitted starter shows running-state guidance
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt has already been submitted
- **AND** the turn is still in flight
- **THEN** the empty-state body, footer, or placeholder SHALL indicate that the current request is running
- **AND** the workbench SHALL stop showing pre-submit quick-start guidance for that state
