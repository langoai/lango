## MODIFIED Requirements

### Requirement: Task info data model
The TaskInfo struct SHALL include ID, Prompt, Status, Elapsed, Result, Error, OriginChannel, and TokensUsed fields. The bgTaskLister adapter SHALL populate all fields from the background.TaskSnapshot.

#### Scenario: TaskInfo includes result and error
- **WHEN** ListTasks() is called
- **THEN** each TaskInfo includes Result and Error from the snapshot

#### Scenario: TaskInfo includes origin and tokens
- **WHEN** ListTasks() is called
- **THEN** each TaskInfo includes OriginChannel and TokensUsed

## ADDED Requirements

### Requirement: TaskActioner interface
The pages package SHALL define a TaskActioner interface with CancelTask(id) and RetryTask(ctx, id) methods for TUI-initiated background task actions.

#### Scenario: Cancel delegates to manager
- **WHEN** CancelTask is called
- **THEN** it delegates to background.Manager.Cancel

#### Scenario: Retry resubmits with original origin
- **WHEN** RetryTask is called
- **THEN** it fetches the original task snapshot and calls Manager.Submit with the same Prompt, OriginChannel, and OriginSession
