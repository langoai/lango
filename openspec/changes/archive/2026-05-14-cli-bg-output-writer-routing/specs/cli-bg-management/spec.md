## MODIFIED Requirements

### Requirement: Background list command
The CLI SHALL provide `lango bg list` that displays all background tasks with columns: ID, Status, Prompt (truncated), Started, Completed.

#### Scenario: Background CLI output uses the command writer
- **WHEN** `lango bg list`, `status`, `cancel`, or `result` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
