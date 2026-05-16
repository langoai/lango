## ADDED Requirements

### Requirement: Architecture and README inventory docs include current ontology-storage packages
The public inventory docs SHALL include the shipped ontology and storage foundation packages that implement ontology governance, shared SQLite opening, and storage-facade composition instead of omitting them from the package inventory.

#### Scenario: Ontology-storage rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `ontology/`, `sqlitedriver/`, and `storage/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe ontology governance/tooling, shared SQLite driver helpers, and broker-aware storage-facade composition truthfully
