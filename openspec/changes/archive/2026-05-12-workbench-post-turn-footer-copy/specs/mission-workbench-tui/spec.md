## ADDED Requirements

### Requirement: Completed-turn workbench footer uses next-prompt wording

The standalone workbench SHALL keep its completed-turn footer aligned with the next-step body and placeholder wording.

#### Scenario: Completed-turn footer says next prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the footer SHALL tell the operator to type the next prompt here instead of generic chat wording
