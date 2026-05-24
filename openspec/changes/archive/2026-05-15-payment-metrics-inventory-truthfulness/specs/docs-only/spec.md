## ADDED Requirements

### Requirement: Architecture and README inventory docs stay aligned with the current payment and metrics CLI surface
The public architecture inventory docs SHALL include the currently implemented `payment x402` and `metrics policy` surfaces rather than outdated subsets.

#### Scenario: Payment and metrics inventory rows stay truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** the payment inventory SHALL include `x402`
- **AND** the metrics inventory SHALL include `policy`
