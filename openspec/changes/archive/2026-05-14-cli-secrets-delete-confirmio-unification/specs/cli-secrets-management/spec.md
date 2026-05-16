## ADDED Requirements

### Requirement: Secrets-delete confirmation uses shared command streams
`lango security secrets delete` SHALL drive its interactive confirmation through the shared confirmation helper using Cobra command input/output streams. When stdin is non-interactive and `--force` is not provided, the command SHALL refuse to continue with explicit `--force` guidance instead of attempting a prompt.

#### Scenario: Secrets-delete denial prints abort message
- **WHEN** `lango security secrets delete api-key` is run interactively and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave the secret undeleted

#### Scenario: Secrets-delete prompt stays on command output
- **WHEN** `lango security secrets delete api-key` prompts for confirmation
- **THEN** the `Delete secret 'api-key'? [y/N]: ` prompt SHALL be written through the Cobra command output stream
- **AND** the response SHALL be read through the Cobra command input stream

#### Scenario: Secrets-delete non-interactive path requires force
- **WHEN** `lango security secrets delete api-key` is run with non-interactive input and without `--force`
- **THEN** the command SHALL return an error directing the user to pass `--force for non-interactive deletion`
