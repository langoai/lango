## ADDED Requirements

### Requirement: Public quick references include implemented security commands
The public quick-reference docs SHALL include the implemented `lango security` command family that is already present in dedicated security docs.

#### Scenario: Implemented security commands stay discoverable
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL include the implemented `lango security` status, `change-passphrase`, deprecated `migrate-passphrase`, secrets, keyring, recovery, legacy db, and kms command entries
