## MODIFIED Requirements

### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Detail mode closes when refresh leaves no selected task
- **WHEN** the Tasks page detail panel is open
- **AND** a refresh leaves the task list empty
- **THEN** the page SHALL exit detail mode
- **AND** the empty-state help SHALL no longer advertise detail-only actions
