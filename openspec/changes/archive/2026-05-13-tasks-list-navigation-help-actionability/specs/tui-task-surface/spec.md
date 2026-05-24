## MODIFIED Requirements
### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: List-mode navigation help appears only when another task row exists
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** two or more task rows exist
- **THEN** the help bar SHALL advertise `↑/k` and `↓/j`

#### Scenario: List-mode navigation help hides inert keys in zero-or-one-row states
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** fewer than two task rows exist
- **THEN** the help bar SHALL omit `↑/k` and `↓/j`
