## ADDED Requirements

### Requirement: Failed completed-turn body lead uses recovery-step wording

The standalone workbench SHALL keep the failed completed-turn body lead aligned with the rest of the recovery-specific copy.

#### Scenario: Failed turn lead says recovery step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the empty-state lead SHALL tell the operator to pick the recovery step instead of the generic next step
