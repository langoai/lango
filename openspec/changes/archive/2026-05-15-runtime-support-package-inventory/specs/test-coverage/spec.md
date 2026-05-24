## ADDED Requirements

### Requirement: Runtime-support package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped runtime-support packages or misdescribe their current responsibilities SHALL be enforced by an executable test.

#### Scenario: Runtime-support package rows remain truthful
- **WHEN** the repository still ships `internal/exportability`, `internal/knowledgeruntime`, `internal/receipts`, `internal/storagebroker`, `internal/streamx`, `internal/tooloutput`, and `internal/toolparam`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits runtime branch selection, receipt progression, storage brokering, stream combinators, tool output retention, or typed tool-parameter extraction
