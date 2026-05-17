## ADDED Requirements

### Requirement: Security passphrase command guards use command input streams
Interactive `lango security` passphrase commands SHALL validate the Cobra command
input stream before hidden passphrase prompts instead of reading process-global
stdin directly.

#### Scenario: Change-passphrase guard uses command input
- **WHEN** `lango security change-passphrase` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream

#### Scenario: Migrate-passphrase guard uses command input
- **WHEN** `lango security migrate-passphrase` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream

#### Scenario: Keyring-store guard uses command input
- **WHEN** `lango security keyring store` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream
