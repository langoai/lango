## Purpose

Define the CLI commands for managing background task execution (list, status, cancel, result).

## Requirements

### Requirement: Background list command
The CLI SHALL provide `lango bg list` that displays all background tasks with columns: ID, Status, Prompt (truncated), Started, Completed.

#### Scenario: List background tasks
- **WHEN** user runs `lango bg list`
- **THEN** the CLI SHALL display all background tasks in tabular format

#### Scenario: No background tasks
- **WHEN** user runs `lango bg list` with no active tasks
- **THEN** the CLI SHALL display "No background tasks."

#### Scenario: Background CLI output uses the command writer
- **WHEN** `lango bg list`, `status`, `cancel`, or `result` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: Background status command
The CLI SHALL provide `lango bg status <id>` that displays detailed task information.

#### Scenario: View task status
- **WHEN** user runs `lango bg status <uuid>`
- **THEN** the CLI SHALL display the task's full details including status, prompt, origin, timing, and tokens used

### Requirement: Background cancel command
The CLI SHALL provide `lango bg cancel <id>` that cancels a running task.

#### Scenario: Cancel a running task
- **WHEN** user runs `lango bg cancel <uuid>` for a running task
- **THEN** the CLI SHALL cancel the task and display confirmation

### Requirement: Background result command
The CLI SHALL provide `lango bg result <id>` that displays the result of a completed task.

#### Scenario: View completed task result
- **WHEN** user runs `lango bg result <uuid>` for a completed task
- **THEN** the CLI SHALL display the full result text

#### Scenario: View result of incomplete task
- **WHEN** user runs `lango bg result <uuid>` for a task that is not yet complete
- **THEN** the CLI SHALL display an error indicating the task is not done

### Requirement: In-memory manager dependency
The bg CLI commands SHALL require access to the running server's in-memory Manager instance via a provider function.

#### Scenario: Server not running
- **WHEN** bg commands are invoked without a running server providing the Manager
- **THEN** the CLI SHALL return an error indicating the background manager is not available
