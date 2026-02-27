## REMOVED Requirements

### Requirement: Configuration
**Reason**: The `security.keyring.enabled` config flag is redundant — `keyring.IsAvailable()` runtime auto-detection is the sole mechanism used by bootstrap. The flag was never consulted at runtime.
**Migration**: Remove `security.keyring.enabled` from config files. No behavioral change — keyring availability was always determined by runtime detection.

## ADDED Requirements

### Requirement: Interactive keyring storage prompt
After a passphrase is acquired interactively (source is `SourceInteractive`) and an OS keyring provider is available, the system SHALL prompt the user to store the passphrase in the OS keyring for future automatic unlock.

#### Scenario: First run with keyring available
- **WHEN** user enters passphrase interactively AND OS keyring is available
- **THEN** system prompts "OS keyring is available. Store passphrase for automatic unlock? [y/N]"

#### Scenario: User accepts keyring storage
- **WHEN** user responds "y" to the keyring storage prompt
- **THEN** system stores the passphrase via `krProvider.Set(Service, KeyMasterPassphrase, pass)`

#### Scenario: User declines keyring storage
- **WHEN** user responds "N" or presses Enter to the keyring storage prompt
- **THEN** system proceeds without storing and does not prompt again until next interactive entry

#### Scenario: Keyring store failure
- **WHEN** user accepts but `krProvider.Set()` returns an error
- **THEN** system prints a warning to stderr and continues startup normally

#### Scenario: Non-interactive passphrase source
- **WHEN** passphrase is acquired from keyring, keyfile, or stdin pipe
- **THEN** system SHALL NOT display the keyring storage prompt

#### Scenario: Keyring unavailable
- **WHEN** OS keyring is not available (headless, CI, Docker)
- **THEN** system SHALL NOT display the keyring storage prompt

## MODIFIED Requirements

### Requirement: Configuration
The OS keyring integration SHALL NOT have a configuration flag. Keyring availability SHALL be determined solely by `keyring.IsAvailable()` runtime auto-detection.

#### Scenario: Keyring availability on supported OS
- **WHEN** the application starts on a system with an OS keyring daemon
- **THEN** `IsAvailable()` returns `Status{Available: true}` and the keyring is used as the highest-priority passphrase source

#### Scenario: Keyring unavailable in headless environment
- **WHEN** the application starts in a headless environment (CI, Docker, SSH)
- **THEN** `IsAvailable()` returns `Status{Available: false}` and the system silently falls back to keyfile or interactive prompt
