## MODIFIED Requirements

### Requirement: Doctor Command Entry Point
The system SHALL include RunLedger diagnostics in the `lango doctor` command output and help text.

#### Scenario: Doctor output uses the command writer
- **WHEN** `lango doctor` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
