## MODIFIED Requirements
### Requirement: Sessions page with session list
The cockpit SHALL include a Sessions page showing session key and relative last update time, ordered newest-first by `UpdatedAt`.

#### Scenario: Sessions are sorted newest-first after load
- **WHEN** the Sessions page loads multiple session summaries with different `UpdatedAt` values
- **THEN** it SHALL render the most recently updated session first
- **AND** older sessions SHALL appear below newer ones

#### Scenario: Sessions help appears only when rows are navigable
- **WHEN** the Sessions page help is rendered with one or more loaded session rows
- **THEN** it SHALL advertise `↑/k` and `↓/j`

#### Scenario: Sessions help hides inert navigation in empty or unavailable states
- **WHEN** the Sessions page help is rendered with no loaded session rows
- **THEN** it SHALL omit vertical navigation bindings
