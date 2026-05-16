## ADDED Requirements

### Requirement: Remaining CLI inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped chat, extension, provenance, run, sandbox, or status CLI surfaces SHALL be enforced by an executable test.

#### Scenario: Remaining CLI inventory remains truthful
- **WHEN** the repository still ships chat, extension, provenance, run, sandbox, and status command families
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
