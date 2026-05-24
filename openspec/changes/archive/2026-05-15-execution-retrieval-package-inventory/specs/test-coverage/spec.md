## ADDED Requirements

### Requirement: Execution-retrieval infrastructure package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped execution-retrieval infrastructure packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Execution-retrieval infrastructure rows remain truthful
- **WHEN** the repository still ships `internal/agentrt`, `internal/gatekeeper`, `internal/retrieval`, `internal/search`, `internal/turnrunner`, `internal/turntrace`, `internal/lineio`, and `internal/storeutil`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits runtime coordination, sanitization, retrieval orchestration, FTS5 search substrate, turn execution/tracing, partial-line reading, or store helper responsibilities
