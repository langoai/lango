## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the current economy CLI surface
The public architecture inventory docs SHALL describe the current economy command surface using the implemented `... status` paths instead of stale family-only shorthand.

#### Scenario: Economy inventory stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the architecture inventory SHALL include `budget status`, `risk status`, `pricing status`, `negotiate status`, and `escrow status/list/show/sentinel status`
- **AND** the README internal tree SHALL describe `lango economy budget status/risk status/pricing status/negotiate status/escrow status/list/show/sentinel status`
