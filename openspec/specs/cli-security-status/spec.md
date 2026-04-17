## Purpose

Capability spec for cli-security-status. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Security status command
Security status reads SHALL use broker-backed storage diagnostics rather than opening the SQLite database directly from the CLI process.

#### Scenario: Status command reads through broker
- **WHEN** the security status command needs database-backed counts or metadata
- **THEN** it SHALL query the broker-backed storage layer instead of opening the database directly in the CLI process

### Requirement: Non-interactive mini-bootstrap for status

The system SHALL provide a `readDBStatusNonInteractive` helper that runs a minimal bootstrap (envelope load → non-interactive passphrase → MK unwrap → read-only DB open → read counts → close) without triggering interactive prompts or schema migration. The helper SHALL handle both envelope-based and legacy installations.

#### Scenario: Envelope-based non-interactive read

- **WHEN** `readDBStatusNonInteractive` is called with an envelope present and a keyring-stored passphrase
- **THEN** the helper unwraps the MK, derives the DB key via `DeriveDBKeyHex(mk)`, opens the DB read-only with `PRAGMA key = "x'<hex>'"`, reads key and secret counts, and closes the DB

#### Scenario: Legacy non-interactive read

- **WHEN** `readDBStatusNonInteractive` is called with no envelope and a keyfile-stored passphrase
- **THEN** the helper uses the passphrase directly as the DB key, opens the DB read-only, reads counts, and closes

#### Scenario: Keyring provider passed to non-interactive acquisition
- **WHEN** `readDBStatusNonInteractive` acquires a passphrase
- **THEN** it SHALL pass `keyring.DetectSecureProvider()` as the `KeyringProvider` option

#### Scenario: Active config loaded when DB available
- **WHEN** `readDBStatusNonInteractive` successfully opens the DB with MK
- **THEN** it SHALL load the active config profile via `configstore.Store.LoadActive`
- **AND** the loaded config SHALL be used for status display instead of `DefaultConfig()`

#### Scenario: Keyfile fallback on stale keyring passphrase
- **WHEN** envelope unwrap fails with a keyring-sourced passphrase
- **THEN** `readDBStatusNonInteractive` SHALL retry with a keyfile-only acquisition
- **AND** legacy DB open failure with keyring passphrase SHALL also retry with keyfile

#### Scenario: No passphrase available

- **WHEN** `readDBStatusNonInteractive` is called and `AcquireNonInteractive` returns an error
- **THEN** the helper returns a zero-valued `dbStatusResult` (all counts 0)
- **AND** no DB open attempt is made

### Requirement: Read-only database open for status

The system SHALL provide an `OpenDatabaseReadOnly` function used by status commands. This function SHALL open the SQLite database in read-only mode (`file:path?mode=ro`), SHALL NOT invoke ent schema migration, and SHALL NOT create any tables or indexes. The contract: read-only, no migration, no prompt, failure returns error (caller degrades gracefully).

#### Scenario: Read-only open succeeds

- **WHEN** `OpenDatabaseReadOnly(dbPath, dbKey, rawKey)` is called with a valid DB and key
- **THEN** the function opens the DB with `mode=ro`
- **AND** does NOT call `Schema.Create`
- **AND** returns an `*ent.Client` backed by a read-only connection

#### Scenario: Read-only open rejects writes

- **WHEN** a write operation is attempted on the read-only client
- **THEN** SQLite returns a "read-only database" error

### Requirement: Status reflects brokered payload protection
The security status surface MUST report brokered payload protection state rather than SQLCipher page-encryption state once the new protection model is active.

#### Scenario: Payload protection status reporting
- **WHEN** the user runs the security status command after payload protection is enabled
- **THEN** the output reports broker/storage/payload-protection state
- **AND** it does not imply that SQLCipher page encryption is active

