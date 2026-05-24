## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the current memory CLI surface
The public architecture inventory docs SHALL include the currently implemented observational and per-agent memory command surface rather than outdated subsets.

#### Scenario: Memory inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the memory inventory SHALL include `agents` and `agent <name>`
