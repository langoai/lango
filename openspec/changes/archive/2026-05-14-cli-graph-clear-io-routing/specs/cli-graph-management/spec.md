## MODIFIED Requirements

### Requirement: Graph clear command I/O routing
`lango graph clear` SHALL route prompt and result output through the Cobra command writer and SHALL read confirmation input through the Cobra command input reader.

#### Scenario: Graph clear prompt uses command streams
- **WHEN** `lango graph clear` prompts for confirmation
- **THEN** it SHALL write the prompt through the Cobra command output writer
- **AND** it SHALL read the operator response through the Cobra command input reader
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` and `cmd.InOrStdin()` SHALL control the interaction
