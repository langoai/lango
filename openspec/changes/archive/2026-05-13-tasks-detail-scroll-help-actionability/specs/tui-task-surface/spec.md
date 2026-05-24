## MODIFIED Requirements
### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Detail mode shows scroll help only when overflow exists
- **WHEN** the Tasks page detail panel is open
- **AND** the selected task detail content exceeds the visible detail panel height
- **THEN** the help bar SHALL advertise `↑/k` and `↓/j` as scroll actions

#### Scenario: Detail mode hides inert scroll help when no overflow exists
- **WHEN** the Tasks page detail panel is open
- **AND** the selected task detail content fits within the visible detail panel height
- **THEN** the help bar SHALL omit `↑/k` and `↓/j`
