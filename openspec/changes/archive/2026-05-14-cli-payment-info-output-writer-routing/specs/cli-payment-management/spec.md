## MODIFIED Requirements

### Requirement: Info command
The system SHALL provide a `lango payment info` command that displays wallet and payment system configuration.

#### Scenario: Payment info output uses the command writer
- **WHEN** `lango payment info` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
