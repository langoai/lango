## MODIFIED Requirements

### Requirement: Report biometric passphrase store outcome
When the bootstrap flow stores a passphrase in the secure keyring provider, it SHALL report the outcome to stderr. On entitlement error (`ErrEntitlement`), the system SHALL warn the user and suggest codesigning. On other failures, the message SHALL be `warning: store passphrase failed: <error>`. On success, the message SHALL be `Passphrase saved. Next launch will load it automatically.`.

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
