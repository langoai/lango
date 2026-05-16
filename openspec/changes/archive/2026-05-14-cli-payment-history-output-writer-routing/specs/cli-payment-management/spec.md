## MODIFIED Requirements

### Requirement: History command
The system SHALL provide a `lango payment history` command that displays recent payment transactions in a table.

#### Scenario: Payment history output uses the command writer
- **WHEN** `lango payment history` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
