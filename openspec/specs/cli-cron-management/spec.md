## Purpose

Define the CLI commands for managing cron job scheduling (add, list, delete, pause, resume, history).

## Requirements

### Requirement: Cron command output routing
The CLI SHALL route human-readable `lango cron` output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Cron output uses the command writer
- **WHEN** `lango cron` renders confirmation text or tabular output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: Cron add command
The CLI SHALL provide `lango cron add` with flags: --name (required), --schedule/--every/--at (mutually exclusive, one required), --prompt (required), --deliver (repeatable), --isolated, --timezone.

#### Scenario: Add cron job with cron expression
- **WHEN** user runs `lango cron add --name "news" --schedule "0 9 * * *" --prompt "Summarize news" --deliver slack`
- **THEN** the CLI SHALL create a cron job with schedule_type "cron" and deliver_to ["slack"]

#### Scenario: Add cron job with interval
- **WHEN** user runs `lango cron add --name "check" --every 1h --prompt "Check servers"`
- **THEN** the CLI SHALL create a cron job with schedule_type "every"

#### Scenario: Add cron job with one-time trigger
- **WHEN** user runs `lango cron add --name "meeting" --at "2026-02-20T15:00:00" --prompt "Prepare meeting notes"`
- **THEN** the CLI SHALL create a cron job with schedule_type "at"

### Requirement: Cron list command
The CLI SHALL provide `lango cron list` that displays all jobs in a table with columns: ID, Name, Schedule, Enabled, Last Run, Next Run.

#### Scenario: List cron jobs
- **WHEN** user runs `lango cron list`
- **THEN** the CLI SHALL display all cron jobs in tabular format

#### Scenario: No cron jobs
- **WHEN** user runs `lango cron list` with no jobs configured
- **THEN** the CLI SHALL display "No cron jobs found."

### Requirement: Cron delete command
The CLI SHALL provide `lango cron delete <id-or-name>` that removes a job by UUID or name.

#### Scenario: Delete by name
- **WHEN** user runs `lango cron delete news`
- **THEN** the CLI SHALL look up the job by name and delete it

### Requirement: Cron pause and resume commands
The CLI SHALL provide `lango cron pause <id-or-name>` and `lango cron resume <id-or-name>`.

#### Scenario: Pause a job
- **WHEN** user runs `lango cron pause news`
- **THEN** the CLI SHALL disable the job in the database

#### Scenario: Resume a job
- **WHEN** user runs `lango cron resume news`
- **THEN** the CLI SHALL enable the job in the database

### Requirement: Cron history command
The CLI SHALL provide `lango cron history [id-or-name]` that displays execution history with columns: Run ID, Job, Status, Started, Completed, Tokens.

#### Scenario: View history for a specific job
- **WHEN** user runs `lango cron history news`
- **THEN** the CLI SHALL display execution history entries for the "news" job

#### Scenario: View all history
- **WHEN** user runs `lango cron history` without arguments
- **THEN** the CLI SHALL display recent execution history for all jobs

### Requirement: Job ID resolution
The CLI SHALL support both UUID and job name for identifying jobs. When a non-UUID string is provided, the CLI SHALL look up the job by name.

#### Scenario: Resolve job by name
- **WHEN** a command receives "news" as the job identifier
- **THEN** the CLI SHALL query the database for a job with name "news" and use its UUID
