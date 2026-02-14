## ADDED Requirements

### Requirement: Passphrase acquisition priority chain
The system SHALL acquire a passphrase using the following priority: (1) keyfile at `~/.lango/keyfile`, (2) interactive terminal prompt, (3) stdin pipe. The system SHALL return an error if no source is available.

#### Scenario: Keyfile exists with correct permissions
- **WHEN** a keyfile exists at the configured path with 0600 permissions
- **THEN** the passphrase is read from the file and `SourceKeyfile` is returned

#### Scenario: Keyfile has wrong permissions
- **WHEN** a keyfile exists but does not have 0600 permissions
- **THEN** the keyfile is skipped and the next source is tried

#### Scenario: Interactive terminal available
- **WHEN** no keyfile is available and stdin is a terminal
- **THEN** the user is prompted for a passphrase via hidden input and `SourceInteractive` is returned

#### Scenario: New passphrase creation
- **WHEN** `AllowCreation` is true and interactive terminal is used
- **THEN** the user is prompted twice (entry + confirmation) and the passphrase must match

#### Scenario: Stdin pipe
- **WHEN** no keyfile is available and stdin is a pipe (not a terminal)
- **THEN** one line is read from stdin and `SourceStdin` is returned

#### Scenario: No source available
- **WHEN** no keyfile exists, stdin is not a terminal, and stdin pipe is empty
- **THEN** the system returns an error

### Requirement: Keyfile management
The system SHALL read and write keyfiles with strict 0600 permission enforcement.

#### Scenario: Write keyfile
- **WHEN** a keyfile is written
- **THEN** the file is created with 0600 permissions and parent directories with 0700

#### Scenario: Read keyfile with valid permissions
- **WHEN** a keyfile is read with 0600 permissions
- **THEN** the passphrase is returned with trailing whitespace trimmed

#### Scenario: Read keyfile with invalid permissions
- **WHEN** a keyfile exists with permissions other than 0600
- **THEN** the system returns a permission validation error
