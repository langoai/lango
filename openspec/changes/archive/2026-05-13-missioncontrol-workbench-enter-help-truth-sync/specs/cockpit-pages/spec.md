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

#### Scenario: Empty workbench help labels starter seeding explicitly
- **WHEN** the standalone workbench surface is in the true empty state
- **AND** the default starter prompt is not yet staged
- **THEN** the `Enter` help label SHALL describe seeding the starter prompt

#### Scenario: Empty workbench help labels starter submission explicitly
- **WHEN** the standalone workbench surface is in the true empty state
- **AND** a starter prompt is already staged in the composer
- **THEN** the `Enter` help label SHALL describe running the starter prompt
