## ADDED Requirements

### Requirement: Armed starter prompts remain replaceable by numeric starter shortcuts

The standalone workbench SHALL let the operator replace an already armed starter prompt with `1`, `2`, or `3` instead of treating those keys as plain text input.

#### Scenario: Numeric starter shortcut replaces an armed starter prompt
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a starter prompt is already armed in the composer
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the composer SHALL switch to the corresponding starter prompt
- **AND** the keypress SHALL NOT append the digit as free text
