## MODIFIED Requirements

### Requirement: Recovery confirmation-word prompt uses shared prompt helper
`lango security recovery setup` SHALL route its confirmation-word prompt through the shared visible line-entry prompt helper using Cobra command input/output streams.

#### Scenario: Recovery confirmation-word accepts matching final line without trailing newline
- **WHEN** the operator enters the correct confirmation word and the input stream ends immediately after that line without a trailing newline
- **THEN** `lango security recovery setup` SHALL accept the confirmation word instead of surfacing a read error
