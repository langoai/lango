## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the current P2P CLI surface
The public architecture inventory docs SHALL include the currently implemented P2P workspace, git, provenance, team, and ZKP surfaces rather than outdated subsets.

#### Scenario: P2P inventory rows stay truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the P2P inventory SHALL include workspace, git, provenance, team, and ZKP command families
