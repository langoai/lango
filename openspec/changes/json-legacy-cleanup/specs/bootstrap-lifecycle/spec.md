## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: Automatic migration from lango.json
**Reason**: Automatic JSON migration on startup is removed. Configuration is managed exclusively through `lango onboard` (TUI) and `lango config` (CLI). Users who need to import existing JSON files SHALL use `lango config import` explicitly.
**Migration**: Use `lango config import <file>` to manually import JSON config files into encrypted profiles.
