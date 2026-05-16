## ADDED Requirements

### Requirement: Ontology-storage package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped ontology-storage packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Ontology-storage rows remain truthful
- **WHEN** the repository still ships `internal/ontology`, `internal/sqlitedriver`, and `internal/storage`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits ontology governance/tooling, SQLite driver helper behavior, or storage-facade/broker-adapter composition responsibilities
