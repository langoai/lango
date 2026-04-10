## MODIFIED Requirements

### Requirement: Unified bootstrap sequence

The system SHALL execute a complete bootstrap sequence with 10 phases: ensure data directory → detect encryption → load envelope file → acquire credential → unwrap or create MK → open database → migrate envelope → load security state → initialize crypto → load config profile. The `Options` struct SHALL include `KeepKeyfile bool` (defaults to false), `SkipSecureDetection bool`, and `LangoDir string` (defaults to `~/.lango/` if empty). The `Options` struct SHALL NOT include a `MigrationPath` field. The result SHALL be a single struct containing all initialized components.

#### Scenario: First-run bootstrap on fresh install

- **WHEN** no envelope file and no legacy salt/checksum exist
- **THEN** the system acquires a new passphrase (with confirmation), generates a random 32-byte MK, creates an envelope with a passphrase KEK slot, persists `envelope.json`, and opens the database (with MK-derived DB key if encryption enabled)
- **AND** returns the Result

#### Scenario: Returning-user bootstrap with envelope

- **WHEN** an envelope file exists with no pending flags
- **THEN** the system loads the envelope, acquires the passphrase, unwraps the MK via `UnwrapFromPassphrase`, derives the DB key via `DeriveDBKey(mk)`, opens the database with `PRAGMA key = "x'<hex>'"`, and loads the active profile

#### Scenario: Wrong passphrase on envelope-based install

- **WHEN** the user provides an incorrect passphrase for an existing envelope
- **THEN** `UnwrapFromPassphrase` returns `ErrUnwrapFailed`
- **AND** the system returns the error wrapped with context
- **AND** the keyfile is NOT shredded

#### Scenario: Legacy-to-envelope migration on first boot after upgrade

- **WHEN** the database has legacy salt/checksum but no envelope file
- **THEN** the system acquires the passphrase, verifies the legacy checksum, runs `MigrateToEnvelope` (data re-encryption + SQLCipher rekey if applicable), and persists the new envelope
- **AND** the user sees a one-time "Upgrading encryption format..." message

#### Scenario: Crash recovery during migration

- **WHEN** the envelope file has `PendingMigration = true` or `PendingRekey = true`
- **THEN** Phase 6 (OpenDatabase) uses the legacy passphrase as the DB key (fallback)
- **AND** Phase 7 (MigrateEnvelope) retries the pending operations
- **AND** on success, the pending flags are cleared and the updated envelope is persisted

#### Scenario: No profiles exist

- **WHEN** no profiles exist in the database
- **THEN** the system creates a default profile with `config.DefaultConfig()` and sets it as active

### Requirement: Data directory initialization

The system SHALL ensure the lango data directory exists with 0700 permissions during bootstrap. The directory path SHALL be `Options.LangoDir` if non-empty, otherwise `~/.lango/`. The envelope file, keyfile, and skills directory SHALL be placed under this directory.

#### Scenario: Default LangoDir

- **WHEN** `Options.LangoDir` is empty
- **THEN** `s.LangoDir` is set to `~/.lango/`
- **AND** the directory is created with 0700 permissions

#### Scenario: Custom LangoDir for testing

- **WHEN** `Options.LangoDir` is set to a custom path (e.g., `t.TempDir()`)
- **THEN** `s.LangoDir` is set to that path
- **AND** the directory is created with 0700 permissions
- **AND** envelope file is stored at `<custom>/envelope.json`

## ADDED Requirements

### Requirement: Envelope file loading phase

The system SHALL execute a `LoadEnvelopeFile` phase early in bootstrap (before DB open) that attempts to read `<LangoDir>/envelope.json` from the filesystem. If the file does not exist, `State.Envelope` SHALL be set to nil. If the file exists but fails to parse, the phase SHALL return an error wrapping `ErrEnvelopeCorrupt`.

#### Scenario: Envelope file exists and is valid

- **WHEN** `<LangoDir>/envelope.json` exists with valid JSON
- **THEN** `State.Envelope` is populated with the parsed `MasterKeyEnvelope`
- **AND** no DB access occurs in this phase

#### Scenario: Envelope file does not exist

- **WHEN** `<LangoDir>/envelope.json` does not exist
- **THEN** `State.Envelope = nil`
- **AND** the phase completes without error

#### Scenario: Envelope file corrupt

- **WHEN** `<LangoDir>/envelope.json` exists but is invalid JSON
- **THEN** the phase returns an error wrapping `ErrEnvelopeCorrupt`

### Requirement: Unwrap or create MK phase

The system SHALL execute an `UnwrapOrCreateMK` phase after credential acquisition and before database open. This phase SHALL: (1) if MK was already unwrapped via mnemonic in Phase 4, return immediately; (2) if an envelope exists, unwrap the MK using the passphrase; (3) if no envelope exists and this is a first run, generate a new MK and envelope and persist the envelope file; (4) if no envelope exists and this is not a first run, mark `LegacyMode = true` for the migration phase.

#### Scenario: Unwrap from existing envelope

- **WHEN** `State.Envelope` is non-nil and `State.MasterKey` is nil
- **THEN** the phase calls `envelope.UnwrapFromPassphrase(passphrase)` and stores the result in `State.MasterKey`

#### Scenario: First run creates new MK and envelope

- **WHEN** `State.Envelope` is nil and bootstrap detects this is a first run
- **THEN** the phase calls `security.NewEnvelope(passphrase)` which generates a random MK and envelope
- **AND** `State.MasterKey` and `State.Envelope` are set
- **AND** `StoreEnvelopeFile` persists the new envelope

#### Scenario: Legacy mode flagged for migration

- **WHEN** `State.Envelope` is nil but legacy salt/checksum exist in the database
- **THEN** `State.LegacyMode = true`
- **AND** the phase does not attempt to create an envelope (MigrateEnvelope phase will handle it)

### Requirement: Migration phase

The system SHALL execute a `MigrateEnvelope` phase after database open. If `State.LegacyMode` is true, the phase SHALL perform full legacy-to-envelope migration. If the envelope has `PendingMigration = true`, the phase SHALL retry data re-encryption using the already-unwrapped MK. If the envelope has `PendingRekey = true` (after successful migration), the phase SHALL retry `PRAGMA rekey`. All crash recovery SHALL be idempotent.

#### Scenario: Full legacy migration

- **WHEN** `State.LegacyMode = true`
- **THEN** the phase calls `MigrateToEnvelope` which generates MK, creates envelope, persists envelope file, backs up DB via `VACUUM INTO`, re-encrypts secrets and config_profiles in a SQL transaction with COUNT verification, and (for SQLCipher) runs `PRAGMA rekey`

#### Scenario: Retry pending migration

- **WHEN** `State.Envelope.PendingMigration = true`
- **THEN** the phase re-runs data re-encryption using the already-unwrapped MK
- **AND** clears `PendingMigration` on success
- **AND** persists the updated envelope

#### Scenario: Retry pending rekey

- **WHEN** `State.Envelope.PendingRekey = true` and `PendingMigration = false`
- **THEN** the phase runs `PRAGMA rekey = "x'<HKDF(mk)>'"` and verifies by reopening with the new key
- **AND** clears `PendingRekey` on success
- **AND** persists the updated envelope

### Requirement: Bootstrap State extensions for envelope

The `State` struct SHALL include the following additional fields to support envelope-based crypto: `Envelope *MasterKeyEnvelope`, `MasterKey []byte`, `LegacyMode bool`, `RecoveryMode bool`. The `MasterKey` field SHALL be zeroed on cleanup. Pending migration / pending rekey retry logic SHALL read `State.Envelope.PendingMigration` and `State.Envelope.PendingRekey` directly rather than mirror them into separate State fields — this keeps the envelope as the single source of truth for crash recovery state.

#### Scenario: MasterKey zeroed on cleanup

- **WHEN** bootstrap completes (success or failure) and pipeline cleanup runs
- **THEN** if `State.MasterKey` is non-nil, `security.ZeroBytes(State.MasterKey)` is called before the state is discarded

#### Scenario: Pending flags read from envelope

- **WHEN** Phase 7 (MigrateEnvelope) needs to decide whether to retry migration or rekey
- **THEN** it reads `State.Envelope.PendingMigration` and `State.Envelope.PendingRekey` directly
- **AND** it does NOT consult any State-level mirror fields
