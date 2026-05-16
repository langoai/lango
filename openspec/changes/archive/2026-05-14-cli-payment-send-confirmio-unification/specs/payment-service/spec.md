## ADDED Requirements

### Requirement: Payment-send confirmation uses shared command streams
`lango payment send` SHALL drive its interactive confirmation through the shared confirmation helper using Cobra command input/output streams. When stdin is non-interactive and `--force` is not provided, the command SHALL refuse to continue with explicit `--force` guidance instead of attempting a prompt.

#### Scenario: Payment-send denial prints abort message
- **WHEN** `lango payment send` is run interactively and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL NOT submit a payment

#### Scenario: Payment-send prompt stays on command output
- **WHEN** `lango payment send` is run interactively
- **THEN** the payment summary and `Confirm [y/N]: ` prompt SHALL be written through the Cobra command output stream

#### Scenario: Payment-send non-interactive path requires force
- **WHEN** `lango payment send` is run with non-interactive input and without `--force`
- **THEN** the command SHALL return an error directing the user to pass `--force for non-interactive mode`
