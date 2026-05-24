## Purpose

Capability spec for tui-task-surface. See requirements below for scope and behavior contracts.
## Requirements
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

### Requirement: Task info data model
The TaskInfo struct SHALL include ID, Prompt, Status, Elapsed, Result, Error, OriginChannel, and TokensUsed fields. The bgTaskLister adapter SHALL populate all fields from the background.TaskSnapshot.

#### Scenario: TaskInfo includes result and error
- **WHEN** ListTasks() is called
- **THEN** each TaskInfo includes Result and Error from the snapshot

#### Scenario: TaskInfo includes origin and tokens
- **WHEN** ListTasks() is called
- **THEN** each TaskInfo includes OriginChannel and TokensUsed

### Requirement: TaskActioner interface
The pages package SHALL define a TaskActioner interface with CancelTask(id) and RetryTask(ctx, id) methods for TUI-initiated background task actions.

#### Scenario: Cancel delegates to manager
- **WHEN** CancelTask is called
- **THEN** it delegates to background.Manager.Cancel

#### Scenario: Retry resubmits with original origin
- **WHEN** RetryTask is called
- **THEN** it fetches the original task snapshot and calls Manager.Submit with the same Prompt, OriginChannel, and OriginSession

### Requirement: Tasks cockpit page
The cockpit SHALL include a Tasks page (PageTasks) showing a table of all background tasks with status, elapsed time, and prompt preview.

#### Scenario: Tasks page with nil manager
- **WHEN** Tasks page is activated and `BackgroundManager` is nil
- **THEN** the page SHALL explain that the background task manager is not configured

#### Scenario: Rendered tasks-page text stays plain and single-line
- **WHEN** task IDs, prompt previews, statuses, origin labels, result text, error text, or transient task-action messages contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Tasks page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
- **AND** wrapped detail content SHALL be derived from that sanitized single-line text

### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Cursor navigation
- **WHEN** user presses `↓` on the Tasks page outside detail mode
- **THEN** the cursor moves to the next task row

#### Scenario: Detail mode help labels the current Esc action accurately
- **WHEN** the Tasks page detail panel is open
- **THEN** the `Esc` help label SHALL describe closing the detail panel rather than a generic back action

#### Scenario: Detail mode shows scroll help only when overflow exists
- **WHEN** the Tasks page detail panel is open
- **AND** the selected task detail content exceeds the visible detail panel height
- **THEN** the help bar SHALL advertise `↑/k` and `↓/j` as scroll actions

#### Scenario: Detail mode hides inert scroll help when no overflow exists
- **WHEN** the Tasks page detail panel is open
- **AND** the selected task detail content fits within the visible detail panel height
- **THEN** the help bar SHALL omit `↑/k` and `↓/j`

#### Scenario: List-mode detail help appears only when a task row exists
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** one or more task rows exist
- **THEN** the help bar SHALL advertise `Enter` for task detail toggling

#### Scenario: List-mode detail help hides inert Enter in empty state
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** zero task rows exist
- **THEN** the help bar SHALL omit `Enter`

#### Scenario: List-mode navigation help appears only when another task row exists
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** two or more task rows exist
- **THEN** the help bar SHALL advertise `↑/k` and `↓/j`

#### Scenario: List-mode navigation help hides inert keys in zero-or-one-row states
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** fewer than two task rows exist
- **THEN** the help bar SHALL omit `↑/k` and `↓/j`

### Requirement: Tasks page cockpit registration
The cockpit SHALL register the Tasks page at `PageTasks` (ID 5) with keyboard shortcut Ctrl+5 and a sidebar menu entry.

#### Scenario: Ctrl+5 switches to Tasks
- **WHEN** user presses Ctrl+5
- **THEN** the cockpit switches to the Tasks page

#### Scenario: Sidebar shows Tasks entry
- **WHEN** the sidebar is rendered
- **THEN** a "Tasks" menu item is visible and clickable

### Requirement: Tasks page distinguishes unavailable from empty task state
The cockpit Tasks page SHALL distinguish between a missing background-task manager and an empty configured task list.

#### Scenario: Nil task manager renders unavailable message
- **WHEN** the Tasks page renders with no configured background-task lister
- **THEN** the page SHALL explain that the background task manager is not configured

#### Scenario: Empty configured task list renders no-tasks message
- **WHEN** the Tasks page renders with a configured task lister that returns zero tasks
- **THEN** the page SHALL display `No active tasks`

### Requirement: Tasks page help reflects the selected task's valid actions
The cockpit Tasks page SHALL expose cancel/retry help only when the currently selected task state supports that action.

#### Scenario: Running or pending task shows cancel help
- **WHEN** the selected task status is `running` or `pending`
- **THEN** the Tasks page help SHALL include the `c` cancel binding
- **AND** it SHALL NOT include retry-only help for that row

#### Scenario: Failed or cancelled task shows retry help
- **WHEN** the selected task status is `failed` or `cancelled`
- **THEN** the Tasks page help SHALL include the `r` retry binding
- **AND** it SHALL NOT include cancel-only help for that row

#### Scenario: Non-actionable task hides task action help
- **WHEN** the selected task status does not support cancel or retry
- **THEN** the Tasks page help SHALL omit both action bindings

#### Scenario: Task action failures surface explicit status wording
- **WHEN** a cancel or retry action fails
- **THEN** the Tasks page SHALL show a transient status message explaining that the task action failed
- **AND** the message SHALL include the underlying error text
