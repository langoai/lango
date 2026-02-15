## MODIFIED Requirements

### Requirement: Config import command
The system SHALL provide a `lango config import <file>` command that imports a JSON config file as an encrypted profile. The source JSON file SHALL be automatically deleted after successful import for security.

#### Scenario: Import JSON file
- **WHEN** `lango config import lango.json --profile migrated` is run
- **THEN** the JSON file is loaded, encrypted, and stored as profile "migrated" (set as active)
- **AND** the source JSON file is deleted after successful import
- **AND** the message "Source file deleted for security." is displayed

#### Scenario: Import with delete failure
- **WHEN** import succeeds but the source file cannot be deleted (e.g., permission denied)
- **THEN** a warning is logged but the command does not fail

### Requirement: Config export command
The system SHALL provide a `lango config export <name>` command that outputs decrypted config as JSON. Passphrase verification is required (handled implicitly by the bootstrap process).

#### Scenario: Export profile
- **WHEN** `lango config export default` is run
- **THEN** the passphrase is verified via bootstrap
- **AND** the decrypted config is printed to stdout as formatted JSON, with a WARNING on stderr

## REMOVED Requirements

### Requirement: Root --config flag
**Reason**: The `--config` persistent flag on the root command existed to specify a JSON migration path. With automatic JSON migration removed from bootstrap, this flag is unnecessary. Users import JSON explicitly via `lango config import`.
**Migration**: Use `lango config import <file>` instead of `--config <file>`.
