## Why

`ExtendableDeadline` documents `ReasonCancelled` as covering explicit stop or parent cancellation, but parent context cancellation currently leaves `Reason()` at the default idle reason. This can misclassify shutdown-driven cancellations as idle timeouts in higher-level timeout reporting.

## What Changes

- Update `internal/deadline.ExtendableDeadline` to observe parent context cancellation.
- Set `ReasonCancelled` when the parent is cancelled before idle or hard-ceiling timers fire.
- Add regression tests for both `internal/deadline` and the backward-compatible `internal/app` alias path.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `adaptive-idle-timeout`: Clarify and enforce parent cancellation reason semantics.
- `auto-extend-timeout`: Clarify and enforce parent cancellation reason semantics for the extracted deadline mechanism.
- `test-coverage`: Add regression coverage for parent cancellation reason classification.

## Impact

- Affected code: `internal/deadline`, plus tests through the `internal/app` compatibility alias.
- No public CLI/API syntax changes.
- Runtime effect: shutdown or caller cancellation is classified as `"cancelled"` instead of `"idle"` when it wins the race.
