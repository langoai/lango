## ADDED Requirements

### Requirement: Ready-profile workbench starter prompts are keyboard-addressable

The standalone workbench SHALL let the operator load the ready-profile starter prompts through direct keyboard shortcuts instead of forcing retyping from the empty-state copy.

#### Scenario: Starter hotkeys load prompts into the empty composer
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile and an empty composer
- **AND** the operator presses `1`, `2`, or `3`
- **THEN** the corresponding starter prompt SHALL be loaded into the composer
- **AND** the operator SHALL remain in control of whether to press `Enter` to run it

#### Scenario: Starter hotkeys are advertised in ready-profile workbench copy
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile and an empty composer
- **THEN** the empty-state body SHALL mention the `1`, `2`, and `3` starter-prompt hotkeys
- **AND** the empty composer placeholder SHALL mention the `Press 1-3` quick-start path
- **AND** the footer SHALL expose the starter-prompt hotkeys while that state is active
