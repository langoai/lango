## ADDED Requirements

### Requirement: Recovery confirmation-word prompt uses shared prompt helper
`lango security recovery setup` SHALL route its confirmation-word prompt through the shared visible line-entry prompt helper using Cobra command input/output streams.

#### Scenario: Recovery confirmation-word prompt uses shared command streams
- **WHEN** `lango security recovery setup` asks the operator to enter a specific mnemonic word
- **THEN** the prompt text SHALL be written through the Cobra command output stream
- **AND** the entered word SHALL be read through the Cobra command input stream via the shared prompt helper
