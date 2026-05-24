## ADDED Requirements

### Requirement: Shared CLI support inventory guard stays executable
Repository-level regressions that let the public inventory docs omit shared CLI support packages or misdescribe their responsibilities SHALL be enforced by an executable test.

#### Scenario: Shared CLI support packages remain visible
- **WHEN** the repository still ships `internal/cli/cliboot` and `internal/cli/clihttp`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or the README internal tree stops describing those packages and their current responsibilities
