## MODIFIED Requirements

### Requirement: Shared ExtendableDeadline package

The system SHALL provide an `internal/deadline` package containing `ExtendableDeadline` with `New()`, `Extend()`, `Stop()`, and `Reason()` methods. The `internal/app` package SHALL re-export via type alias for backward compatibility. Parent context cancellation SHALL cancel the derived context and classify the deadline reason as `"cancelled"` when it happens before idle or hard-ceiling expiry. The derived context SHALL preserve parent context deadline metadata and standard `Err()` semantics.

#### Scenario: Parent cancellation is classified as cancelled
- **WHEN** an ExtendableDeadline is created from a parent context
- **AND** the parent context is cancelled before idle timeout or max timeout fires
- **THEN** the derived context SHALL be cancelled
- **AND** `Reason()` SHALL return `"cancelled"`

#### Scenario: Parent deadline semantics are preserved
- **WHEN** an ExtendableDeadline is created from a parent context with an existing deadline
- **THEN** the derived context SHALL expose the parent deadline via `Deadline()`
- **AND** parent deadline expiry SHALL leave the derived context with `context.DeadlineExceeded`
- **AND** `Reason()` SHALL return `"cancelled"` when the parent deadline fires first
