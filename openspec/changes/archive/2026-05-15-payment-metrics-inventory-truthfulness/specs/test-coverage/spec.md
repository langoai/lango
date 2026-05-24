## ADDED Requirements

### Requirement: Payment/metrics inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe stale payment or metrics CLI surfaces SHALL be enforced by an executable test.

#### Scenario: Payment and metrics inventory remains truthful
- **WHEN** the repository still ships `lango payment x402` and `lango metrics policy`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
