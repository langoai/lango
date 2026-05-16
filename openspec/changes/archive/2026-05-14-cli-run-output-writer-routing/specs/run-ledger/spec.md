## MODIFIED Requirements

### Requirement: CLI Journal Inspection
The system SHALL let operators inspect persistent RunLedger data from the CLI.

#### Scenario: RunLedger CLI output uses the command writer
- **WHEN** `lango run list`, `lango run status`, or `lango run journal <run-id>` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
