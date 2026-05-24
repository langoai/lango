## MODIFIED Requirements

### Requirement: ExtendableDeadline mechanism

The system SHALL provide an `ExtendableDeadline` in the `internal/deadline` package (extracted from `internal/app`) that wraps a context with a resettable idle timer. Each call to `Extend()` resets the deadline by `idleTimeout` from now, but never beyond `maxTimeout` from creation time. The type SHALL expose a `Reason()` method returning the cause of expiry: `"idle"`, `"max_timeout"`, or `"cancelled"`. Parent context cancellation SHALL be treated as `"cancelled"` when it happens before idle or max-timeout expiry. The derived context SHALL preserve parent context deadline metadata and standard `Err()` semantics.

#### Scenario: Parent cancellation reports cancelled
- **WHEN** the parent context is cancelled before the idle or max timer fires
- **THEN** the derived context SHALL be cancelled and `Reason()` SHALL return `"cancelled"`

#### Scenario: Parent deadline semantics are preserved
- **WHEN** the parent context has an existing deadline
- **THEN** the derived context SHALL expose that deadline via `Deadline()`
- **AND** parent deadline expiry SHALL leave the derived context with `context.DeadlineExceeded`
- **AND** `Reason()` SHALL return `"cancelled"` when the parent deadline fires first
