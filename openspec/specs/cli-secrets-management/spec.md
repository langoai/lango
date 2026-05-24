## Purpose

Capability spec for cli-secrets-management. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Secrets command stream routing
`lango security secrets list`, `set`, and `delete` SHALL route human-readable and JSON output through the Cobra command writer. Interactive `set` and `delete` prompts SHALL write prompt text through the Cobra command output writer. Interactive delete confirmation SHALL read input through the Cobra command input reader.

#### Scenario: Secrets output uses the command writer
- **WHEN** `lango security secrets` commands render table, JSON, prompt, or success output
- **THEN** they SHALL write the full visible output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Secrets set prompt uses command output
- **WHEN** `lango security secrets set <name>` prompts for an interactive secret value
- **THEN** it SHALL write the hidden-input prompt through the Cobra command output writer

#### Scenario: Secrets delete prompt uses command streams
- **WHEN** `lango security secrets delete <name>` prompts for confirmation
- **THEN** it SHALL write the prompt through the Cobra command output writer
- **AND** it SHALL read the operator response through the Cobra command input reader

### Requirement: Secrets list command
The system SHALL provide a `lango security secrets list [--output table|json]` command that displays metadata for all stored secrets. Secret values SHALL never be displayed. Table output SHALL show NAME, KEY, CREATED, UPDATED, and ACCESS_COUNT columns.

#### Scenario: List secrets
- **WHEN** user runs `lango security secrets list`
- **THEN** the command displays a table of all secret metadata without revealing values

#### Scenario: No secrets
- **WHEN** user runs `lango security secrets list` with no stored secrets
- **THEN** the command displays "No secrets stored." and exits with code 0

#### Scenario: JSON output
- **WHEN** user runs `lango security secrets list --output json`
- **THEN** the command outputs a JSON array of secret metadata objects with `name`, `key_name`, `created_at`, `updated_at`, and `access_count` fields

### Requirement: Secrets set command
The system SHALL provide a `lango security secrets set <name>` command that stores an encrypted secret value either interactively (via passphrase prompt) or non-interactively (via `--value-hex` flag). When `--value-hex` is provided, the command SHALL hex-decode the input (stripping an optional `0x` prefix) and store the raw bytes. When `--value-hex` is not provided, the command SHALL require an interactive command input stream and prompt for the value. The name SHALL be a positional argument.

#### Scenario: Interactive secret storage
- **WHEN** user runs `lango security secrets set api-key` in an interactive terminal without `--value-hex`
- **THEN** the command prompts for the secret value with hidden input, encrypts it, stores it, and displays a success message

#### Scenario: Non-interactive hex secret storage
- **WHEN** user runs `lango security secrets set wallet.privatekey --value-hex 0xac0974...` in a non-interactive environment
- **THEN** the command SHALL hex-decode the value (stripping `0x` prefix), store the raw bytes encrypted, and print success

#### Scenario: Non-interactive without value-hex flag
- **WHEN** user runs `lango security secrets set api-key` with a non-interactive command input stream and without `--value-hex`
- **THEN** the command exits with an error suggesting `--value-hex` for non-interactive use

#### Scenario: Invalid hex value
- **WHEN** user runs `lango security secrets set mykey --value-hex "not-hex"`
- **THEN** the command SHALL return a hex decode error

#### Scenario: Secrets-set guard uses command input stream
- **WHEN** `lango security secrets set <name>` prompts interactively
- **THEN** its interactive guard SHALL validate `cmd.InOrStdin()` instead of reading process-global stdin directly

#### Scenario: Update existing secret
- **WHEN** user runs `lango security secrets set api-key` for a name that already exists
- **THEN** the command overwrites the existing secret with the new encrypted value

### Requirement: Secrets delete command
The system SHALL provide a `lango security secrets delete <name>` command that removes a stored secret. The command SHALL prompt for confirmation before deletion. The `--force` flag SHALL skip the confirmation prompt.

#### Scenario: Delete with confirmation
- **WHEN** user runs `lango security secrets delete api-key` and confirms with "y"
- **THEN** the command deletes the secret and displays a success message

#### Scenario: Force delete
- **WHEN** user runs `lango security secrets delete api-key --force`
- **THEN** the command deletes the secret without prompting

#### Scenario: Delete nonexistent secret
- **WHEN** user runs `lango security secrets delete nonexistent --force`
- **THEN** the command exits with an error indicating the secret was not found

### Requirement: Crypto initialization helper
The system SHALL provide a shared crypto initialization function (`secretsStoreFromBoot`) that creates a `SecretsStore` directly from a `*bootstrap.Result`. The function SHALL reuse the already-initialized `CryptoProvider` and `*ent.Client` from the bootstrap result, register the default encryption key, and return a ready-to-use `SecretsStore`. The function SHALL NOT independently acquire a passphrase — the passphrase MUST be acquired exactly once during `bootstrap.Run()`.

#### Scenario: Passphrase acquired once
- **WHEN** user runs any security secrets command
- **THEN** the passphrase SHALL be prompted exactly once during bootstrap, not again during secrets store initialization

#### Scenario: First-time setup
- **WHEN** no salt exists in the database
- **THEN** the bootstrap process handles salt generation and checksum storage before the secrets store is created

#### Scenario: Incorrect passphrase
- **WHEN** the provided passphrase does not match the stored checksum
- **THEN** the bootstrap process returns an "incorrect passphrase" error before any secrets store creation occurs

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

### Requirement: Secrets-delete EOF aborts cleanly
`lango security secrets delete <name>` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Secrets-delete EOF aborts without deletion
- **WHEN** `lango security secrets delete api-key` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave the secret undeleted
