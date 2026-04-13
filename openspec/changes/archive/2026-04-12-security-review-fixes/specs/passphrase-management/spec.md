## MODIFIED Requirements

### Requirement: Passphrase change updates stored credentials

#### Scenario: Keyring updated after passphrase change
- **WHEN** `lango security change-passphrase` succeeds
- **THEN** the command SHALL attempt to update the secure keyring with the new passphrase
- **AND** failure SHALL print a warning with manual fix instructions

#### Scenario: Keyfile updated after passphrase change
- **WHEN** `lango security change-passphrase` succeeds and a keyfile exists
- **THEN** the command SHALL write the new passphrase to the keyfile

#### Scenario: Recovery restore updates stored credentials
- **WHEN** `lango security recovery restore` succeeds
- **THEN** the same keyring and keyfile update logic SHALL apply as in passphrase change
