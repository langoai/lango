## ADDED Requirements

### Requirement: Dead-letter retry confirmation uses shared command streams
`lango status dead-letter retry` SHALL drive its interactive confirmation through the shared confirmation helper using Cobra command input/output streams. Empty input or EOF SHALL continue to abort the retry without invoking the retry mutation.

#### Scenario: Dead-letter retry denial aborts cleanly
- **WHEN** `lango status dead-letter retry <id>` is run and the operator answers `n`
- **THEN** the command SHALL print an aborted message
- **AND** it SHALL NOT invoke the retry mutation

#### Scenario: Dead-letter retry EOF aborts cleanly
- **WHEN** `lango status dead-letter retry <id>` is run and the confirmation input reaches EOF without approval
- **THEN** the command SHALL abort the retry without invoking the retry mutation

#### Scenario: Dead-letter retry prompt uses command output
- **WHEN** `lango status dead-letter retry <id>` prompts for confirmation
- **THEN** the sanitized transaction receipt ID and `[y/N]:` suffix SHALL be written through the Cobra command output stream
