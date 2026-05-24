## ADDED Requirements

### Requirement: Completed-turn empty body calls out failed turns explicitly

The standalone workbench SHALL distinguish a failed prior turn from a successful one in its completed-turn empty body lead.

#### Scenario: Failed turn changes the completed-turn lead
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the empty-state lead SHALL indicate that the last turn needs attention instead of using the neutral finished wording
