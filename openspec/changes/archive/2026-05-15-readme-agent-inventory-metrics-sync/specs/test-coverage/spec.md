## ADDED Requirements

### Requirement: A2A/agent inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped A2A or agent diagnostics surfaces, or keep stale duplicate chat inventory rows, SHALL be enforced by an executable test.

#### Scenario: A2A and agent inventory remains truthful
- **WHEN** the repository still ships A2A card/check and agent trace/graph diagnostics commands
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
- **AND** it SHALL fail if the README internal tree reintroduces the stale duplicate `chat` inventory row
