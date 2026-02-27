## MODIFIED Requirements

### Requirement: Keyring provider is nil when no secure hardware is available
The passphrase acquisition flow SHALL receive a nil `KeyringProvider` when the bootstrap determines no secure hardware backend is available (`TierNone`). This effectively disables keyring auto-read, forcing keyfile or interactive/stdin acquisition.

#### Scenario: Nil keyring provider skips keyring step
- **WHEN** `Acquire()` is called with `KeyringProvider` set to nil
- **THEN** the keyring step SHALL be skipped entirely, and acquisition SHALL proceed to keyfile or interactive prompt

#### Scenario: Secure keyring provider attempts read
- **WHEN** `Acquire()` is called with a non-nil `KeyringProvider` (biometric or TPM)
- **THEN** it SHALL attempt to read the passphrase from the secure provider first
