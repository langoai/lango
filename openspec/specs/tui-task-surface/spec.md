## ADDED Requirements

### Requirement: Task strip in chat view
The chat view SHALL include a 1-2 line task strip above the footer displaying a summary of active background tasks when a BackgroundManager is available.

#### Scenario: Task strip with running tasks
- **WHEN** BackgroundManager reports 2 running tasks
- **THEN** the strip displays: `Tasks: 2 running | [task-name] 2m12s`

#### Scenario: Task strip hidden when no manager
- **WHEN** BackgroundManager is nil
- **THEN** the task strip renders as empty string (zero height)

#### Scenario: Task strip hidden when no active tasks
- **WHEN** BackgroundManager reports 0 tasks
- **THEN** the task strip renders as empty string (zero height)

#### Scenario: Task strip refreshes periodically
- **WHEN** a `TaskStripTickMsg` arrives every 2 seconds
- **THEN** the task strip re-queries the manager and updates its display

### Requirement: Tasks cockpit page
The cockpit SHALL include a Tasks page (PageTasks) showing a table of all background tasks with status, elapsed time, and prompt preview.

#### Scenario: Tasks page with active tasks
- **WHEN** Tasks page is activated and BackgroundManager has 3 tasks
- **THEN** a table displays each task with: ID (truncated), prompt (truncated), status, elapsed time

#### Scenario: Tasks page with nil manager
- **WHEN** Tasks page is activated and BackgroundManager is nil
- **THEN** the page displays "No background tasks available"

#### Scenario: Tasks page lifecycle
- **WHEN** Tasks page is activated via Activate()
- **THEN** a 2-second refresh tick starts; when deactivated via Deactivate(), the tick stops

### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Cursor navigation
- **WHEN** user presses `↓` on the Tasks page
- **THEN** the cursor moves to the next task row

### Requirement: Tasks page cockpit registration
The cockpit SHALL register the Tasks page at `PageTasks` (ID 5) with keyboard shortcut Ctrl+5 and a sidebar menu entry.

#### Scenario: Ctrl+5 switches to Tasks
- **WHEN** user presses Ctrl+5
- **THEN** the cockpit switches to the Tasks page

#### Scenario: Sidebar shows Tasks entry
- **WHEN** the sidebar is rendered
- **THEN** a "Tasks" menu item is visible and clickable
