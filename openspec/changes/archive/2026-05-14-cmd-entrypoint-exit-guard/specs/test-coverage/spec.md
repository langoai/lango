## ADDED Requirements

### Requirement: Cmd entrypoint exit guards stay executable
Repository-level top-level entrypoint exit regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Cmd entrypoint exit regressions are rejected
- **WHEN** `cmd/` production code reintroduces direct `os.Exit(...)` references outside explicit seam declarations
- **THEN** an executable repository test SHALL fail
