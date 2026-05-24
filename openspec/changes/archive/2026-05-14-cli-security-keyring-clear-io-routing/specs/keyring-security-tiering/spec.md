## MODIFIED Requirements

### Requirement: Keyring clear command I/O routing
`lango security keyring clear` SHALL route prompt and result output through the Cobra command writer, SHALL route warnings through the Cobra error writer, and SHALL read confirmation input through the Cobra command input reader.

#### Scenario: Keyring clear prompt uses command streams
- **WHEN** `lango security keyring clear` prompts for confirmation
- **THEN** it SHALL write the prompt through the Cobra command output writer
- **AND** it SHALL read the operator response through the Cobra command input reader
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` and `cmd.InOrStdin()` SHALL control the interaction
