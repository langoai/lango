## Context

The `internal/agentrt` recovery loop currently retries failures with a single global `maxRetries` counter and no delay between retries. All error classes are treated equally: a malformed JSON tool-call response gets the same retry budget as a transient provider error. This causes unnecessary retries on unrecoverable errors and can amplify rate-limit pressure.

The existing backoff pattern in `internal/economy/escrow/usdc_settler.go:206-228` provides a proven exponential backoff implementation that can be adapted for recovery retries.

## Goals / Non-Goals

**Goals:**
- Add exponential backoff between recovery retries to prevent thundering-herd effects
- Introduce per-error-class retry limits so unrecoverable errors fail fast
- Classify malformed tool-call JSON as a distinct cause class with minimal retry budget
- Extend provider failure tracking in the delegation guard circuit breaker
- Add structured recovery decision events for observability

**Non-Goals:**
- Changing the recovery decision logic (retry vs escalate vs direct-answer)
- Adding jitter to backoff (can be added later)
- Modifying config schema or CLI commands
- Changing wiring or initialization code

## Decisions

### Decision 1: Exponential backoff function
Use `time.Duration(1<<uint(attempt)) * baseDelay` with a 1-second base delay and 30-second max cap. This matches the existing pattern in `usdc_settler.go`. The function is a standalone `ComputeBackoff(attempt int) time.Duration` for testability.

**Alternative**: Fixed delay. Rejected because it does not adapt to repeated failures.

### Decision 2: Per-error-class retry limits as a map
Use `map[CauseClass]int` to define per-class limits. The `Decide()` method checks the class-specific limit before the global `maxRetries` limit. Unknown cause classes fall through to the global limit.

**Alternative**: Configurable per-class limits in `RecoveryCfg`. Rejected for now — the defaults are sufficient and config complexity is not warranted.

### Decision 3: CauseClass type alias
Introduce `type CauseClass string` in `recovery.go` to give semantic meaning to cause class strings in the retry-limit map. Map existing `adk.CauseProviderRateLimit` etc. to recovery-layer constants.

### Decision 4: Provider failure tracking via existing circuit breaker
Extend `DelegationGuard` to accept provider-name keys (not just agent names) by adding a `RecordProviderFailure(provider string, success bool)` method. This reuses the same `circuitBreaker` struct but with a `"provider:"` prefix to avoid key collisions with agent names.

### Decision 5: RecoveryDecisionEvent as a separate event type
Add `RecoveryDecisionEvent` alongside the existing `RecoveryEvent`. The new event carries richer metadata (cause class, backoff duration, attempt count) while the existing `RecoveryEvent` remains for backward compatibility.

## Risks / Trade-offs

- [Risk] `time.Sleep` in `runWithRecovery` blocks the goroutine during backoff → Mitigation: use `select` with `ctx.Done()` so context cancellation interrupts backoff.
- [Risk] Per-class retry limits bypass the global `maxRetries` for classes with higher limits → Mitigation: the effective limit is `min(classLimit, maxRetries)` unless the class limit is explicitly higher, in which case the class limit wins. The global limit serves as a reasonable default, not a hard cap.
- [Trade-off] Provider failure tracking reuses the delegation guard rather than a separate component → Acceptable because the circuit breaker logic is identical; the prefix-based key separation is simple and collision-free.
