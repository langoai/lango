## ADDED Requirements

### Requirement: Task detail inline expansion
The Tasks page SHALL display a detail panel below the table when the user presses Enter on a selected task. The panel SHALL show full prompt, result, error, origin channel, tokens used, and timing.

#### Scenario: Open detail view
- **WHEN** user presses Enter on a selected task
- **THEN** a detail panel appears below the table showing all task fields

#### Scenario: Close detail view
- **WHEN** user presses Esc while detail view is open
- **THEN** the detail panel closes and cursor returns to list mode

#### Scenario: Scroll detail content
- **WHEN** detail view is open and user presses up/down
- **THEN** the detail content scrolls instead of moving the table cursor

#### Scenario: Height clamp
- **WHEN** terminal height is less than 14 rows
- **THEN** table shrinks to tableMinHeight=6 and detail gets remaining space

### Requirement: Task cancel action
The Tasks page SHALL allow cancelling pending or running tasks via the `c` key when a TaskActioner is available.

#### Scenario: Cancel running task
- **WHEN** user presses `c` on a task with status "running"
- **THEN** TaskActioner.CancelTask is called and a transient status message is shown

#### Scenario: Cancel ignored for done task
- **WHEN** user presses `c` on a task with status "done"
- **THEN** no action is taken

#### Scenario: Nil actioner
- **WHEN** no TaskActioner is available and user presses `c`
- **THEN** no action is taken and no panic occurs

### Requirement: Task retry action
The Tasks page SHALL allow retrying failed or cancelled tasks via the `r` key, resubmitting with the original prompt, origin channel, and origin session.

#### Scenario: Retry failed task
- **WHEN** user presses `r` on a task with status "failed"
- **THEN** TaskActioner.RetryTask is called with the original prompt and origin

#### Scenario: Retry ignored for running task
- **WHEN** user presses `r` on a task with status "running"
- **THEN** no action is taken

### Requirement: Action feedback display
The Tasks page SHALL display transient status messages after cancel/retry actions, auto-clearing after 3 seconds.

#### Scenario: Status message shown
- **WHEN** a cancel or retry action completes
- **THEN** a status message is displayed at the bottom of the page

#### Scenario: Status message cleared
- **WHEN** 3 seconds have elapsed since the status message was set
- **THEN** the message is cleared on the next tick
