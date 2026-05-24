## MODIFIED Requirements

### Requirement: Keyring store output routing
`lango security keyring store` SHALL route non-error status output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Keyring store output uses the command writer
- **WHEN** `lango security keyring store` reports that the passphrase is already stored or that storage completed successfully
- **THEN** it SHALL write the full status output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
