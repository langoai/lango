## ADDED Requirements

### Requirement: Post-turn empty workbench copy names the next-step loop explicitly

The standalone workbench SHALL describe the completed-turn empty state as a next-step interaction loop rather than reusing first-run quick-start wording.

#### Scenario: Completed turn uses next-step copy
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state body or footer SHALL describe the default `Enter` prompt as the next step instead of the initial quick start
