## MODIFIED Requirements

### Requirement: Bootstrap uses secure hardware provider for passphrase storage
The bootstrap process SHALL use `DetectSecureProvider()` to determine the keyring provider for passphrase acquisition. When no secure hardware is available (`TierNone`), the keyring provider SHALL be nil, disabling automatic keyring reads.

#### Scenario: Biometric available during bootstrap
- **WHEN** bootstrap runs on macOS with Touch ID
- **THEN** the passphrase acquisition SHALL use `BiometricProvider` as the keyring provider

#### Scenario: No secure hardware during bootstrap
- **WHEN** bootstrap runs on a system without biometric or TPM
- **THEN** the keyring provider SHALL be nil, and passphrase SHALL be acquired from keyfile or interactive prompt only

#### Scenario: Interactive passphrase with secure storage offer
- **WHEN** the passphrase source is interactive and a secure provider is available
- **THEN** the system SHALL offer to store the passphrase in the secure backend with a confirmation prompt showing the tier label

### Requirement: SkipSecureDetection option for testing
The `Options` struct SHALL include a `SkipSecureDetection` boolean. When true, secure hardware detection SHALL be skipped and the keyring provider SHALL be nil regardless of available hardware.

#### Scenario: SkipSecureDetection in test
- **WHEN** `Run()` is called with `SkipSecureDetection: true`
- **THEN** the bootstrap SHALL not probe for biometric or TPM hardware
