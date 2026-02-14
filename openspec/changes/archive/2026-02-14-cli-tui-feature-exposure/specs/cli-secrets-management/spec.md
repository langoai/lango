## ADDED Requirements

### Requirement: Secrets list command
The system SHALL provide a `lango security secrets list` command that displays metadata for all stored secrets. Secret values SHALL never be displayed. Table output SHALL show NAME, KEY, CREATED, UPDATED, and ACCESS_COUNT columns. The command SHALL support `--json` for JSON output.

#### Scenario: List secrets
- **WHEN** user runs `lango security secrets list`
- **THEN** the command displays a table of all secret metadata without revealing values

#### Scenario: No secrets
- **WHEN** user runs `lango security secrets list` with no stored secrets
- **THEN** the command displays "No secrets stored." and exits with code 0

#### Scenario: JSON output
- **WHEN** user runs `lango security secrets list --json`
- **THEN** the command outputs a JSON array of secret metadata objects

### Requirement: Secrets set command
The system SHALL provide a `lango security secrets set <name>` command that stores an encrypted secret. The command SHALL require an interactive terminal and prompt for the secret value using hidden input. The name SHALL be a positional argument.

#### Scenario: Store a secret
- **WHEN** user runs `lango security secrets set api-key` in an interactive terminal
- **THEN** the command prompts for the secret value with hidden input, encrypts it, stores it, and displays a success message

#### Scenario: Non-interactive terminal
- **WHEN** user runs `lango security secrets set api-key` in a non-interactive terminal
- **THEN** the command exits with an error indicating an interactive terminal is required

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
The system SHALL provide a shared crypto initialization function that resolves the passphrase using a 3-tier fallback: `LANGO_PASSPHRASE` environment variable, then `security.passphrase` config value, then interactive prompt. The function SHALL load or generate salt, verify checksum, and return a ready-to-use SecretsStore.

#### Scenario: Passphrase from environment
- **WHEN** `LANGO_PASSPHRASE` is set and valid
- **THEN** the function uses it without prompting

#### Scenario: First-time setup
- **WHEN** no salt exists in the database
- **THEN** the function generates a random salt, stores it, initializes the crypto provider, and stores the checksum

#### Scenario: Incorrect passphrase
- **WHEN** the provided passphrase does not match the stored checksum
- **THEN** the function returns an "incorrect passphrase" error
