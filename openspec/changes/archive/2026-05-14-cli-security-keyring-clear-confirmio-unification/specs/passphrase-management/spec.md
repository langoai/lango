## ADDED Requirements

### Requirement: Keyring-clear confirmation uses shared command streams
`lango security keyring clear` SHALL drive its interactive confirmation through the shared confirmation helper using Cobra command input/output streams. When stdin is non-interactive and `--force` is not provided, the command SHALL refuse to continue with explicit `--force` guidance instead of attempting a prompt.

#### Scenario: Keyring-clear denial prints abort message
- **WHEN** `lango security keyring clear` is run interactively and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave stored credentials untouched

#### Scenario: Keyring-clear confirm prints prompt on command output
- **WHEN** `lango security keyring clear` is run interactively and the user answers `y`
- **THEN** the confirmation prompt SHALL be written through the Cobra command output stream
- **AND** the command SHALL continue with the backend clearing flow

#### Scenario: Keyring-clear non-interactive path requires force
- **WHEN** `lango security keyring clear` is run with non-interactive stdin and without `--force`
- **THEN** the command SHALL return an error directing the user to pass `--force for non-interactive deletion`
