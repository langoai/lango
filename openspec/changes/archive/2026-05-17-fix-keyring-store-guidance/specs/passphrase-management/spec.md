## MODIFIED Requirements

### Requirement: Passphrase change updates stored credentials
Successful passphrase change or recovery restore SHALL update any stored
keyring or keyfile credentials so local unlock paths remain consistent with the
new passphrase.

#### Scenario: Keyring updated after passphrase change
- **WHEN** `lango security change-passphrase` succeeds
- **THEN** the command SHALL attempt to update the secure keyring with the new
  passphrase
- **AND** failure SHALL print a warning with manual fix instructions
- **AND** the manual fix instructions SHALL point to
  `lango security keyring store`
- **AND** the manual fix instructions SHALL NOT point to nonexistent keyring
  subcommands

#### Scenario: Recovery restore updates stored credentials
- **WHEN** `lango security recovery restore` succeeds
- **THEN** the same keyring and keyfile update logic SHALL apply as in
  passphrase change
- **AND** keyring update failure guidance SHALL point to
  `lango security keyring store`
