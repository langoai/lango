## ADDED Requirements

### Requirement: Architecture and README inventory docs include current execution-retrieval infrastructure packages
The public inventory docs SHALL include the shipped execution and retrieval infrastructure packages that implement runtime coordination, response sanitization, retrieval orchestration, search substrate, turn execution/tracing, and store/line helpers instead of omitting them from the package inventory.

#### Scenario: Execution-retrieval infrastructure rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `agentrt/`, `gatekeeper/`, `retrieval/`, `search/`, `turnrunner/`, `turntrace/`, `lineio/`, and `storeutil/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe runtime coordination, response sanitization, fact/temporal retrieval orchestration, domain-agnostic FTS5 search, turn execution/tracing, partial-line reading, and store copy/JSON helpers truthfully
