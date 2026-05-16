## MODIFIED Requirements

### Requirement: Limits command
The system SHALL provide a `lango payment limits` command that displays spending limits and daily usage.

#### Scenario: Payment limits output uses the command writer
- **WHEN** `lango payment limits` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
