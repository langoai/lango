# Proposal: Fix Turn Runner Retry Safety

## Problem

Codex review found 2 issues in the turn runner retry loop:

1. **P2 Context cancellation ignored**: The `<-parent.Done()` select branch stops the timer but doesn't break the loop. Next iteration calls `RunStreamingDetailed()` with a cancelled context.
2. **P2 Recovery trace incomplete**: `recordRecovery()` only serializes `action`, `agent`, `error` in the trace event. `CauseClass`, `Attempt`, and `Backoff` are available in `RecoveryInfo` but not persisted, making trace reconstruction impossible.

## Proposed Solution

- Fix 4: Add a labeled loop (`retryLoop:`) and `break retryLoop` on `<-parent.Done()`. Override result with `context_cancelled` error code after the loop.
- Fix 5: Add the 3 missing fields to `recordRecovery`'s `marshalTracePayload` call.
