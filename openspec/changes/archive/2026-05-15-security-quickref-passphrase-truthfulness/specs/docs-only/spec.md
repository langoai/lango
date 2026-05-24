## ADDED Requirements

### Requirement: Security quick references distinguish canonical and deprecated passphrase rotation paths
Public quick-reference docs SHALL describe `lango security change-passphrase` as the canonical passphrase-rotation path and `lango security migrate-passphrase` as the deprecated legacy full re-encryption path.

#### Scenario: Passphrase rotation quick-reference wording stays truthful
- **WHEN** a maintainer updates `README.md` or `docs/cli/index.md`
- **THEN** those quick references SHALL describe `lango security change-passphrase` as a non-reencrypting passphrase change
- **AND** they SHALL mark `lango security migrate-passphrase` as deprecated legacy migration
