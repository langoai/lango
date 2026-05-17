## MODIFIED Requirements

### Requirement: Configuration File Check
The system SHALL verify that an encrypted configuration profile exists and is valid, instead of checking for a JSON file. When doctor bootstrap fails before a configuration can be loaded, the doctor output SHALL include a dedicated failing bootstrap diagnostic that preserves the original bootstrap error.

#### Scenario: Valid encrypted profile
- **WHEN** an active encrypted profile is loaded successfully via bootstrap
- **THEN** check passes with message "Encrypted configuration profile valid"

#### Scenario: Bootstrap failure is surfaced
- **WHEN** bootstrap fails while `lango doctor` is loading the encrypted profile
- **THEN** doctor output SHALL include a failing "Bootstrap" diagnostic
- **AND** the diagnostic details SHALL include the original bootstrap error
- **AND** doctor SHALL continue running the remaining checks in best-effort mode

#### Scenario: No active profile loaded
- **WHEN** bootstrap fails to load an active profile but `lango.db` exists
- **THEN** check fails with message "No active configuration profile loaded" and suggestion to run `lango onboard`

#### Scenario: No profile database
- **WHEN** `~/.lango/lango.db` does not exist
- **THEN** check fails with message "Encrypted profile database not found" and is marked as fixable
- **AND** the fix action guides the user to run `lango onboard`
