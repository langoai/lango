## ADDED Requirements

### Requirement: Workbench composer placeholder follows profile readiness
The standalone workbench composer placeholder SHALL mirror the same readiness split as the workbench empty-state body when the composer is empty.

#### Scenario: Incomplete profile shows setup-first composer hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with an incomplete profile and the composer is empty
- **THEN** the composer placeholder SHALL instruct the operator to use `lango onboard`, `lango settings`, or `lango doctor`

#### Scenario: Ready profile shows starter-prompt composer hint
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile and the composer is empty
- **THEN** the composer placeholder SHALL suggest starter prompts for repository summary, project-structure explanation, and recent-change review
