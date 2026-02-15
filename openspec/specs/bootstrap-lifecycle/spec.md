## Purpose

Define the bootstrap sequence that initializes Lango's runtime: data directory, database, passphrase, crypto, and config profile loading.
## Requirements
### Requirement: Unified bootstrap sequence
The system SHALL execute a complete bootstrap sequence: ensure data directory → open database → acquire passphrase → initialize crypto → load config profile. The result SHALL be a single struct containing all initialized components. The `Options` struct SHALL NOT include a `MigrationPath` field.

#### Scenario: First-run bootstrap
- **WHEN** no salt exists in the database (first run)
- **THEN** the system acquires a new passphrase (with confirmation), generates a salt, stores the checksum, creates a default config profile, and returns the Result

#### Scenario: Returning-user bootstrap
- **WHEN** salt and checksum exist in the database
- **THEN** the system acquires the passphrase, verifies it against the stored checksum, and loads the active profile

#### Scenario: Wrong passphrase on returning user
- **WHEN** the user provides an incorrect passphrase for an existing database
- **THEN** the system returns a "passphrase checksum mismatch" error

#### Scenario: No profiles exist
- **WHEN** no profiles exist in the database
- **THEN** the system creates a default profile with `config.DefaultConfig()` and sets it as active

### Requirement: Shared database client
The bootstrap Result SHALL include the `*ent.Client` so downstream components (session store, key registry) can reuse it without opening a second connection.

#### Scenario: DB client reuse
- **WHEN** the bootstrap Result is passed to `app.New()`
- **THEN** the session store uses `NewEntStoreWithClient()` with the bootstrap's client

### Requirement: Data directory initialization
The system SHALL ensure `~/.lango/` exists with 0700 permissions during bootstrap.

#### Scenario: Directory does not exist
- **WHEN** `~/.lango/` does not exist
- **THEN** the directory is created with 0700 permissions

