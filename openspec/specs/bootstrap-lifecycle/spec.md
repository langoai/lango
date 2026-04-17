## Purpose

Define the bootstrap sequence that initializes Lango's runtime: data directory, database, passphrase, crypto, and config profile loading.
## Requirements
### Requirement: Unified bootstrap sequence
The system SHALL execute a complete bootstrap sequence with broker-owned database initialization: ensure data directory → detect encryption/header state → load envelope file → acquire credential → spawn storage broker → open database through broker → load security state/config/profile via broker-backed storage → initialize runtime services. The result SHALL be a single struct containing the initialized runtime handles, but SHALL NOT expose direct `*sql.DB` or `*ent.Client` ownership to callers once broker mode is active.

#### Scenario: Broker bootstrap on returning user
- **WHEN** bootstrap runs for a normal application start
- **THEN** the parent process SHALL spawn the storage broker before loading config profiles
- **AND** the broker SHALL own the SQLite open/migration step

#### Scenario: Broker bootstrap on first run
- **WHEN** bootstrap runs on a fresh install
- **THEN** credential acquisition and master-key setup SHALL complete before the broker `open_db` handshake is attempted
- **AND** the broker SHALL prepare the database before profile creation proceeds

### Requirement: Profile loading applies PostLoad normalization
The `phaseLoadProfile` phase SHALL call `config.PostLoad()` exactly once at the end, after all branches (explicit profile, active profile, default profile) have set the config. No branch SHALL return early before PostLoad is applied.

#### Scenario: Explicit profile gets PostLoad applied
- **WHEN** `ForceProfile` is set and the profile is loaded successfully
- **THEN** `PostLoad()` is called on the loaded config before the phase completes

#### Scenario: Active profile gets PostLoad applied
- **WHEN** an active profile exists and is loaded
- **THEN** `PostLoad()` is called on the loaded config before the phase completes

#### Scenario: Default profile gets PostLoad applied
- **WHEN** no active profile exists and a default is created via `handleNoProfile`
- **THEN** `PostLoad()` is called on the created config before the phase completes

#### Scenario: PostLoad failure fails the phase
- **WHEN** `PostLoad()` returns an error on the loaded config
- **THEN** the phase returns that error wrapped with context

### Requirement: Shared database client
The bootstrap Result SHALL include the `*ent.Client` so downstream components (session store, key registry) can reuse it without opening a second connection. The underlying `*sql.DB` SHALL be configured with WAL journal mode, a busy_timeout of 5000ms, MaxOpenConns of 4, and MaxIdleConns of 4. These settings SHALL be applied in bootstrap before creating the Ent client, and no downstream component SHALL override connection pool settings on the shared `*sql.DB`.

#### Scenario: DB client reuse
- **WHEN** the bootstrap Result is passed to `app.New()`
- **THEN** the session store uses `NewEntStoreWithClient()` with the bootstrap's client

#### Scenario: WAL mode enabled at connection open
- **WHEN** the SQLite database is opened during bootstrap
- **THEN** the connection string includes `_journal_mode=WAL` and `_busy_timeout=5000`

#### Scenario: Connection pool configured centrally
- **WHEN** the `*sql.DB` is created during bootstrap
- **THEN** `MaxOpenConns` is set to 4 and `MaxIdleConns` is set to 4
- **AND** no other component overrides these settings

#### Scenario: Concurrent audit log write during active operation
- **WHEN** a background goroutine writes an audit log while another operation holds a write lock
- **THEN** the audit log write waits (up to busy_timeout) and succeeds without "database table is locked" error

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

### Requirement: Bootstrap uses secure hardware provider for passphrase storage
The bootstrap process SHALL use `DetectSecureProvider()` to determine the keyring provider for passphrase acquisition. When no secure hardware is available (`TierNone`), the keyring provider SHALL be nil, disabling automatic keyring reads.

#### Scenario: Biometric available during bootstrap
- **WHEN** bootstrap runs on macOS with Touch ID
- **THEN** the passphrase acquisition SHALL use `BiometricProvider` as the keyring provider

#### Scenario: No secure hardware during bootstrap
- **WHEN** bootstrap runs on a system without biometric or TPM
- **THEN** the keyring provider SHALL be nil, and passphrase SHALL be acquired from keyfile or interactive prompt only

#### Scenario: Interactive passphrase with secure storage offer
- **WHEN** the passphrase source is interactive and a secure provider is available
- **THEN** the system SHALL offer to store the passphrase in the secure backend with a confirmation prompt showing the tier label

### Requirement: Report biometric passphrase store outcome
When the bootstrap flow stores a passphrase in the secure keyring provider, it SHALL report the outcome to stderr. On entitlement error (`ErrEntitlement`), the system SHALL warn the user and suggest codesigning. On other failures, the message SHALL be `warning: store passphrase failed: <error>`. On success, the message SHALL be `Passphrase saved. Next launch will load it automatically.`.

#### Scenario: Biometric store succeeds
- **WHEN** `secureProvider.Set()` returns nil
- **THEN** stderr SHALL contain `Passphrase saved. Next launch will load it automatically.`

#### Scenario: Biometric store fails with entitlement error
- **WHEN** `secureProvider.Set()` returns an error satisfying `errors.Is(err, keyring.ErrEntitlement)`
- **THEN** stderr SHALL contain `warning: biometric storage unavailable (binary not codesigned)`
- **AND** stderr SHALL contain a codesign tip

#### Scenario: Biometric store fails with non-entitlement error
- **WHEN** `secureProvider.Set()` returns an error NOT satisfying `errors.Is(err, keyring.ErrEntitlement)`
- **THEN** stderr SHALL contain `warning: store passphrase failed: <error detail>`

### Requirement: SkipSecureDetection option for testing
The `Options` struct SHALL include a `SkipSecureDetection` boolean. When true, secure hardware detection SHALL be skipped and the keyring provider SHALL be nil regardless of available hardware.

#### Scenario: SkipSecureDetection in test
- **WHEN** `Run()` is called with `SkipSecureDetection: true`
- **THEN** the bootstrap SHALL not probe for biometric or TPM hardware

### Requirement: Ephemeral keyfile shredding after crypto initialization
The system SHALL shred the passphrase keyfile after successful crypto initialization and checksum verification when the passphrase source is keyfile and `KeepKeyfile` is false (default). Shred failure SHALL emit a warning to stderr but SHALL NOT prevent bootstrap from completing.

#### Scenario: Keyfile shredded after successful bootstrap
- **WHEN** the passphrase source is `SourceKeyfile` and `KeepKeyfile` is false
- **AND** crypto initialization and checksum verification succeed
- **THEN** the keyfile is securely shredded and no longer exists on disk

#### Scenario: Keyfile kept when opted out
- **WHEN** the passphrase source is `SourceKeyfile` and `KeepKeyfile` is true
- **THEN** the keyfile remains on disk after bootstrap

#### Scenario: Non-keyfile source unaffected
- **WHEN** the passphrase source is `SourceInteractive` or `SourceStdin`
- **THEN** no shredding is attempted regardless of `KeepKeyfile` value

#### Scenario: Shred failure is non-fatal
- **WHEN** `ShredKeyfile()` returns an error during bootstrap
- **THEN** a warning is printed to stderr and bootstrap continues with the already-initialized crypto provider

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

### Requirement: KMS bootstrap env config

The system SHALL provide a `KMSConfigFromEnv()` function that reads KMS KEK configuration from environment variables. Provider-specific env vars:

#### Scenario: AWS KMS / GCP KMS env config
- **WHEN** `LANGO_KMS_PROVIDER` is `aws-kms` or `gcp-kms`
- **THEN** the function reads `LANGO_KMS_KEY_ID`, `LANGO_KMS_REGION`, `LANGO_KMS_ENDPOINT`

#### Scenario: Azure Key Vault env config
- **WHEN** `LANGO_KMS_PROVIDER` is `azure-kv`
- **THEN** the function reads `LANGO_KMS_KEY_ID`, `LANGO_KMS_AZURE_VAULT_URL`, `LANGO_KMS_AZURE_KEY_VERSION`

#### Scenario: PKCS#11 HSM env config
- **WHEN** `LANGO_KMS_PROVIDER` is `pkcs11`
- **THEN** the function reads `LANGO_KMS_PKCS11_MODULE`, `LANGO_KMS_PKCS11_SLOT_ID`, `LANGO_KMS_PKCS11_KEY_LABEL`, `LANGO_PKCS11_PIN`

#### Scenario: Missing provider env var
- **WHEN** `LANGO_KMS_PROVIDER` is not set
- **THEN** the function returns nil config and empty provider name

