## MODIFIED Requirements

### Requirement: Balance command
The system SHALL provide a `lango payment balance` command that displays the wallet's USDC balance, address, and network.

#### Scenario: Payment balance output uses the command writer
- **WHEN** `lango payment balance` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
