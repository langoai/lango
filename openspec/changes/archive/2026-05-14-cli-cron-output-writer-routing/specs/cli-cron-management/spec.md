## MODIFIED Requirements

### Requirement: Cron management commands
The CLI SHALL provide `lango cron add`, `list`, `delete`, `pause`, `resume`, and `history` commands for managing scheduled cron jobs through the configured cron store.

#### Scenario: Cron command output uses the command writer
- **WHEN** any human-readable `lango cron` subcommand renders confirmation text or tabular output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
