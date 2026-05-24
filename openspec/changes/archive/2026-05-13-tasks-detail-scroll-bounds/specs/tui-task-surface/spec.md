## MODIFIED Requirements
### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Cursor navigation
- **WHEN** user presses `↓` on the Tasks page outside detail mode
- **THEN** the cursor moves to the next task row

#### Scenario: Detail scrolling is bounded to content
- **WHEN** the Tasks page detail panel is open
- **AND** the operator repeatedly presses `↓`
- **THEN** the detail scroll offset SHALL stop at the last meaningful content offset rather than increasing indefinitely
