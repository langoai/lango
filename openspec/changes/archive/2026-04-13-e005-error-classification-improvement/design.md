## Overview

Improve error classification so that common provider errors (auth failures, connection errors) receive specific error codes and curated user-facing messages instead of the opaque E005 fallback. Keep raw error details in operator/log paths only.

## Design Decisions

### D1: Case-insensitive matching with narrowed auth patterns

**Decision**: Use `strings.ToLower(msg)` for new pattern blocks only. Auth patterns limited to: `"401"`, `"403"`, `"unauthorized"`, `"invalid api key"`, `"invalid_api_key"`, `"authentication failed"`. Excludes broad `"authentication"` to avoid false positives.

**Rationale**: Provider error messages vary in casing (e.g., `"Unauthorized"` vs `"unauthorized"`). Narrowed patterns reduce false-positive risk while covering the common cases from Anthropic, OpenAI, Ollama, and custom providers.

### D2: Curated user messages only — no raw detail exposure

**Decision**: `UserMessage()` returns curated strings for auth (`"Check your API key configuration"`) and connection (`"Check your network and provider URL"`). The E005 `default` case remains unchanged — no `CauseDetail` in user-facing output.

**Rationale**: `UserMessage()` output flows through `error_format.go:16` directly to users and through `turnrunner/runner.go:339-340` into turn traces. Raw details could contain noisy internal errors or sensitive provider responses.

### D3: Connection errors are retryable; auth errors are not

**Decision**: `CauseProviderConnection` → `CauseTransient` in `classifyForRetry()` → `RecoveryRetry`. `CauseProviderAuth` → `CauseUnknown` → `RecoveryEscalate`.

**Rationale**: Connection failures (refused, DNS, reset) are often transient. Auth failures (wrong API key) are permanent until configuration changes.

### D4: Inline truncation for OperatorSummary

**Decision**: Use `msg[:min(len(msg), 200)]` inline instead of a separate helper function.

**Rationale**: Single use site. Go 1.25.4 supports `min()` built-in. No need for a helper.

## Implementation Steps

### Part A: New cause constants + classifyError patterns

**File**: `internal/adk/errors.go`

1. Add `CauseProviderAuth` and `CauseProviderConnection` constants after line 43.
2. In `classifyError()`, between the 500/503 check (line 253) and the `"tool"` catch-all (line 255), add two blocks:
   - Auth: lowercase match on `401`, `403`, `unauthorized`, `invalid api key`, `invalid_api_key`, `authentication failed` → `ErrModelError` + `CauseProviderAuth`
   - Connection: lowercase match on `connection refused`, `no such host`, `dial tcp`, `connection reset` → `ErrModelError` + `CauseProviderConnection`
3. Update nil-error path (line 128-133): add `CauseDetail: "classifyError called with nil error (defensive)"`.
4. Update E005 fallback (line 264-269): `OperatorSummary` includes truncated `msg`.

### Part B: UserMessage() curated messages

**File**: `internal/adk/errors.go`

Replace `ErrModelError` case (line 101-102) with a switch on `CauseClass`:
- `CauseProviderAuth` → auth guidance
- `CauseProviderConnection` → network guidance
- `default` → existing generic message

### Part C: Coordinating executor log improvement

**File**: `internal/agentrt/coordinating_executor.go`

Add `"cause_detail"` and `"error"` fields to the AgentError branch of the recovery log (lines 166-172).

### Part D: Recovery policy update

**File**: `internal/agentrt/recovery.go`

1. `classifyForRetry()`: add `CauseProviderAuth → CauseUnknown` and `CauseProviderConnection → CauseTransient`.
2. `Decide()` `ErrModelError` case: add explicit auth escalation and connection retry branches.

### Part E: Tests

**File**: `internal/adk/errors_test.go` — classifyError + UserMessage test cases
**File**: `internal/agentrt/recovery_test.go` — recovery policy test cases

## Risks

| Risk | Mitigation |
|------|-----------|
| Lowercase normalization could interact with existing numeric patterns (500, 503) | New blocks placed after existing 500/503 check; numbers are case-insensitive anyway |
| `"dial tcp"` could match non-provider errors | Only provider-wrapped errors reach `classifyError`; the pattern is specific enough |
| Auth/connection patterns may not cover all providers | Unmatched errors still fall through to E005 with the improved OperatorSummary showing the actual message |
