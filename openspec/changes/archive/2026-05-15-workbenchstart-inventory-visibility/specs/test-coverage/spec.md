## ADDED Requirements

### Requirement: Workbenchstart inventory guard stays executable
Repository-level regressions that let the public inventory docs omit the shipped `workbenchstart` support package SHALL be enforced by an executable test.

#### Scenario: Workbenchstart package remains visible
- **WHEN** the repository still ships `internal/cli/workbenchstart`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or the README internal tree stops describing that package and its current responsibilities
