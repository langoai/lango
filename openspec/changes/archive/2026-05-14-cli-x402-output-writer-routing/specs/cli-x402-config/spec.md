## MODIFIED Requirements

### Requirement: Payment x402 command
The system SHALL provide a `lango payment x402 [--json]` command that displays the X402 protocol configuration including enabled state, wallet address, payment endpoint, and accepted token types. The command SHALL use cfgLoader (config only).

#### Scenario: X402 output uses the command writer
- **WHEN** `lango payment x402` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
