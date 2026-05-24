## MODIFIED Requirements
### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: List-mode detail help appears only when a task row exists
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** one or more task rows exist
- **THEN** the help bar SHALL advertise `Enter` for task detail toggling

#### Scenario: List-mode detail help hides inert Enter in empty state
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** zero task rows exist
- **THEN** the help bar SHALL omit `Enter`
