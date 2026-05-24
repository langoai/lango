## ADDED Requirements

### Requirement: Architecture graph/metrics CLI truthfulness guard stays executable
Repository-level regressions that let the architecture project-structure docs describe stale graph or metrics CLI surfaces SHALL be enforced by an executable test.

#### Scenario: Project-structure graph and metrics rows remain truthful
- **WHEN** the repository still ships `lango graph add/export/import` and `lango metrics policy`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` stops describing those current command surfaces
