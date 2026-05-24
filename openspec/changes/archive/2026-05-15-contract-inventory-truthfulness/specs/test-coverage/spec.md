## ADDED Requirements

### Requirement: Contract inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe a stale contract CLI subset SHALL be enforced by an executable test.

#### Scenario: Contract inventory remains truthful
- **WHEN** the repository still ships `lango contract read`, `lango contract call`, and `lango contract abi load`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
