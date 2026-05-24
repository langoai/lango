## ADDED Requirements

### Requirement: Failed completed-turn workbench copy uses recovery wording

The standalone workbench SHALL switch its completed-turn copy from generic next-step wording to recovery-specific wording when the latest turn failed.

#### Scenario: Failed turn uses recovery starter copy
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **THEN** the body or placeholder SHALL describe the default `Enter` starter as a recovery step
- **AND** the footer SHALL describe `Enter` as a recovery starter
