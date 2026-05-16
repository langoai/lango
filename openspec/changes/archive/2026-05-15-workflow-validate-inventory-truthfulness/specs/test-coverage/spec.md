## ADDED Requirements

### Requirement: Workflow validate inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit the shipped workflow `validate` surface SHALL be enforced by an executable test.

#### Scenario: Workflow validate remains in inventory docs
- **WHEN** the repository still ships `lango workflow validate <file>`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing that current command surface
