## MODIFIED Requirements

### Requirement: Non-interactive mini-bootstrap for status

#### Scenario: Keyring provider passed to non-interactive acquisition
- **WHEN** `readDBStatusNonInteractive` acquires a passphrase
- **THEN** it SHALL pass `keyring.DetectSecureProvider()` as the `KeyringProvider` option

#### Scenario: Active config loaded when DB available
- **WHEN** `readDBStatusNonInteractive` successfully opens the DB with MK
- **THEN** it SHALL load the active config profile via `configstore.Store.LoadActive`
- **AND** the loaded config SHALL be used for status display instead of `DefaultConfig()`

#### Scenario: Keyfile fallback on stale keyring passphrase
- **WHEN** envelope unwrap fails with a keyring-sourced passphrase
- **THEN** `readDBStatusNonInteractive` SHALL retry with a keyfile-only acquisition
- **AND** legacy DB open failure with keyring passphrase SHALL also retry with keyfile
