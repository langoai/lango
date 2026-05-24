## ADDED Requirements

### Requirement: Smart-account inventory truthfulness guard stays executable
Repository-level regressions that let smart-account inventory docs describe stale command subsets SHALL be enforced by an executable test.

#### Scenario: Smart-account inventory remains truthful
- **WHEN** the repository still ships `lango account session create/revoke`, `module install`, `policy set`, and `paymaster approve`
- **THEN** an executable repository test SHALL fail if `docs/cli/smartaccount.md`, `docs/architecture/project-structure.md`, or `README.md` stops describing those current command surfaces
