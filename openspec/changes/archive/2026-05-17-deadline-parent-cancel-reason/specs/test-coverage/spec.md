## ADDED Requirements

### Requirement: ExtendableDeadline parent cancellation coverage

Executable tests SHALL cover parent context cancellation reason classification for the shared deadline package and the app compatibility alias. Tests SHALL also cover preservation of parent deadline metadata and `context.DeadlineExceeded` semantics.

#### Scenario: Shared deadline parent cancellation is covered
- **WHEN** an `internal/deadline` test cancels the parent context before timer expiry
- **THEN** the test SHALL assert the derived context is cancelled and `Reason()` is `"cancelled"`

#### Scenario: App alias parent cancellation is covered
- **WHEN** an `internal/app` test cancels the parent context before timer expiry through `NewExtendableDeadline`
- **THEN** the test SHALL assert the derived context is cancelled and `Reason()` is `"cancelled"`

#### Scenario: Shared deadline parent deadline semantics are covered
- **WHEN** an `internal/deadline` test creates an ExtendableDeadline from a parent context with a deadline
- **THEN** the test SHALL assert the derived context exposes the parent deadline
- **AND** parent deadline expiry leaves the derived context with `context.DeadlineExceeded`
- **AND** `Reason()` is `"cancelled"`
