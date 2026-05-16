## MODIFIED Requirements

### Requirement: Report biometric passphrase store outcome
When the bootstrap flow stores a passphrase in the secure keyring provider, it SHALL report the outcome to stderr. On entitlement error (`ErrEntitlement`), the system SHALL warn the user and suggest codesigning. On other failures, the message SHALL be `warning: store passphrase failed: <error>`. On success, the message SHALL be `Passphrase saved. Next launch will load it automatically.`.

#### Scenario: Legacy migration banner uses bootstrap stderr path
- **WHEN** a legacy install triggers the one-time envelope migration
- **THEN** the `Upgrading encryption format (one-time migration)...` banner SHALL be emitted through the bootstrap stderr writer path

#### Scenario: Keyfile shred warning uses bootstrap stderr path
- **WHEN** keyfile shredding fails after successful crypto initialization
- **THEN** the warning SHALL be emitted through the bootstrap stderr writer path
