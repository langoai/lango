## ADDED Requirements
### Requirement: Package-consolidation spec stays truthful about current package locations
The `package-consolidation` main spec SHALL not keep deleted package-directory paths once the consolidation is complete and the current packages already live elsewhere.

#### Scenario: Deleted consolidated package paths are rejected
- **WHEN** a maintainer updates the `package-consolidation` main spec
- **THEN** it SHALL not claim `internal/ctxutil/`, `internal/passphrase/`, or `internal/zkp/` as current package locations
