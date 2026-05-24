## ADDED Requirements
### Requirement: Package-consolidation deleted-path guard stays executable
Repository-level regressions that reintroduce deleted consolidated package-directory paths into the `package-consolidation` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted consolidated package paths are rejected
- **WHEN** the current packages already live in `internal/types`, `internal/security/passphrase`, and `internal/p2p/zkp`
- **THEN** an executable repository test SHALL fail if the `package-consolidation` main spec claims `internal/ctxutil/`, `internal/passphrase/`, or `internal/zkp/` are still the current package locations
