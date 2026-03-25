## ADDED Requirements

### Requirement: Failed turn cleanup closes dangling tool calls
The runtime SHALL close dangling parent-visible tool calls after a failed turn before the next retry or turn reuses the session history.

#### Scenario: Failed specialist leaves assistant tool call without tool response
- **WHEN** a turn fails after persisting an assistant tool call into parent-visible history and before a matching tool response is recorded
- **THEN** the runtime SHALL append a matching synthetic tool response exactly once for each unanswered tool call
- **AND** later retries SHALL not rely on provider-side orphan repair for that same dangling call

#### Scenario: Cleanup preserves isolation contract
- **WHEN** failed turn cleanup closes a dangling tool call for an isolated specialist
- **THEN** the runtime SHALL not merge raw child-session assistant or tool messages into persisted parent history
- **AND** the persisted parent history SHALL continue to contain only root-authored summaries, discard notes, or cleanup-safe tool response closure
