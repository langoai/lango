## ADDED Requirements

### Requirement: History-aware prompt enrichment
The executor SHALL enrich cron job prompts with recent execution history before sending to the agent runner. The system SHALL query up to 10 recent history entries for the job, prepend them as a "previous outputs" section instructing the LLM not to repeat them, and truncate each result preview to 200 characters. If the history query fails, the executor SHALL gracefully fall back to the original prompt without enrichment.

#### Scenario: Prompt enriched with history
- **WHEN** a cron job executes and has 3 previous history entries
- **THEN** the executor SHALL prepend a "Previous outputs — do NOT repeat these" section listing all 3 results before the original prompt

#### Scenario: No history available
- **WHEN** a cron job executes for the first time with no history entries
- **THEN** the executor SHALL use the original prompt without modification

#### Scenario: History query failure
- **WHEN** the history query returns an error
- **THEN** the executor SHALL log a debug-level message and use the original prompt without modification

#### Scenario: History saved with original prompt
- **WHEN** a cron job executes with history enrichment
- **THEN** the history entry SHALL record the original prompt (not the enriched version) to prevent prefix accumulation

### Requirement: In-flight execution cancellation
The scheduler SHALL track currently executing jobs via an `inFlight` map of jobID to context.CancelFunc. When `RemoveJob()` or `PauseJob()` is called, the scheduler SHALL cancel any in-flight execution for that job in addition to unregistering it from the cron runner. When `Stop()` is called, the scheduler SHALL cancel all in-flight executions before stopping the cron runner.

#### Scenario: Remove cancels in-flight execution
- **WHEN** `RemoveJob()` is called while the job is executing
- **THEN** the scheduler SHALL cancel the execution context AND delete the job from the store

#### Scenario: Pause cancels in-flight execution
- **WHEN** `PauseJob()` is called while the job is executing
- **THEN** the scheduler SHALL cancel the execution context AND mark the job as disabled

#### Scenario: Stop cancels all in-flight executions
- **WHEN** `Stop()` is called with jobs currently executing
- **THEN** the scheduler SHALL cancel all in-flight execution contexts before waiting for the cron runner to drain

#### Scenario: In-flight map cleanup
- **WHEN** a job execution completes normally
- **THEN** the scheduler SHALL remove the job's entry from the inFlight map via defer

### Requirement: Name-or-ID job resolution
The scheduler SHALL provide a `ResolveJobID(ctx, nameOrID) (string, error)` method that accepts either a UUID string or a job name. If the input is a valid UUID, it SHALL be returned as-is. Otherwise, the scheduler SHALL look up the job by name via `store.GetByName()` and return the job's ID.

#### Scenario: Resolve by UUID
- **WHEN** `ResolveJobID` is called with a valid UUID string
- **THEN** the method SHALL return the UUID without a store query

#### Scenario: Resolve by name
- **WHEN** `ResolveJobID` is called with a non-UUID string matching an existing job name
- **THEN** the method SHALL return the job's UUID from the store

#### Scenario: Name not found
- **WHEN** `ResolveJobID` is called with a non-UUID string that does not match any job name
- **THEN** the method SHALL return an error

## MODIFIED Requirements

### Requirement: Job lifecycle management
The system SHALL support adding, removing, pausing, and resuming cron jobs at runtime without restarting the scheduler.

`AddJob` SHALL use the `*Job` returned by `Upsert` directly, without an additional `GetByName` query.

The `cron_remove`, `cron_pause`, and `cron_resume` tool handlers SHALL accept either a job ID (UUID) or job name, using `scheduler.ResolveJobID()` to resolve names to IDs before calling the scheduler methods.

#### Scenario: Pause a running job
- **WHEN** a job is paused via PauseJob()
- **THEN** the job SHALL be marked as disabled, removed from the cron runner, and any in-flight execution SHALL be cancelled

#### Scenario: Resume a paused job
- **WHEN** a paused job is resumed via ResumeJob()
- **THEN** the job SHALL be re-registered with the cron runner and marked as enabled

#### Scenario: Remove a job
- **WHEN** a job is removed via RemoveJob()
- **THEN** the job SHALL be deleted from the database, unregistered from the cron runner, and any in-flight execution SHALL be cancelled

#### Scenario: Remove by name
- **WHEN** `cron_remove` is called with a job name instead of UUID
- **THEN** the handler SHALL resolve the name to a UUID via `ResolveJobID` and proceed with removal

#### Scenario: AddJob creates new job
- **WHEN** AddJob is called with a new job name
- **THEN** the scheduler SHALL upsert the job, register it with the cron runner, and return `(false, nil)`

#### Scenario: AddJob updates existing job
- **WHEN** AddJob is called with an existing job name
- **THEN** the scheduler SHALL upsert the job, unregister the old entry, re-register with the new schedule, and return `(true, nil)`
