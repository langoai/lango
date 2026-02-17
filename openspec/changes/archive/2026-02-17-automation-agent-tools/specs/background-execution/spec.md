## ADDED Requirements

### Requirement: Manager shutdown
The `Manager` SHALL provide a `Shutdown()` method that cancels all Pending and Running tasks.

#### Scenario: Graceful shutdown
- **WHEN** `Shutdown()` is called
- **THEN** all tasks with Pending or Running status SHALL be cancelled

### Requirement: Start notification
The `Notification` SHALL provide a `NotifyStart(ctx, task)` method that sends a start notification to the task's origin channel.

#### Scenario: Start notification sent
- **WHEN** a background task transitions to Running state
- **THEN** the manager SHALL call `NotifyStart` to notify the origin channel

#### Scenario: No origin channel
- **WHEN** `NotifyStart` is called for a task with empty origin channel
- **THEN** the method SHALL skip notification and return nil
