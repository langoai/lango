## ADDED Requirements

### Requirement: Graph clear confirmation uses shared command streams
`lango graph clear` SHALL drive its confirmation prompt through the shared confirmation helper using Cobra command input/output streams.

#### Scenario: Graph clear aborts on denial
- **WHEN** `lango graph clear` prompts for confirmation and the operator answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave the graph store unchanged

#### Scenario: Graph clear prompt uses command streams
- **WHEN** `lango graph clear` prompts for confirmation
- **THEN** the warning line and `Continue? [y/N]: ` prompt SHALL be written through the Cobra command output stream
- **AND** the operator response SHALL be read through the Cobra command input stream
