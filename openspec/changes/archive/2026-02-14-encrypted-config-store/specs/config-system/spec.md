## MODIFIED Requirements

### Requirement: Configuration loading
The system SHALL load configuration through the bootstrap process from an encrypted SQLite database profile instead of directly from a plaintext JSON file. The `config.Load()` function SHALL be retained for migration purposes only.

#### Scenario: Normal startup
- **WHEN** the application starts via `lango serve`
- **THEN** configuration is loaded via `bootstrap.Run()` which reads the active encrypted profile

#### Scenario: Migration loading
- **WHEN** `config.Load()` is called during JSON import
- **THEN** the JSON file is read with environment variable substitution (existing behavior preserved)

### Requirement: Configuration save
The system SHALL save configuration through `configstore.Store.Save()` which encrypts and stores in the database. The legacy `config.Save()` function SHALL be simplified to a basic JSON marshal without sanitization.

#### Scenario: Save via configstore
- **WHEN** a config is saved through the configstore
- **THEN** it is JSON-serialized, AES-256-GCM encrypted, and stored in the database

## REMOVED Requirements

### Requirement: SecurityConfig.Passphrase field
**Reason**: Storing passphrases in the config struct is a security anti-pattern. Passphrase acquisition is now handled by the dedicated `passphrase` package.
**Migration**: Remove `security.passphrase` from `lango.json` before import, or let the migration process ignore it.

### Requirement: Config sanitization on save
**Reason**: With encrypted storage, plaintext sanitization (`***REDACTED***` replacement) is unnecessary. The entire config blob is encrypted at rest.
**Migration**: Use `lango config export` to get plaintext when needed.
