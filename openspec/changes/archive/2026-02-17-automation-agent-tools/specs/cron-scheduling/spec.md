## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Delivery nil sender log level
The `Delivery.Deliver()` method SHALL log at Warn level (not Debug) when no channel sender is configured.

#### Scenario: Nil sender warning
- **WHEN** `Deliver()` is called with a nil sender
- **THEN** the system SHALL log a Warn-level message instead of Debug
