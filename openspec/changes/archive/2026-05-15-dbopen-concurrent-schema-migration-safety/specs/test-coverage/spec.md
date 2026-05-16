## ADDED Requirements

### Requirement: Concurrent dbopen migration safety stays executable
Repository-level regressions that reintroduce the concurrent `OpenManaged` migration crash SHALL be enforced by executable package tests.

#### Scenario: Concurrent OpenManaged remains safe
- **WHEN** the repository still ships `internal/dbopen.OpenManaged`
- **THEN** executable package tests SHALL fail if concurrent `OpenManaged` invocations panic or return migration/open errors caused by unsynchronized Ent schema setup
