## ADDED Requirements

### Requirement: Economy inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe stale economy family shorthand instead of the current status-oriented command surface SHALL be enforced by an executable test.

#### Scenario: Economy inventory remains truthful
- **WHEN** the repository still ships `lango economy budget status`, `risk status`, `pricing status`, `negotiate status`, and `escrow status/list/show/sentinel status`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
