## Why

Post-Phase 4 code review uncovered residual runtime bugs and authority contract gaps in RunLedger that were not addressed in the 5 archived phases. The `confirmResume` gating is nested inside `DetectResumeIntent`, causing confirmed resumes to never execute. `run_*` tools are blanket-allowed for all roles, and empty agent names silently grant orchestrator privilege. Two config values (`ValidatorTimeout`, `MaxRunHistory`) are declared but not wired to runtime, and `StaleTTL` is still hardcoded in the gateway.

## What Changes

- **Fix `confirmResume` gating + fall-through** (`gateway/server.go:186-215`): Check `confirmResume && resumeRunId != ""` **before** `DetectResumeIntent`; add `return` after resume success
- **Replace `context.Background()` in resume** (`gateway/server.go:189,207`): Use a timeout-bounded context derived from `s.shutdownCtx` for cancellation support
- **Wire config `StaleTTL`** (`gateway/server.go:187`): Pass `config.RunLedger.StaleTTL` to `NewResumeManager` instead of hardcoded `time.Hour`
- **Fix session key parsing** (`tool_profile_guard.go:53-57`): Use structured run context stored in `internal/session` instead of fragile colon-split positional index
- **BREAKING: Replace `run_*` blanket allow** (`tool_profile_guard.go:62-92`): Execution agents can no longer access `run_apply_policy`, `run_resume`, `run_approve_step`, `run_create`; replaced with per-role explicit allowlists
- **BREAKING: Replace prefix matching** (`tool_profile_guard.go:70`): `strings.HasPrefix(toolName, "exec")` replaced with exact tool name list to prevent matching unrelated tools like `execute_payment`
- **Secure empty agent name** (`tools.go:602`): `agentName == ""` no longer silently grants orchestrator; requires explicit system caller identity
- **Filter run summary injection** (`adk/context_model.go:313-340`): Only active/paused runs injected into LLM context, not completed/failed/stale
- **Wire `ValidatorTimeout`**: Apply as context deadline during `PEVEngine.Verify`
- **Wire `MaxRunHistory`**: Add store-level GC/pruning beyond CLI-only usage

## Capabilities

### New Capabilities

_(none — all fixes are within existing capability boundaries)_

### Modified Capabilities

- `run-ledger`: Resume control flow, tool profile authority model, empty-agent identity contract, config value wiring, run summary filtering

## Impact

- **Code**: `internal/gateway/server.go`, `internal/runledger/tool_profile_guard.go`, `internal/runledger/tools.go`, `internal/runledger/pev.go`, `internal/adk/context_model.go`, `internal/app/wiring_knowledge.go`, `internal/session/context.go`
- **Breaking**: Execution agents lose access to orchestrator-only `run_*` tools; tools matching `exec` prefix are now exact-match only
- **Downstream**: CLI help text, docs, README must reflect authority changes
- **Risk**: Medium — authority changes affect tool access control; requires comprehensive test coverage
