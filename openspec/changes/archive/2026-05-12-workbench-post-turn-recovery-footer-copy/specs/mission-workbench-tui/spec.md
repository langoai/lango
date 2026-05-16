## ADDED Requirements

### Requirement: Failed completed-turn footer uses recovery-prompt wording

The standalone workbench SHALL keep its failed completed-turn footer aligned with the recovery-specific body and placeholder wording.

#### Scenario: Failed turn footer says recovery prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the footer SHALL say `Type recovery prompt here` instead of `Type next prompt here`
