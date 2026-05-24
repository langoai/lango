## ADDED Requirements

### Requirement: Alerts inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs omit the current alerts CLI surface SHALL be enforced by an executable test.

#### Scenario: Alerts inventory remains truthful
- **WHEN** the repository still ships `lango alerts list` and `lango alerts summary`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
