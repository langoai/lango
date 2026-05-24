## ADDED Requirements

### Requirement: Completed-turn empty workbench hint invites the next prompt explicitly

The standalone workbench SHALL update its completed-turn empty-state hint so it explicitly invites the next prompt instead of generic chat wording.

#### Scenario: Completed-turn hint says next prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state hint SHALL tell the operator to type the next prompt here
