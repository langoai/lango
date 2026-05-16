## ADDED Requirements

### Requirement: Architecture project-structure docs stay aligned with the current security CLI surface
The public architecture project-structure reference SHALL describe `cli/security/` using the current canonical `change-passphrase` command and mark `migrate-passphrase` as deprecated.

#### Scenario: Project-structure security row stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** the `cli/security/` row SHALL include `change-passphrase`
- **AND** it SHALL describe `migrate-passphrase` as deprecated legacy surface rather than as the primary passphrase-rotation command
- **AND** it SHALL continue to mention `keyring store/clear/status`, `recovery setup/restore`, and `kms status/test/keys/wrap/detach`
