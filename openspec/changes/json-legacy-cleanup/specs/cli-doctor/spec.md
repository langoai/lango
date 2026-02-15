## MODIFIED Requirements

### Requirement: Configuration File Check
The system SHALL verify that an encrypted configuration profile exists and is valid, instead of checking for a JSON file.

#### Scenario: Valid encrypted profile
- **WHEN** an active encrypted profile is loaded successfully via bootstrap
- **THEN** check passes with message "Encrypted configuration profile valid"

#### Scenario: No active profile loaded
- **WHEN** bootstrap fails to load an active profile but `lango.db` exists
- **THEN** check fails with message "No active configuration profile loaded" and suggestion to run `lango onboard`

#### Scenario: No profile database
- **WHEN** `~/.lango/lango.db` does not exist
- **THEN** check fails with message "Encrypted profile database not found" and is marked as fixable
- **AND** the fix action guides the user to run `lango onboard`

### Requirement: Auto-Fix Mode
The system SHALL support a `--fix` flag that attempts to automatically repair common issues.

#### Scenario: Fix creates missing database directory
- **WHEN** `--fix` is provided AND database directory does not exist
- **THEN** system creates the directory and reports "Created ~/.lango directory"

#### Scenario: Fix guides to onboard for missing profile
- **WHEN** `--fix` is provided AND no encrypted profile exists
- **THEN** system displays guidance: "Run 'lango onboard' to set up your configuration"

### Requirement: Doctor command description lists all checks
The `lango doctor` command Long description SHALL enumerate all diagnostic checks performed.

#### Scenario: Long description content
- **WHEN** user runs `lango doctor --help`
- **THEN** the description SHALL list: Encrypted configuration profile validity, API key and provider configuration, Channel token validation, Session database accessibility, Server port availability, Security configuration, Companion connectivity

## REMOVED Requirements

### Requirement: Doctor --config flag
**Reason**: The `--config` flag on the doctor command existed to specify a JSON config file path. Doctor now loads configuration via bootstrap from the encrypted profile database.
**Migration**: Configuration is loaded automatically from the encrypted profile. No flag needed.
