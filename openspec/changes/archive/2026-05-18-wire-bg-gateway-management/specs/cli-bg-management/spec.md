## MODIFIED Requirements

### Requirement: Background list command
The CLI SHALL provide `lango bg list` that displays all background tasks with columns: ID, Status, Prompt (truncated), Started, Duration.

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

#### Scenario: Background CLI supports JSON output
- **WHEN** user runs `lango bg list`, `status`, `cancel`, or `result` with `--output json`
- **THEN** the CLI SHALL write the command result as JSON through the Cobra command output writer

### Requirement: Background status command
The CLI SHALL provide `lango bg status <id>` that displays detailed task information.

#### Scenario: View task status
- **WHEN** user runs `lango bg status <uuid>`
- **THEN** the CLI SHALL display the task's full details including status, prompt, origin, timing, and result/error fields when present

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

### Requirement: Background CLI client boundary
The bg CLI commands SHALL operate through a narrow background task client interface that supports both in-process manager adapters and gateway-backed remote adapters.

#### Scenario: Embedded in-process manager remains supported
- **WHEN** `internal/cli/bg.NewBgCmd` is constructed with an in-process background manager adapter
- **THEN** `list`, `status`, `cancel`, and `result` SHALL continue to operate on that in-process manager

#### Scenario: Gateway-backed root CLI
- **WHEN** root `lango bg` is constructed by `cmd/lango/main.go`
- **THEN** it SHALL use a gateway-backed background client
- **AND** it SHALL resolve the gateway address from `--addr` when supplied
- **AND** it SHALL otherwise resolve the address from configured server host and port

#### Scenario: Gateway unavailable
- **WHEN** the gateway-backed bg client cannot connect to the configured gateway
- **THEN** the CLI SHALL return an actionable connection error that includes gateway context
