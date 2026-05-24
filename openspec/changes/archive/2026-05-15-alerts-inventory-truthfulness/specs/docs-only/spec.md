## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the current alerts CLI surface
The public architecture inventory docs SHALL include the currently implemented alerts command surface rather than omitting it.

#### Scenario: Alerts inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the alerts inventory SHALL include `list` and `summary`
