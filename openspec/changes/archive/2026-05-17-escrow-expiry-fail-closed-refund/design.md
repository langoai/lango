## Context

`Engine.Expire` already performs refund-before-expired-state for funded and active escrows. `checkExpiry` is called at the beginning of lifecycle methods and should enforce the same safety invariant.

## Design

Introduce a single expiry helper that:

1. Checks whether an entry is eligible to expire.
2. Refunds locked funds when the current status is `funded` or `active`.
3. Updates the entry status to `expired`.
4. Propagates settlement or store errors with context.

`Expire` and `checkExpiry` should both call this helper so explicit and implicit expiry cannot drift again.

## Error Handling

- Refund failure returns an error and does not mark the entry expired.
- Store update failure returns an error instead of being ignored.
- The implicit guard still wraps the final error with `ErrEscrowExpired` so callers can detect expiry.
- Expiry uses reached-time semantics: `now == ExpiresAt` is expired; only `now < ExpiresAt` is early.
- `DanglingDetector` continues to use `maxPending` as a detection threshold, but it does not call `Expire` until ExpiresAt has been reached.

## Non-Goals

- Do not change settlement executor selection.
- Do not add new escrow statuses.
- Do not alter release or dispute semantics.
