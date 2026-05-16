## ADDED Requirements

### Requirement: Architecture and README inventory docs include current operational-support packages
The public inventory docs SHALL include the shipped operational-support packages that implement alerting, canonical artifact approval flow, architecture enforcement tests, and managed database opening instead of omitting them from the package inventory.

#### Scenario: Operational-support rows stay visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal package tree
- **THEN** the architecture inventory SHALL include `alerting/`, `approvalflow/`, `archtest/`, and `dbopen/`
- **AND** the README internal tree SHALL include those same package rows
- **AND** those rows SHALL describe threshold-based alerting, artifact release decision mapping, architecture boundary/bootstrap enforcement tests, and managed read-write/read-only database opening truthfully
