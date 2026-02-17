## Purpose

Define the cron scheduling system that enables periodic, interval-based, and one-time agent task execution with persistent job storage and multi-channel result delivery.

## Requirements

### Requirement: Cron job persistence
The system SHALL persist cron jobs in the Ent ORM with fields: id (UUID), name (unique), schedule_type (at/every/cron), schedule, prompt, session_mode, deliver_to ([]string), timezone, enabled, last_run_at, next_run_at, and timestamps.

#### Scenario: Create a cron job
- **WHEN** a cron job is created with name "news-summary", schedule "0 9 * * *", and prompt "Summarize today's news"
- **THEN** the job SHALL be persisted in the database with enabled=true and schedule_type="cron"

#### Scenario: Unique name constraint
- **WHEN** a cron job is created with a name that already exists
- **THEN** the system SHALL return an error indicating the name is already taken

### Requirement: Schedule type support
The system SHALL support three schedule types: "cron" (standard cron expressions), "every" (interval durations like "1h"), and "at" (one-time ISO 8601 timestamps).

#### Scenario: Cron expression schedule
- **WHEN** a job is created with schedule_type "cron" and schedule "0 9 * * *"
- **THEN** the scheduler SHALL execute the job daily at 09:00 in the configured timezone

#### Scenario: Interval schedule
- **WHEN** a job is created with schedule_type "every" and schedule "1h"
- **THEN** the scheduler SHALL execute the job every hour

#### Scenario: One-time schedule
- **WHEN** a job is created with schedule_type "at" and schedule "2026-02-20T15:00:00"
- **THEN** the scheduler SHALL execute the job once at the specified time

### Requirement: Job lifecycle management
The system SHALL support adding, removing, pausing, and resuming cron jobs at runtime without restarting the scheduler.

#### Scenario: Pause a running job
- **WHEN** a job is paused via PauseJob()
- **THEN** the job SHALL be marked as disabled and removed from the cron runner

#### Scenario: Resume a paused job
- **WHEN** a paused job is resumed via ResumeJob()
- **THEN** the job SHALL be re-registered with the cron runner and marked as enabled

#### Scenario: Remove a job
- **WHEN** a job is removed via RemoveJob()
- **THEN** the job SHALL be deleted from the database and unregistered from the cron runner

### Requirement: Isolated session execution
The system SHALL execute each cron job in an isolated agent session with a key following the pattern "cron:<jobName>:<timestamp>".

#### Scenario: Job execution creates isolated session
- **WHEN** a cron job executes
- **THEN** the executor SHALL create a new session with key "cron:<name>:<unix-timestamp>" and run the prompt in that session

### Requirement: Multi-channel delivery
The system SHALL deliver job results to configured target channels (Telegram, Slack, Discord) via the ChannelSender interface.

#### Scenario: Deliver to multiple channels
- **WHEN** a job completes with deliver_to ["slack", "telegram"]
- **THEN** the delivery system SHALL send the formatted result to both Slack and Telegram channels

#### Scenario: No delivery targets
- **WHEN** a job completes with empty deliver_to
- **THEN** the result SHALL only be recorded in history without channel delivery

### Requirement: Execution history
The system SHALL record each job execution in CronJobHistory with status (running/completed/failed), result, error message, tokens used, and timing information.

#### Scenario: Successful execution history
- **WHEN** a cron job executes successfully
- **THEN** a history entry SHALL be created with status "completed" and the agent's response as the result

#### Scenario: Failed execution history
- **WHEN** a cron job execution fails
- **THEN** a history entry SHALL be created with status "failed" and the error message recorded

### Requirement: Concurrency limiting
The system SHALL limit concurrent job executions to the configured maxConcurrentJobs value using a semaphore.

#### Scenario: Max concurrent jobs reached
- **WHEN** maxConcurrentJobs (e.g. 5) jobs are already running and another triggers
- **THEN** the new job SHALL wait for a semaphore slot before executing

### Requirement: Timezone support
The system SHALL support per-job timezone configuration with a default timezone from the global cron config.

#### Scenario: Job with specific timezone
- **WHEN** a job is created with timezone "Asia/Seoul"
- **THEN** the scheduler SHALL interpret the cron expression in the Asia/Seoul timezone

### Requirement: Startup job loading
The system SHALL load all enabled jobs from the database on startup and register them with the cron runner.

#### Scenario: Scheduler startup
- **WHEN** the scheduler starts
- **THEN** all enabled jobs SHALL be loaded from the database and registered with the cron runner

### Requirement: Scheduler history delegation
The `Scheduler` SHALL provide `History(ctx, jobID, limit)` and `AllHistory(ctx, limit)` methods that delegate to the underlying store's `ListHistory` and `ListAllHistory`.

#### Scenario: Query history for a specific job
- **WHEN** `History()` is called with a job ID
- **THEN** the scheduler SHALL return execution history entries for that job from the store

#### Scenario: Query all history
- **WHEN** `AllHistory()` is called
- **THEN** the scheduler SHALL return execution history across all jobs from the store

### Requirement: Delivery start notification
The `Delivery` SHALL provide a `DeliverStart(ctx, jobName, targets)` method that sends a start notification to configured channels before job execution.

#### Scenario: Start notification sent
- **WHEN** a cron job begins execution with configured delivery targets
- **THEN** the executor SHALL call `DeliverStart` before running the agent prompt

#### Scenario: No sender configured
- **WHEN** `DeliverStart` is called with nil sender
- **THEN** the method SHALL log a Warn-level message and return without error

### Requirement: Delivery nil sender log level
The `Delivery.Deliver()` method SHALL log at Warn level (not Debug) when no channel sender is configured.

#### Scenario: Nil sender warning
- **WHEN** `Deliver()` is called with a nil sender
- **THEN** the system SHALL log a Warn-level message instead of Debug
