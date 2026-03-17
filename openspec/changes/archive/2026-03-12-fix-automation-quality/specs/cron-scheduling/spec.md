## MODIFIED Requirements

### Requirement: Cron job persistence
The system SHALL persist cron jobs in the Ent ORM with fields: id (UUID), name (unique), schedule_type (at/every/cron), schedule, prompt, session_mode, deliver_to ([]string), timezone, enabled, last_run_at, next_run_at, and timestamps.

The `Store.Upsert` method SHALL return `(*Job, bool, error)` where the first return value is the persisted job (with generated ID and defaults populated), the second indicates whether an existing job was updated, and the third is any error.

#### Scenario: Create a cron job
- **WHEN** a cron job is created with name "news-summary", schedule "0 9 * * *", and prompt "Summarize today's news"
- **THEN** the job SHALL be persisted in the database with enabled=true and schedule_type="cron"

#### Scenario: Upsert returns persisted job on create
- **WHEN** `Upsert` is called for a job name that does not exist
- **THEN** the method SHALL create the job, read it back to populate the generated ID, and return `(*Job, false, nil)`

#### Scenario: Upsert returns persisted job on update
- **WHEN** `Upsert` is called for a job name that already exists
- **THEN** the method SHALL update the existing job, preserving its ID and CreatedAt, and return `(*Job, true, nil)`

#### Scenario: Unique name constraint
- **WHEN** a cron job is created with a name that already exists
- **THEN** the system SHALL update the existing job via Upsert rather than returning an error

### Requirement: Job lifecycle management
The system SHALL support adding, removing, pausing, and resuming cron jobs at runtime without restarting the scheduler.

`AddJob` SHALL use the `*Job` returned by `Upsert` directly, without an additional `GetByName` query.

#### Scenario: Pause a running job
- **WHEN** a job is paused via PauseJob()
- **THEN** the job SHALL be marked as disabled and removed from the cron runner

#### Scenario: Resume a paused job
- **WHEN** a paused job is resumed via ResumeJob()
- **THEN** the job SHALL be re-registered with the cron runner and marked as enabled

#### Scenario: Remove a job
- **WHEN** a job is removed via RemoveJob()
- **THEN** the job SHALL be deleted from the database and unregistered from the cron runner

#### Scenario: AddJob creates new job
- **WHEN** AddJob is called with a new job name
- **THEN** the scheduler SHALL upsert the job, register it with the cron runner, and return `(false, nil)`

#### Scenario: AddJob updates existing job
- **WHEN** AddJob is called with an existing job name
- **THEN** the scheduler SHALL upsert the job, unregister the old entry, re-register with the new schedule, and return `(true, nil)`

## ADDED Requirements

### Requirement: Idempotent scheduler shutdown
The `Scheduler.Stop()` method SHALL be idempotent — calling it multiple times SHALL NOT panic. The method SHALL use `sync.Once` to ensure the shutdown sequence (close shutdownCh, drain cron entries, clear entries map) executes exactly once.

#### Scenario: Stop called once
- **WHEN** `Stop()` is called on a started scheduler
- **THEN** the scheduler SHALL close shutdownCh, wait for the cron runner to drain, clear entries, and log "cron scheduler stopped"

#### Scenario: Stop called twice
- **WHEN** `Stop()` is called twice on the same scheduler
- **THEN** the second call SHALL be a no-op without panic

#### Scenario: Stop called without Start
- **WHEN** `Stop()` is called on a scheduler that was never started (cron is nil)
- **THEN** the method SHALL be a no-op without panic

### Requirement: One-time job unregistration is single-responsibility
The `disableOneTimeJob` method SHALL only handle DB persistence (setting `Enabled=false`). It SHALL NOT call `unregisterJob`, as the `sync.Once` wrapper in `registerJob` already handles cron entry removal for one-time jobs.

#### Scenario: One-time job disabled after execution
- **WHEN** a one-time ("at") job fires
- **THEN** `registerJob`'s sync.Once wrapper SHALL call `unregisterJob`, and `disableOneTimeJob` SHALL only update the job's Enabled flag to false in the database
