## MODIFIED Requirements

### Requirement: Tasks cockpit page
The cockpit SHALL include a Tasks page (PageTasks) showing a table of all background tasks with status, elapsed time, and prompt preview.

#### Scenario: Tasks page with nil manager
- **WHEN** Tasks page is activated and `BackgroundManager` is nil
- **THEN** the page SHALL explain that the background task manager is not configured
