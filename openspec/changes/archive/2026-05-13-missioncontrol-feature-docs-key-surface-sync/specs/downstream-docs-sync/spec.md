## ADDED Requirements
### Requirement: Cockpit feature docs describe the Mission Control key surface
The public cockpit feature reference SHALL describe the current Mission Control key surface beyond the general page overview.

#### Scenario: Cockpit feature page includes Mission Control keys subsection
- **WHEN** `docs/features/cockpit.md` documents Mission Control
- **THEN** it SHALL include a dedicated key-surface subsection
- **AND** it SHALL describe the `tab`/`enter` core actions, populated-state `↑/↓` navigation, and the reduced help surface in the true empty state
