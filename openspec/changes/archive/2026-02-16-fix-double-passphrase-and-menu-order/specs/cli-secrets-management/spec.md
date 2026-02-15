## MODIFIED Requirements

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
