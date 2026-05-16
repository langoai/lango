## ADDED Requirements

### Requirement: Architecture security CLI truthfulness guard stays executable
Repository-level regressions that let the architecture project-structure docs describe a stale security CLI surface SHALL be enforced by an executable test.

#### Scenario: Project-structure security row remains truthful
- **WHEN** the repository still ships canonical `lango security change-passphrase` and deprecated `migrate-passphrase`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` stops describing that canonical/deprecated distinction
