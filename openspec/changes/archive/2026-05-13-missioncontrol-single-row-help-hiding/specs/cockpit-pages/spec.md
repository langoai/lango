## MODIFIED Requirements
### Requirement: Mission Control surfaces unavailable projector state explicitly
The cockpit Mission Control page SHALL distinguish an unavailable mission-control projector from a normal empty mission state.

#### Scenario: Nil projector renders degraded note
- **WHEN** Mission Control renders after activation with no configured projector
- **THEN** the page SHALL still render its empty-state shell
- **AND** it SHALL include a degraded note explaining that Mission Control data is not configured

#### Scenario: Empty Mission Control help hides inert navigation keys
- **WHEN** Mission Control is in its true empty state with no missions, no pending decision, and no loops
- **THEN** the page help SHALL omit `↑/k` and `↓/j`
- **AND** it SHALL still advertise the focus and submit keys that remain actionable

#### Scenario: Populated Mission Control help hides inert navigation in single-row focus states
- **WHEN** Mission Control has content but the currently focused lane has fewer than two navigable rows
- **THEN** the page help SHALL omit `↑/k` and `↓/j`
