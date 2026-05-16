## ADDED Requirements

### Requirement: Operational-support docs describe dbopen migration serialization truthfully
The public package inventory SHALL not imply that `dbopen` performs fully parallel-safe Ent migration when managed opens are intentionally serialized to avoid Atlas concurrency hazards.

#### Scenario: dbopen row stays truthful
- **WHEN** a maintainer updates the `dbopen/` package rows in `README.md` or `docs/architecture/project-structure.md`
- **THEN** those rows SHALL continue to describe managed read-write and read-only database opening
- **AND** they SHALL mention serialized schema migration rather than implying unconstrained parallel Ent migration
