## REMOVED Requirements

### Requirement: FallbackProvider wiring for macOS
**Reason**: The OSProvider fallback was used to read passphrase items stored without biometric ACL in the macOS Keychain. Since the plain OS keyring (go-keyring) has been removed due to same-UID attack vulnerability, this wiring is no longer needed.
**Migration**: Bootstrap no longer creates or passes a FallbackProvider. Passphrase acquisition falls through from hardware keyring to keyfile → interactive → stdin.

## MODIFIED Requirements

### Requirement: Report biometric passphrase store outcome
When the bootstrap flow stores a passphrase in the secure keyring provider, it SHALL report the outcome to stderr. On entitlement error (`ErrEntitlement`), the system SHALL warn the user and suggest codesigning instead of falling back to OSProvider. On other failures, the message SHALL be `warning: store passphrase failed: <error>`. On success, the message SHALL be `Passphrase saved. Next launch will load it automatically.`.

#### Scenario: Biometric store succeeds
- **WHEN** `secureProvider.Set()` returns nil
- **THEN** stderr SHALL contain `Passphrase saved. Next launch will load it automatically.`

#### Scenario: Biometric store fails with entitlement error
- **WHEN** `secureProvider.Set()` returns an error satisfying `errors.Is(err, keyring.ErrEntitlement)`
- **THEN** stderr SHALL contain `warning: biometric storage unavailable (binary not codesigned)`
- **AND** stderr SHALL contain a codesign tip

#### Scenario: Biometric store fails with non-entitlement error
- **WHEN** `secureProvider.Set()` returns an error NOT satisfying `errors.Is(err, keyring.ErrEntitlement)`
- **THEN** stderr SHALL contain `warning: store passphrase failed: <error detail>`
