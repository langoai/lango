## ADDED Requirements

### Requirement: Architecture passphrase package-path guard stays executable
Repository-level regressions that let the architecture project-structure docs reintroduce the deleted top-level `passphrase/` package path SHALL be enforced by an executable test.

#### Scenario: Project-structure passphrase row remains truthful
- **WHEN** the repository still ships passphrase helpers under `internal/security/passphrase`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` reintroduces `passphrase/` instead of `security/passphrase/`
