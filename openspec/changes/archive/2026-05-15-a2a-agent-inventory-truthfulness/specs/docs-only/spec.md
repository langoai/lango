## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the current A2A and agent CLI surface
The public architecture inventory docs SHALL include the currently implemented A2A and agent diagnostics surfaces rather than omitting or abbreviating them to stale subsets.

#### Scenario: A2A and agent inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the inventories SHALL include A2A card/check
- **AND** they SHALL include the agent trace/graph diagnostics surface
- **AND** the README internal tree SHALL not keep a stale duplicate `chat` row
