## ADDED Requirements

### Requirement: Engine shutdown
The `Engine` SHALL provide a `Shutdown()` method that cancels all running workflow executions via the cancels map.

#### Scenario: Graceful shutdown
- **WHEN** `Shutdown()` is called
- **THEN** all cancel functions in the cancels map SHALL be invoked

### Requirement: RunStatus StartedAt field
The `RunStatus` struct SHALL include a `StartedAt time.Time` field populated from the workflow run record.

#### Scenario: StartedAt in GetRunStatus
- **WHEN** `GetRunStatus()` is called
- **THEN** the returned `RunStatus` SHALL include the `StartedAt` timestamp from the database record

#### Scenario: StartedAt in ListRuns
- **WHEN** `ListRuns()` is called
- **THEN** each returned `RunStatus` SHALL include the `StartedAt` timestamp
