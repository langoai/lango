## Why

When an LLM provider returns an authentication failure, connection error, or any error that does not match a known heuristic pattern, `classifyError()` falls through to E005 `internal_runtime_error`. The user sees only `"[E005] An internal error occurred. Please try again."` with zero indication of the actual problem. Common provider errors (invalid API key, connection refused) are indistinguishable from genuine internal bugs. The recovery policy also escalates E005 immediately with no retry, even for transient connection errors that could succeed on retry.

## What Changes

- **Add provider auth/connection error classification**: New `CauseProviderAuth` and `CauseProviderConnection` cause classes with case-insensitive pattern matching for common provider error messages (401, 403, unauthorized, invalid api key, connection refused, dial tcp, etc.).
- **Add curated user-facing messages**: `UserMessage()` returns specific guidance for auth and connection errors without exposing raw error details.
- **Improve operator diagnostics**: E005 fallback `OperatorSummary` includes the actual error message (truncated). Recovery log in `coordinating_executor.go` adds `cause_detail` and `error` fields.
- **Fix nil-error defensive path**: `classifyError(nil)` now sets `CauseDetail` to a descriptive string instead of leaving it empty.
- **Update recovery policy**: Auth errors escalate immediately (not retryable). Connection errors classified as transient (retryable).

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `agent-error-handling`: Add provider auth/connection error classification scenarios, curated user messages for new cause classes, and RecoveryPolicy scenarios for auth escalation and connection retry.

## Impact

- **`internal/adk/errors.go`**: New cause constants, `classifyError()` new pattern blocks, `UserMessage()` `ErrModelError` sub-cases, nil-error CauseDetail, E005 fallback OperatorSummary.
- **`internal/agentrt/coordinating_executor.go`**: Recovery log gains `cause_detail` and `error` fields.
- **`internal/agentrt/recovery.go`**: `classifyForRetry()` new cases, `Decide()` `ErrModelError` branch expanded.
- **`internal/adk/errors_test.go`**: New test cases for auth, connection, case-insensitive matching, UserMessage.
- **`internal/agentrt/recovery_test.go`**: New test cases for auth escalation and connection retry.
