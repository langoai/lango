## ADDED Requirements

### Requirement: README graph inventory truthfulness guard stays executable
Repository-level regressions that let the README internal tree describe a stale graph CLI subset SHALL be enforced by an executable test.

#### Scenario: README graph inventory remains truthful
- **WHEN** the repository still ships `lango graph add/export/import`
- **THEN** an executable repository test SHALL fail if `README.md` stops describing those current command surfaces
