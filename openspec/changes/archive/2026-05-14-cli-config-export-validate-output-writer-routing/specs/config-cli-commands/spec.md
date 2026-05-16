## MODIFIED Requirements

### Requirement: Config export command
The system SHALL provide a `lango config export <name>` command that outputs decrypted config as JSON. Passphrase verification is required (handled implicitly by the bootstrap process).

#### Scenario: Config export/validate output uses the command writer
- **WHEN** `lango config export` or `lango config validate` renders output
- **THEN** it SHALL write the full command output through the Cobra command output or error writers
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` SHALL capture the output
