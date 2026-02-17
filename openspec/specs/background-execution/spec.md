## Purpose

Define the background task execution system that enables async agent operations with concurrency control, status tracking, and completion notifications.

## Requirements

### Requirement: Background task submission
The system SHALL accept task submissions with a prompt and origin information, returning a unique task ID (UUID).

#### Scenario: Submit a background task
- **WHEN** a task is submitted with prompt "Analyze this project" and origin channel "telegram"
- **THEN** the manager SHALL create a task with status Pending, assign a UUID, and return the ID

### Requirement: Task state machine
Each task SHALL follow a strict state machine: Pending -> Running -> Done/Failed/Cancelled. Status transitions SHALL be protected by a mutex.

#### Scenario: Task completes successfully
- **WHEN** a running task finishes without error
- **THEN** the task status SHALL transition to Done with the result and CompletedAt timestamp

#### Scenario: Task fails
- **WHEN** a running task encounters an error
- **THEN** the task status SHALL transition to Failed with the error message recorded

#### Scenario: Task is cancelled
- **WHEN** Cancel() is called on a running task
- **THEN** the task's context SHALL be cancelled and status SHALL transition to Cancelled

### Requirement: Concurrency limiting
The system SHALL limit concurrent background tasks to the configured maxConcurrentTasks value.

#### Scenario: Max concurrent tasks reached
- **WHEN** maxConcurrentTasks (e.g. 3) tasks are already running and another is submitted
- **THEN** the new task SHALL wait for a semaphore slot before starting execution

### Requirement: Task lifecycle operations
The system SHALL support Cancel, Status, List, and Result operations for managing background tasks.

#### Scenario: List active tasks
- **WHEN** List() is called
- **THEN** the system SHALL return snapshots of all tasks with their current status

#### Scenario: Get task result
- **WHEN** Result() is called for a completed task
- **THEN** the system SHALL return the task's result text

#### Scenario: Get result of incomplete task
- **WHEN** Result() is called for a task with status other than Done
- **THEN** the system SHALL return an error indicating the task is not yet complete

### Requirement: Completion notifications
The system SHALL send completion notifications to the origin channel via the ChannelNotifier interface.

#### Scenario: Task completes with notification
- **WHEN** a background task completes and has an origin channel
- **THEN** the notification system SHALL send a formatted completion message to that channel

### Requirement: Task monitoring
The system SHALL provide monitoring capabilities including active task count and summary information.

#### Scenario: Monitor active tasks
- **WHEN** ActiveCount() is called
- **THEN** the monitor SHALL return the count of tasks in Pending or Running status

### Requirement: In-memory task storage
Background tasks SHALL be stored in-memory only (not persisted). Tasks are ephemeral and lost on server restart.

#### Scenario: Server restart clears tasks
- **WHEN** the server restarts
- **THEN** all previous background tasks SHALL no longer be accessible

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
