## MODIFIED Requirements

### Requirement: Task state machine
Each task SHALL follow a strict state machine: Pending -> Running -> Done/Failed/Cancelled. Status transitions SHALL be protected by a mutex.

The `Fail()` and `Complete()` methods SHALL guard against overwriting the `Cancelled` status. If the task is already in `Cancelled` state when either method is called, the transition SHALL be skipped and the `Cancelled` status SHALL be preserved.

The `execute()` method SHALL check `ctx.Err()` after the runner returns. If the context was cancelled (by user cancellation or timeout), the method SHALL return early without calling `Fail()` or `Complete()`, preserving the cancellation status set by `Cancel()`.

#### Scenario: Task completes successfully
- **WHEN** a running task finishes without error
- **THEN** the task status SHALL transition to Done with the result and CompletedAt timestamp

#### Scenario: Task fails
- **WHEN** a running task encounters an error
- **THEN** the task status SHALL transition to Failed with the error message recorded

#### Scenario: Task is cancelled
- **WHEN** Cancel() is called on a running task
- **THEN** the task's context SHALL be cancelled and status SHALL transition to Cancelled

#### Scenario: Fail does not overwrite Cancelled
- **WHEN** a task is cancelled and the runner subsequently returns an error
- **THEN** `Fail()` SHALL be a no-op and the task status SHALL remain Cancelled

#### Scenario: Complete does not overwrite Cancelled
- **WHEN** a task is cancelled and the runner subsequently returns a result
- **THEN** `Complete()` SHALL be a no-op and the task status SHALL remain Cancelled

#### Scenario: Context cancellation early return
- **WHEN** a task's runner returns and `ctx.Err()` is non-nil
- **THEN** `execute()` SHALL return without calling Fail or Complete
