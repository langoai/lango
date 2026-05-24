## ADDED Requirements

### Requirement: Operational-support package inventory guard stays executable
Repository-level regressions that let architecture or README inventory docs omit shipped operational-support packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Operational-support rows remain truthful
- **WHEN** the repository still ships `internal/alerting`, `internal/approvalflow`, `internal/archtest`, and `internal/dbopen`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those packages and their current responsibilities
- **AND** it SHALL fail if the docs fall back to stale or generic wording that omits alerting thresholds/delivery, artifact release decision mapping, architecture enforcement testing, or managed read-write/read-only database opening responsibilities
