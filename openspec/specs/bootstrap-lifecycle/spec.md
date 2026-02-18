## Purpose

Define the bootstrap sequence that initializes Lango's runtime: data directory, database, passphrase, crypto, and config profile loading.
## Requirements
### Requirement: Unified bootstrap sequence
The system SHALL execute a complete bootstrap sequence: ensure data directory → open database → acquire passphrase → initialize crypto → shred keyfile (if applicable) → load config profile. The result SHALL be a single struct containing all initialized components. The `Options` struct SHALL NOT include a `MigrationPath` field. The `Options` struct SHALL include a `KeepKeyfile bool` field that defaults to false (secure by default).

#### Scenario: First-run bootstrap
- **WHEN** no salt exists in the database (first run)
- **THEN** the system acquires a new passphrase (with confirmation), generates a salt, stores the checksum, shreds the keyfile if source is keyfile, creates a default config profile, and returns the Result

#### Scenario: Returning-user bootstrap
- **WHEN** salt and checksum exist in the database
- **THEN** the system acquires the passphrase, verifies it against the stored checksum, shreds the keyfile if source is keyfile, and loads the active profile

#### Scenario: Wrong passphrase on returning user
- **WHEN** the user provides an incorrect passphrase for an existing database
- **THEN** the system returns a "passphrase checksum mismatch" error and the keyfile is NOT shredded

#### Scenario: No profiles exist
- **WHEN** no profiles exist in the database
- **THEN** the system creates a default profile with `config.DefaultConfig()` and sets it as active

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
The system SHALL ensure `~/.lango/` exists with 0700 permissions during bootstrap.

#### Scenario: Directory does not exist
- **WHEN** `~/.lango/` does not exist
- **THEN** the directory is created with 0700 permissions

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
