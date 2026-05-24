## MODIFIED Requirements
### Requirement: Dead Letters remains registered as a cockpit degraded surface
The cockpit SHALL keep Dead Letters available as a registered page route even when the dead-letter bridge is unavailable, relying on page-level unavailable messaging instead of suppressing page registration.

#### Scenario: Dead Letters page remains registered without bridge callbacks
- **WHEN** cockpit startup wiring runs without a ready dead-letter bridge
- **THEN** the Dead Letters page route SHALL still be registered

#### Scenario: Dead Letters degraded page reports missing list function immediately
- **WHEN** the registered Dead Letters page renders without a configured list callback
- **THEN** the page SHALL explain that the dead-letter backlog is not configured

#### Scenario: Dead Letters activation still surfaces missing callback error
- **WHEN** the registered Dead Letters page is activated without a configured list callback
- **THEN** activation SHALL yield a load error that explains the dead-letter list function is not configured

#### Scenario: Row-navigation help appears only when another backlog row exists
- **WHEN** the Dead Letters help is rendered with two or more backlog rows
- **THEN** it SHALL advertise `↑/k` and `↓/j`

#### Scenario: Row-navigation help hides inert keys with zero or one backlog row
- **WHEN** the Dead Letters help is rendered with zero or one backlog row
- **THEN** it SHALL omit `↑/k` and `↓/j`

#### Scenario: Backspace help describes active text-filter editing
- **WHEN** the Dead Letters help is rendered
- **THEN** the `Backspace` binding SHALL describe editing the active text filter rather than only the query field

### Requirement: Mission Control surfaces unavailable projector state explicitly
The cockpit Mission Control page SHALL distinguish an unavailable mission-control projector from a normal empty mission state.

#### Scenario: Nil projector renders degraded note
- **WHEN** Mission Control renders after activation with no configured projector
- **THEN** the page SHALL still render its empty-state shell
- **AND** it SHALL include a degraded note explaining that Mission Control data is not configured

#### Scenario: Empty Mission Control help hides inert navigation keys
- **WHEN** Mission Control is in its true empty state with no missions, no pending decision, and no loops
- **THEN** the page help SHALL omit `↑/k` and `↓/j`
- **AND** it SHALL still advertise the focus and remaining actionable keys

#### Scenario: Empty workbench help labels starter seeding explicitly
- **WHEN** the standalone workbench surface is in the true empty state
- **AND** the default starter prompt is not yet staged
- **THEN** the `Enter` help label SHALL describe seeding the starter prompt

#### Scenario: Empty workbench help labels starter submission explicitly
- **WHEN** the standalone workbench surface is in the true empty state
- **AND** a starter prompt is already staged in the composer
- **THEN** the `Enter` help label SHALL describe running the starter prompt

#### Scenario: Populated Mission Control help hides inert navigation in single-row focus states
- **WHEN** Mission Control has content but the currently focused lane has fewer than two navigable rows
- **THEN** the page help SHALL omit `↑/k` and `↓/j`

#### Scenario: Composer/activity focus routes vertical navigation to the activity lane
- **WHEN** Mission Control focus is on the composer/activity lane
- **AND** multiple activity rows exist
- **THEN** pressing `↑/k` or `↓/j` SHALL move the activity cursor instead of being swallowed by the composer

### Requirement: Mission Control proposed-mission actions fail closed when backing services are absent
The cockpit Mission Control page SHALL return explicit system messages when the operator triggers proposed-mission actions that require unavailable backing services.

#### Scenario: Proposal row help exposes accept and dismiss keys
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** the missions lane is focused
- **THEN** the help bar SHALL label `Enter` as accepting the proposal
- **AND** it SHALL include the `d` dismiss binding

#### Scenario: Dismiss help hides while another lane is focused
- **WHEN** the selected Mission Control row is a proposed mission
- **AND** focus is not on the missions lane
- **THEN** the help bar SHALL omit the `d` dismiss binding

#### Scenario: Ordinary mission row hides inert Enter help
- **WHEN** the focused Mission Control lane has an ordinary mission row selected
- **AND** no starter-flow or composer submit path is active
- **THEN** the help bar SHALL omit the generic `Enter` binding

#### Scenario: Decisions lane help hides inert Enter key
- **WHEN** Mission Control focus is on the decisions lane
- **THEN** the help bar SHALL omit the generic `Enter` submit binding

#### Scenario: True empty cockpit hides inert Enter help
- **WHEN** Mission Control is in the true empty state on the ordinary cockpit surface
- **AND** no starter-flow path is available
- **THEN** the help bar SHALL omit the generic `Enter` binding

#### Scenario: Accept proposed mission fails closed without mission service
- **WHEN** the operator accepts a selected proposed mission
- **AND** no mission service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot accept the proposal into a durable mission row

#### Scenario: Dismiss proposed mission fails closed without proposal service
- **WHEN** the operator dismisses a selected proposed mission
- **AND** no proposal service is configured
- **THEN** the page SHALL emit a system message explaining that Mission Control cannot dismiss the proposal
