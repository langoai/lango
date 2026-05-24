## ADDED Requirements

### Requirement: Session-store migration safety stays executable
Repository-level regressions that reintroduce concurrent Ent schema-migration crashes in the session store constructor SHALL be enforced by executable package tests.

#### Scenario: Concurrent NewEntStore remains safe
- **WHEN** the repository still ships `internal/session.NewEntStore`
- **THEN** executable package tests SHALL fail if concurrent `NewEntStore` invocations panic or return migration/open errors caused by unsynchronized Ent schema setup
