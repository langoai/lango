## Context

The TUI chat model (`internal/cli/chat/chat.go`) blocks input during streaming via `input.go:63` `Blur()` and `chat.go:389` `inputAcceptsText()`. The turn runner (`internal/turnrunner/runner.go:192`) calls `executor.RunStreamingDetailed()` once per turn with no retry. `ContextAwareModelAdapter` (`internal/adk/context_model.go:169`) assembles the system prompt but has no compactor — `CompactMessages` is wired only to the observational memory buffer (`app.go:704`). Error classification exists (`errors.go:136`) with 16 cause classes but no recovery action mapping.

## Goals / Non-Goals

**Goals:**
- Users can interrupt streaming and redirect the agent with a single keypress (Enter)
- Transient provider errors (rate limit, connection, transient) auto-recover without user intervention (max 3 attempts)
- Stale streams (no chunk for 30s) are detected and retried automatically
- Context overflow is handled by inline emergency compaction before the LLM call fails

**Non-Goals:**
- Background hygiene compaction (Phase 3 scope)
- Model fallback chain (switching to a different model on failure)
- New turntrace outcome for `canceled` (reuses `OutcomeTimeout`)
- Modifying the existing `budgets.Degraded` semantics (it remains "config issue, not session issue")

## Decisions

### D1: Redirect via `pendingRedirectInput` queue, NOT debounce
**Decision**: Store redirect input in a field, consume it in the `DoneMsg` handler after the cancelled turn completes.
**Alternative**: Time-based debounce (50ms) after `cancelFn()`.
**Rationale**: bubbletea is a message-based event loop. The DoneMsg/ErrorMsg arrival naturally signals that the previous turn has cleaned up. A debounce introduces a race condition between the timer and message delivery. The queue pattern guarantees ordering.

### D2: Redirect consumes in DoneMsg, not ErrorMsg
**Decision**: `pendingRedirectInput` is consumed in the `DoneMsg` handler, short-circuiting before the `stateFailed` + error status path (`chat.go:212`).
**Alternative**: Consume in `ErrorMsg` handler (`chat.go:236`).
**Rationale**: `submitCmd()` returns `DoneMsg` for all outcomes except raw `turnRunner.Run()` errors. `classifyResult()` wraps all `AgentError`s into `Result` and returns `nil` error. Cancel/timeout are always `DoneMsg`.

### D3: Redirect judgment uses local state only, not Outcome/CauseClass
**Decision**: Check `m.pendingRedirectInput != ""` — no Outcome or CauseClass inspection.
**Alternative**: Check `Result.Outcome == OutcomeTimeout && CauseClass == CauseTimeoutCanceled`.
**Rationale**: `context.Canceled` and `context.DeadlineExceeded` both map to `CauseTimeoutHard` (`errors.go:147`). There is no separate `CauseTimeoutCanceled`. The presence of a queued redirect IS the proof of user-initiated cancel.

### D4: Runner is single retry owner, not Agent
**Decision**: `Runner.Run()` wraps `executor.RunStreamingDetailed()` in a retry loop, creating a new context per attempt.
**Alternative**: Retry inside `Agent.RunStreamingDetailed()` (`agent.go:879`).
**Rationale**: Stale detection cancels the attempt context from the Runner level. If retry lived inside the Agent, it would operate on a dead context. Single owner at the Runner level ensures fresh context per attempt and avoids split retry logic.

### D5: RecoveryAction is Retry or Abort only — no CompressAndRetry
**Decision**: Runner decides `Retry` or `AbortWithHint`. Context overflow compaction is entirely `ContextAwareModelAdapter`'s responsibility.
**Alternative**: Runner decides `CompressAndRetry` and invokes compaction.
**Rationale**: Compaction requires access to session store, message indices, and summary generation — all of which live in the context model layer. Having the Runner own compaction would create a tight coupling between the turn orchestration layer and the session/prompt layer. The context model already runs inline during `GenerateContent()` and can detect overflow before the LLM call is made.

### D6: SessionCompactor injected via WithSessionCompactor pattern
**Decision**: Add `SessionCompactor` interface to `internal/adk/` and inject via `WithSessionCompactor()` on `ContextAwareModelAdapter`.
**Alternative**: Pass compactor via context or global registration.
**Rationale**: Follows the existing `WithMemory()`, `WithRuntimeAdapter()`, `WithBudgetManager()` pattern. Compile-time safe, testable, no global state.

### D7: budgets.Degraded is NOT a compaction trigger
**Decision**: `Degraded` means "base prompt alone exceeds available budget" — a configuration problem that compaction cannot fix. Emergency compaction triggers only on `measured > modelWindow × 0.9`.
**Alternative**: Use Degraded as a compaction signal.
**Rationale**: When Degraded is true, available budget is 0 or negative because the base prompt (system prompt + static sections) is too large. Compacting session messages won't help — the problem is the base prompt size, not session history.

### D8: Retry events accumulate in one turn trace
**Decision**: All retry attempts and recovery events accumulate within a single `traceRecorder` instance (created once per turn at `runner.go:175`).
**Alternative**: Create a new trace per attempt.
**Rationale**: From the operator's perspective, retries are part of the same logical turn. Creating separate traces would fragment the diagnostic story. The existing `recordRecovery()` method already supports recording recovery events within a trace.

## Risks / Trade-offs

- **[Stale false positive during tool execution]** → Mitigation: Activate stale timer only after first chunk arrives. Tool execution happens before streaming begins, so the timer is inactive during tool runs.
- **[Retry amplifies cost on persistent errors]** → Mitigation: Max 3 attempts with jittered exponential backoff. `CauseProviderAuth` maps to `AbortWithHint` (no retry).
- **[Emergency compaction delays the current turn]** → Mitigation: Compaction is synchronous but lightweight (single DB operation). Summary generation is the expensive part — use a concise prompt. Phase 3 adds background hygiene compaction to prevent emergency triggers.
- **[DoneMsg short-circuit may miss edge cases]** → Mitigation: The short-circuit checks `pendingRedirectInput != ""` before any other DoneMsg processing. If the field is empty, the entire existing path runs unmodified.
