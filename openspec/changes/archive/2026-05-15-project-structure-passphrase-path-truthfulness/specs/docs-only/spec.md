## ADDED Requirements

### Requirement: Architecture project-structure docs stay aligned with the current passphrase package path
The public architecture project-structure reference SHALL not keep the deleted top-level `passphrase/` package path once passphrase helpers live under the security subtree.

#### Scenario: Project-structure passphrase row stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** it SHALL describe `security/passphrase/`
- **AND** it SHALL NOT reintroduce the deleted `passphrase/` package path
