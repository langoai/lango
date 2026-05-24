## ADDED Requirements

### Requirement: Ready-profile workbench empty state suggests starter prompts
The standalone workbench empty state SHALL suggest concrete starter prompts when the active profile is ready for normal use.

#### Scenario: Ready profile shows starter prompts
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **THEN** the empty state SHALL mention starter prompts for repository summary, project-structure explanation, and recent-change review
- **AND** SHALL keep the existing chat and cockpit navigation hints
