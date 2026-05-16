## ADDED Requirements

### Requirement: Architecture index links every architecture reference page
The public architecture index SHALL provide an explicit catalog of every dedicated page under `docs/architecture/` so deep-dive architecture references remain discoverable from the top-level architecture landing page.

#### Scenario: Architecture references stay linked from the index
- **WHEN** a maintainer updates `docs/architecture/index.md`
- **THEN** it SHALL include links to every dedicated page under `docs/architecture/` other than `index.md`
- **AND** that catalog SHALL cover pages such as `overview.md`, `project-structure.md`, `data-flow.md`, `knowledge-exchange-runtime.md`, `settlement-progression.md`, `actual-settlement-execution.md`, `retry-dead-letter-handling.md`, and `p2p-knowledge-exchange-track.md`
