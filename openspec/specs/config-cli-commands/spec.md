## ADDED Requirements

### Requirement: Config list command
The system SHALL provide a `lango config list` command that displays all profiles with name, active marker, version, and timestamps in a table format.

#### Scenario: List with profiles
- **WHEN** `lango config list` is run with existing profiles
- **THEN** a table is displayed with columns: NAME, ACTIVE, VERSION, CREATED, UPDATED

#### Scenario: List with no profiles
- **WHEN** no profiles exist
- **THEN** the message "No profiles found." is displayed

### Requirement: Config create command
The system SHALL provide a `lango config create <name>` command that creates a new profile with default configuration.

#### Scenario: Create new profile
- **WHEN** `lango config create staging` is run and "staging" does not exist
- **THEN** a profile named "staging" is created with default config

#### Scenario: Create duplicate profile
- **WHEN** `lango config create default` is run and "default" already exists
- **THEN** an error is returned: "profile \"default\" already exists"

### Requirement: Config use command
The system SHALL provide a `lango config use <name>` command that switches the active profile.

#### Scenario: Switch active profile
- **WHEN** `lango config use production` is run
- **THEN** the "production" profile becomes active and all others are deactivated

### Requirement: Config delete command
The system SHALL provide a `lango config delete <name>` command with confirmation prompt.

#### Scenario: Delete with confirmation
- **WHEN** `lango config delete staging` is run without `--force`
- **THEN** a confirmation prompt is shown before deletion

#### Scenario: Delete with force flag
- **WHEN** `lango config delete staging --force` is run
- **THEN** the profile is deleted without confirmation

### Requirement: Config import command
The system SHALL provide a `lango config import <file>` command that imports a JSON config file as an encrypted profile.

#### Scenario: Import JSON file
- **WHEN** `lango config import lango.json --profile migrated` is run
- **THEN** the JSON file is loaded, encrypted, and stored as profile "migrated" (set as active)

### Requirement: Config export command
The system SHALL provide a `lango config export <name>` command that outputs decrypted config as JSON with a security warning.

#### Scenario: Export profile
- **WHEN** `lango config export default` is run
- **THEN** the decrypted config is printed to stdout as formatted JSON, with a WARNING on stderr

### Requirement: Config validate command
The system SHALL provide a `lango config validate` command that validates the active profile's configuration.

#### Scenario: Valid config
- **WHEN** the active profile's config passes validation
- **THEN** the message "Profile \"default\" configuration is valid." is displayed
