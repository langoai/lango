## ADDED Requirements

### Requirement: Completed-turn empty composer placeholder uses next-step wording

The standalone workbench SHALL keep its completed-turn composer placeholder aligned with the next-step body and footer wording.

#### Scenario: Completed-turn placeholder says next step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty composer placeholder SHALL say `Next step: press Enter ...` instead of the original first-run wording
