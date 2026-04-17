# Design: Fix Turn Runner Retry Safety

## Fix 4: Context cancellation breaks retry loop

Add `retryLoop:` label. In `<-parent.Done()` branch, `break retryLoop`. After loop, check `parent.Err()` and unconditionally override result.ErrorCode and result.Outcome to reflect cancellation (prior transient error may still be in result).

## Fix 5: Recovery trace persistence

Add `causeClass`, `attempt`, `backoffMs` to the `marshalTracePayload` map in `recordRecovery()`. This aligns the durable trace with the `RecoveryDecisionEvent` already published to the EventBus.
