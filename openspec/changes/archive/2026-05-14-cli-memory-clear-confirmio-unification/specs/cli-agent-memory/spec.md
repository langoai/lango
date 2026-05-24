## ADDED Requirements

### Requirement: Memory clear confirmation uses shared command streams
`lango memory clear <session-key>` SHALL drive its confirmation prompt through the shared confirmation helper using Cobra command input/output streams.

#### Scenario: Memory clear aborts on denial
- **WHEN** `lango memory clear user-123` prompts for confirmation and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave memory entries untouched

#### Scenario: Memory clear prompt uses command streams
- **WHEN** `lango memory clear user-123` prompts for confirmation
- **THEN** the warning line and `Continue? [y/N]: ` prompt SHALL be written through the Cobra command output stream
- **AND** the operator response SHALL be read through the Cobra command input stream
