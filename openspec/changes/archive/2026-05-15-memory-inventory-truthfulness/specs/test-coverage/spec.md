## ADDED Requirements

### Requirement: Memory inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe a stale memory CLI subset SHALL be enforced by an executable test.

#### Scenario: Memory inventory remains truthful
- **WHEN** the repository still ships `lango memory agents` and `lango memory agent <name>`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
