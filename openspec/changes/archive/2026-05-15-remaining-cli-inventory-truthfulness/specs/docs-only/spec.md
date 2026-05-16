## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the remaining shipped CLI surfaces
The public architecture inventory docs SHALL include the currently implemented chat, extension, provenance, run, sandbox, and status CLI families rather than omitting them.

#### Scenario: Remaining CLI inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** those inventories SHALL include chat, extension, provenance, run, sandbox, and status surfaces
