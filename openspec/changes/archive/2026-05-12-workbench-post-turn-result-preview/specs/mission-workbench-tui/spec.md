## ADDED Requirements

### Requirement: Completed-turn empty workbench previews the last result

The standalone workbench SHALL surface the latest assistant summary directly in the completed-turn empty body so the operator can see what just happened without scanning the activity lane.

#### Scenario: Completed-turn body shows last result preview
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** a compact assistant activity summary is available
- **THEN** the empty body SHALL include that latest result as a compact last-result preview
