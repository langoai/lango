## ADDED Requirements

### Requirement: Unified bootstrap sequence
The system SHALL execute a complete bootstrap sequence: ensure data directory → open database → acquire passphrase → initialize crypto → load config profile. The result SHALL be a single struct containing all initialized components.

#### Scenario: First-run bootstrap
- **WHEN** no salt exists in the database (first run)
- **THEN** the system acquires a new passphrase (with confirmation), generates a salt, stores the checksum, creates a default config profile, and returns the Result

#### Scenario: Returning-user bootstrap
- **WHEN** salt and checksum exist in the database
- **THEN** the system acquires the passphrase, verifies it against the stored checksum, and loads the active profile

#### Scenario: Wrong passphrase on returning user
- **WHEN** the user provides an incorrect passphrase for an existing database
- **THEN** the system returns a "passphrase checksum mismatch" error

### Requirement: Automatic migration from lango.json
The system SHALL detect existing `lango.json` files when no config profiles exist and migrate them to encrypted storage.

#### Scenario: Auto-detect lango.json
- **WHEN** no profiles exist and `lango.json` is found in the current directory or `~/.lango/`
- **THEN** the system migrates the JSON config as the "default" active profile

#### Scenario: Explicit migration path
- **WHEN** a `MigrationPath` option is provided
- **THEN** the system uses that path as the primary migration source

#### Scenario: No config found
- **WHEN** no profiles exist and no JSON file is found
- **THEN** the system creates a default profile with `config.DefaultConfig()`

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
