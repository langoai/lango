## ADDED Requirements

### Requirement: Completed-turn empty body presents the primary next step before the typing hint

The standalone workbench SHALL present the completed-turn next-step starter guidance before the generic next-prompt typing hint so the recommended action reads first.

#### Scenario: Completed-turn body orders starter before typing hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the completed-turn next-step starter guidance SHALL appear before the generic next-prompt hint in the empty body
