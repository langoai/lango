## 1. Backoff and Cause Classes (recovery.go)

- [x] 1.1 Add `CauseClass` type alias and cause class constants (`CauseRateLimit`, `CauseTransient`, `CauseMalformedToolCall`, `CauseTimeout`)
- [x] 1.2 Add `ComputeBackoff(attempt int) time.Duration` function with exponential backoff capped at 30s
- [x] 1.3 Add `defaultRetryLimits` map and `classifyForRetry` helper to map `AgentError.CauseClass` to recovery `CauseClass`
- [x] 1.4 Add per-class retry count tracking to `RecoveryContext` and extend `Decide()` to check per-class limits
- [x] 1.5 Add tests for `ComputeBackoff`, per-class retry limits, and malformed tool-call classification

## 2. Executor Backoff Integration (coordinating_executor.go)

- [x] 2.1 Add context-aware backoff sleep before retry in `runWithRecovery` (both `RecoveryRetry` and `RecoveryRetryWithHint` paths)
- [x] 2.2 Publish `RecoveryDecisionEvent` on event bus when recovery decision is made

## 3. Provider Failure Tracking (delegation_guard.go)

- [x] 3.1 Add `RecordProviderFailure(provider string, success bool)` method with `"provider:"` key prefix
- [x] 3.2 Add `IsProviderOpen(provider string) bool` method
- [x] 3.3 Add tests for provider failure tracking independence from agent circuits

## 4. Recovery Events (events.go)

- [x] 4.1 Add `RecoveryDecisionEvent` struct with `CauseClass`, `Action`, `Attempt`, `Backoff`, `SessionKey` fields
- [x] 4.2 Add `EventName()` method returning `"agent.recovery.decision"`

## 5. Verification

- [x] 5.1 Run `go build ./internal/agentrt/...` and fix any compilation errors
- [x] 5.2 Run `go test ./internal/agentrt/... -v` and fix any test failures
