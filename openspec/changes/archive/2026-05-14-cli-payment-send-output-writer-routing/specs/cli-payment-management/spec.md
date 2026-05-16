## MODIFIED Requirements

### Requirement: Send command
The system SHALL provide a `lango payment send` command that sends USDC to a recipient address with required flags --to, --amount, and --purpose.

#### Scenario: Payment send output uses the command streams
- **WHEN** `lango payment send` renders a confirmation prompt, success output, or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** it SHALL read confirmation input from `cmd.InOrStdin()`
- **AND** wrappers or tests that replace those streams SHALL capture the interaction
