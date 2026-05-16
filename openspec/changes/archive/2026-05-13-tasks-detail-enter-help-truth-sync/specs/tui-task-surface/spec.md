## MODIFIED Requirements
### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Cursor navigation
- **WHEN** user presses `↓` on the Tasks page outside detail mode
- **THEN** the cursor moves to the next task row

#### Scenario: Detail mode help labels the current Enter action accurately
- **WHEN** the Tasks page detail panel is open
- **THEN** the `Enter` help label SHALL describe closing the detail panel rather than opening details
