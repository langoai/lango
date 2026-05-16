## MODIFIED Requirements
### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Cursor navigation
- **WHEN** user presses `↓` on the Tasks page outside detail mode
- **THEN** the cursor moves to the next task row

#### Scenario: Detail scroll is re-clamped after content shrink
- **WHEN** the Tasks page detail panel is open
- **AND** the selected task detail content shrinks after a refresh
- **THEN** the effective detail scroll offset SHALL be clamped back into the valid viewport range
- **AND** the page SHALL not render from below the last meaningful viewport offset
