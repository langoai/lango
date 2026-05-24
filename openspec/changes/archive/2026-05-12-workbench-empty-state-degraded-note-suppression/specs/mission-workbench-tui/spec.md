## ADDED Requirements

### Requirement: Empty standalone workbench suppresses degraded-note noise

The standalone workbench SHALL avoid surfacing cockpit-style degraded warnings on its empty first screen when no active mission/control content exists yet.

#### Scenario: Empty workbench hides degraded note
- **WHEN** bare `lango` renders an empty Mission Control workbench state
- **AND** the projected header contains a degraded note
- **THEN** the empty workbench SHALL hide that degraded note from the first-screen shell

#### Scenario: Cockpit surface still shows degraded note
- **WHEN** the explicit `lango cockpit` Mission Control page renders an empty state with a degraded note
- **THEN** the degraded note SHALL remain visible
