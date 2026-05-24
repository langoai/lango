## ADDED Requirements

### Requirement: Completed-turn empty body uses neutral finished wording

The standalone workbench SHALL describe the completed-turn empty state with neutral `finished` wording rather than success-specific `complete` wording.

#### Scenario: Completed-turn body says finished
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state body SHALL say that the last turn finished rather than that it completed
