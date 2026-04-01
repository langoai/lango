## Why

The recovery loop in `internal/agentrt` retries failures uniformly without backoff, has a single global retry limit regardless of error type, and does not recognize malformed tool-call responses as a distinct failure class. This leads to unnecessary retries on unrecoverable errors, thundering-herd retry storms on rate-limited providers, and wasted compute on malformed JSON responses that will never self-correct.

## What Changes

- Add exponential backoff (`ComputeBackoff`) to the recovery retry path so retries do not hammer the provider immediately.
- Introduce per-error-class retry limits (e.g., rate-limit errors get 5 retries, malformed tool calls get 1) alongside the existing global `maxRetries` cap.
- Add `CauseMalformedToolCall` cause class to classify JSON parse errors in tool-call responses.
- Integrate backoff delay into `CoordinatingExecutor.runWithRecovery` before each retry attempt.
- Extend `DelegationGuard` circuit breaker to also track provider-level failures (not just delegation failures).
- Add `RecoveryDecisionEvent` structured event type for recovery decision observability.

## Capabilities

### New Capabilities

### Modified Capabilities
- `agent-control-plane`: Add exponential backoff, per-error-class retry limits, malformed tool-call cause class, provider failure tracking in circuit breaker, and recovery decision event type.

## Impact

- `internal/agentrt/recovery.go` — backoff logic, per-error-class retry limits, malformed JSON error class
- `internal/agentrt/coordinating_executor.go` — integrate backoff delay into `runWithRecovery`
- `internal/agentrt/delegation_guard.go` — extend provider failure tracking
- `internal/agentrt/events.go` — recovery metrics event types
- No CLI, TUI, or external API changes
