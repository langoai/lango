## MODIFIED Requirements

### Requirement: Encrypted config profile storage via onboard
The `lango onboard` command SHALL save configuration via `configstore.Store.Save()` to the encrypted SQLite profile store (`~/.lango/lango.db`) instead of writing plain-text `lango.json` via `config.Save()`.

#### Scenario: Save new profile via onboard
- **WHEN** user completes the onboard wizard and selects "Save & Exit"
- **THEN** the configuration SHALL be saved as an encrypted profile via `configstore.Store.Save()`
- **AND** no `lango.json` file SHALL be created

#### Scenario: Save to named profile
- **WHEN** user runs `lango onboard --profile myprofile` and saves
- **THEN** the configuration SHALL be saved under the profile name "myprofile"

#### Scenario: New profile activation
- **WHEN** a profile with the given name does not exist before onboard
- **THEN** the saved profile SHALL be activated via `configstore.Store.SetActive()`

#### Scenario: Existing profile not re-activated
- **WHEN** a profile with the given name already exists before onboard
- **THEN** the profile's active status SHALL remain unchanged after save

## ADDED Requirements

### Requirement: Profile flag for onboard command
The `lango onboard` command SHALL accept a `--profile` flag to specify the profile name to create or edit. The default value SHALL be "default".

#### Scenario: Default profile name
- **WHEN** user runs `lango onboard` without `--profile`
- **THEN** the wizard SHALL operate on the "default" profile

#### Scenario: Custom profile name
- **WHEN** user runs `lango onboard --profile staging`
- **THEN** the wizard SHALL operate on the "staging" profile

### Requirement: Pre-load existing profile into wizard
The onboard wizard SHALL load an existing profile's configuration as the initial form values when editing a returning user's profile. If no profile exists, the wizard SHALL use `config.DefaultConfig()`.

#### Scenario: Edit existing profile
- **WHEN** user runs `lango onboard` and a "default" profile exists
- **THEN** the wizard forms SHALL be pre-populated with the existing profile's values

#### Scenario: New user onboard
- **WHEN** user runs `lango onboard` and no "default" profile exists
- **THEN** the wizard forms SHALL be pre-populated with default config values

### Requirement: Bootstrap before TUI
The onboard command SHALL run `bootstrap.Run()` to initialize the database, crypto, and configstore before starting the BubbleTea TUI program. This ensures passphrase acquisition does not conflict with TUI terminal capture.

#### Scenario: Passphrase then TUI
- **WHEN** user runs `lango onboard`
- **THEN** the passphrase prompt SHALL appear before the TUI wizard starts

#### Scenario: Bootstrap failure
- **WHEN** bootstrap fails (e.g., DB error, wrong passphrase)
- **THEN** the onboard command SHALL return the bootstrap error without starting the TUI

### Requirement: Updated post-save messaging
After saving, the onboard command SHALL display the profile name, storage path (`~/.lango/lango.db`), and profile management commands (`lango config list`, `lango config use`).

#### Scenario: Post-save output
- **WHEN** user saves configuration via onboard
- **THEN** the output SHALL include the encrypted profile name and storage path
- **AND** the output SHALL include profile management command hints

### Requirement: Save menu text reflects encrypted storage
The "Save & Exit" menu item description SHALL read "Save encrypted profile" instead of "Write config to file".

#### Scenario: Menu description
- **WHEN** user views the configuration menu
- **THEN** the "Save & Exit" item description SHALL be "Save encrypted profile"
