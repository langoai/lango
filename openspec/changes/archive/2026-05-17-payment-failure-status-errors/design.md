## Overview

`failTx` is the central helper for transitioning a pending payment transaction to `failed`. It should not silently discard persistence errors because that leaves payment history and operational state inconsistent with the returned `Send` error.

## Decisions

### Return Joined Failure Information

Change `failTx` to return an error. If the failed-status update succeeds, it returns `nil`. If the update fails, it returns an error that preserves both the original transaction failure and the persistence failure, using Go error wrapping/joining semantics.

### Callers Keep Stage Context

`Send` call sites keep their existing stage labels such as `build transaction`, `sign transaction`, `submit transaction`, and `confirm transaction`. When `failTx` returns an error, the stage wrapper should wrap the combined error so callers still know which payment stage failed.

### Keep Confirmed Spending Behavior Unchanged

The existing best-effort `limiter.Record` after confirmed on-chain success remains non-fatal because the transaction has already succeeded on-chain. This change is limited to failed transaction status persistence.

## Risks

- Error strings become more detailed for rare persistence failures. Existing tests assert substrings rather than exact full messages for these paths.
