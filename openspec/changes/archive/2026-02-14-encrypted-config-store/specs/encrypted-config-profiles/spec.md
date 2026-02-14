## ADDED Requirements

### Requirement: Encrypted config profile storage
The system SHALL store application configuration as AES-256-GCM encrypted blobs in a SQLite database table (`config_profiles`). Each profile SHALL contain a UUID id, unique name, encrypted data, active flag, version counter, and timestamps.

#### Scenario: Save a new profile
- **WHEN** a config is saved with a name that does not exist
- **THEN** a new profile row is created with the encrypted config data and version 1

#### Scenario: Update an existing profile
- **WHEN** a config is saved with a name that already exists
- **THEN** the existing profile's encrypted data is updated and version is incremented

### Requirement: Profile load and decrypt
The system SHALL decrypt a named profile's data using the initialized CryptoProvider and deserialize it into a `config.Config` struct.

#### Scenario: Load an existing profile
- **WHEN** a profile with the given name exists
- **THEN** the encrypted data is decrypted and returned as a valid Config

#### Scenario: Load a non-existent profile
- **WHEN** no profile with the given name exists
- **THEN** the system returns `ErrProfileNotFound`

#### Scenario: Wrong passphrase decryption
- **WHEN** decryption is attempted with an incorrect passphrase
- **THEN** the system returns a decryption error

### Requirement: Active profile management
The system SHALL support exactly one active profile at a time. Activating a profile SHALL deactivate all others in a single transaction.

#### Scenario: Load active profile
- **WHEN** an active profile exists
- **THEN** the system returns the profile name and decrypted config

#### Scenario: No active profile
- **WHEN** no profile has active=true
- **THEN** the system returns `ErrNoActiveProfile`

#### Scenario: Switch active profile
- **WHEN** SetActive is called with a valid profile name
- **THEN** all profiles are deactivated and the named profile is activated

### Requirement: Profile listing without decryption
The system SHALL list all profiles with metadata (name, active, version, timestamps) without decrypting any data.

#### Scenario: List profiles
- **WHEN** profiles exist in the database
- **THEN** all profiles are returned with metadata, ordered by name

### Requirement: Profile deletion
The system SHALL delete profiles by name but SHALL NOT delete the currently active profile.

#### Scenario: Delete inactive profile
- **WHEN** a non-active profile is deleted
- **THEN** the profile is removed from the database

#### Scenario: Delete active profile
- **WHEN** deletion is attempted on the active profile
- **THEN** the system returns `ErrDeleteActive`
