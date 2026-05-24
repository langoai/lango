## ADDED Requirements

### Requirement: Architecture index reference-catalog guard stays executable
Repository-level regressions that let the architecture landing page drop links to dedicated architecture references SHALL be enforced by an executable test.

#### Scenario: Every architecture reference remains linked
- **WHEN** the repository still ships dedicated architecture pages under `docs/architecture/`
- **THEN** an executable repository test SHALL fail if `docs/architecture/index.md` stops linking any of those pages
- **AND** it SHALL therefore catch omissions affecting references such as `overview.md`, `project-structure.md`, `data-flow.md`, `knowledge-exchange-runtime.md`, `settlement-progression.md`, `actual-settlement-execution.md`, `retry-dead-letter-handling.md`, or `p2p-knowledge-exchange-track.md`
