## 1. Implementation

- [x] Unify explicit and implicit expiry paths behind one helper.
- [x] Refund funded/active escrows before implicit expired-state persistence.
- [x] Propagate expiry store update errors from implicit lifecycle checks.
- [x] Reject explicit `Expire` before ExpiresAt.
- [x] Gate dangling pending cleanup on ExpiresAt instead of forcing early expiry.

## 2. Tests

- [x] Add failing coverage for active implicit expiry refund via `CompleteMilestone`.
- [x] Add failing coverage for implicit expiry store update error propagation.
- [x] Add failing coverage for early explicit `Expire`.
- [x] Add coverage for dangling detector skipping old pending escrows before ExpiresAt.
- [x] Run focused escrow tests.

## 3. OpenSpec

- [x] Update main specs for expiry refund and test coverage invariants.
- [x] Update dangling detector specs and public economy docs.
- [x] Validate and archive the OpenSpec change.
