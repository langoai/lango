## ADDED Requirements

### Requirement: README mission-projection package guard stays executable
Repository-level regressions that let the README internal package tree omit shipped mission-projection packages SHALL be enforced by an executable test.

#### Scenario: Mission projection packages remain visible
- **WHEN** the repository still ships `internal/proposal`, `internal/loopview`, and `internal/collabview`
- **THEN** an executable repository test SHALL fail if the README internal tree stops describing those packages and their current responsibilities
