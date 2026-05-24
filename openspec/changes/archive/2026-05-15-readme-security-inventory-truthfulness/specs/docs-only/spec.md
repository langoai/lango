## ADDED Requirements

### Requirement: README internal tree stays aligned with the current security CLI surface
The README internal CLI inventory SHALL describe the current security command families instead of collapsing canonical, deprecated, recovery, and KMS surfaces into stale shorthand.

#### Scenario: README security inventory stays truthful
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe canonical `change-passphrase` and deprecated `migrate-passphrase`
- **AND** it SHALL continue to mention `secrets`, `keyring store/clear/status`, `recovery setup/restore`, `kms status/test/keys/wrap/detach`, and legacy `db-*` tombstones
