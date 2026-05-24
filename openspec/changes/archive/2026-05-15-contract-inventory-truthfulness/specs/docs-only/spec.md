## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the current contract CLI surface
The public architecture inventory docs SHALL include the currently implemented contract command surface rather than truncated stale shorthand.

#### Scenario: Contract inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the architecture inventory SHALL include `abi load`
- **AND** the README internal tree SHALL describe `lango contract read/call/abi load`
