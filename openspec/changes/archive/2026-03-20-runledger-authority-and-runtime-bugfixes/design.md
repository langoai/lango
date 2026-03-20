## Context

Post-Phase 4 code review of RunLedger revealed residual defects spanning the gateway resume flow, tool profile authority model, agent identity, run summary injection, and unwired config values. These bugs were introduced across phases 1-4 and survived because they involve cross-cutting interactions between gateway, tool guard middleware, and ADK context assembly. The fixes stay within existing runtime boundaries, but the structured run context must live in a shared package to avoid import cycles between `runledger` and `workflow/background`.

Key source locations:
- `internal/gateway/server.go:186-215` — resume control flow
- `internal/runledger/tool_profile_guard.go` — tool governance middleware
- `internal/runledger/tools.go:602` — agent identity check
- `internal/adk/context_model.go:313-340` — run summary injection
- `internal/runledger/pev.go:46` — PEV engine verify
- `internal/config/types_runledger.go` — config declarations
- `internal/session/context.go` — shared context transport for session metadata

## Goals / Non-Goals

**Goals:**
- Fix `confirmResume` gating so confirmed resumes actually execute
- Use a timeout-bounded context derived from `s.shutdownCtx` for resume operations instead of `context.Background()`
- Wire `config.RunLedger.StaleTTL` to `NewResumeManager` instead of hardcoded `time.Hour`
- Replace fragile colon-split session key parsing with structured context values stored in a shared package
- Replace `run_*` blanket allow with per-role explicit allowlists (BREAKING)
- Replace prefix matching (`strings.HasPrefix(toolName, "exec")`) with exact tool name lists (BREAKING)
- Reject `agentName == ""` as orchestrator identity; require explicit system caller
- Filter run summary injection to only active/paused runs
- Wire `ValidatorTimeout` as context deadline in `PEVEngine.Verify`
- Wire `MaxRunHistory` to store-level pruning

**Non-Goals:**
- Redesigning the resume protocol (fix the control flow, not the interaction pattern)
- Adding new tool profiles or categories
- Changing the PEV engine interface or validator contract
- Adding new config fields (only wiring existing declared-but-unused ones that have a real runtime consumer in this change)

## Decisions

### Decision 1: Hoist `confirmResume` check out of `DetectResumeIntent`

**Choice**: Move the `confirmResume && resumeRunId != ""` check to run **before** `DetectResumeIntent`, then add an explicit `return` after successful resume confirmation.

**Rationale**: The current code nests the confirmation check inside the `DetectResumeIntent` branch as an `else if`. This means:
1. If the user sends `confirmResume: true`, the system calls `DetectResumeIntent` on the confirm message text — which may not match, causing the entire block to be skipped.
2. Even when the confirmation path executes, there is no `return`, so the message continues to the normal agent handler, causing a duplicate invocation.

Hoisting the check and adding `return` fixes both problems with minimal code change.

### Decision 2: Use shutdown-derived timeout-bounded context for resume operations

**Choice**: Replace `context.Background()` at lines 189 and 207 with a context derived from `s.shutdownCtx`, bounded by the request timeout / hard ceiling configuration.

**Rationale**: `context.Background()` ignores server shutdown signals, meaning a resume operation could outlive the server's graceful shutdown window. `s.shutdownCtx` is already used as the parent for normal agent operations in the same function. Wrapping it with a timeout keeps resume operations bounded even though `handleChatMessage` does not receive an HTTP request context directly.

### Decision 3: Replace `run_*` blanket allow with role-based allowlists

**Choice**: Define three explicit tool sets:
- `orchestratorOnlyTools`: `run_create`, `run_apply_policy`, `run_approve_step`, `run_resume`
- `executionOnlyTools`: `run_propose_step_result`
- `anyRoleTools`: `run_read`, `run_active`, `run_note`

In `toolAllowedForProfiles`, replace `strings.HasPrefix(toolName, "run_") → return true` with a lookup against the appropriate set based on the caller's role context.

**Rationale**: The blanket allow defeats the purpose of tool profiles. An execution agent with a `coding` profile can currently call `run_create` or `run_apply_policy`, which are orchestrator-only operations. The existing `checkRole` function in tools.go already enforces per-tool role checks, but the profile guard short-circuits before `checkRole` ever runs.

**Breaking change**: Execution agents that previously relied on `run_*` blanket access will lose access to orchestrator-only tools. The existing role check in `checkRole` would have rejected these calls anyway, so in practice no correctly-behaving agent is affected.

### Decision 4: Replace prefix matching with exact tool name sets

**Choice**: Replace prefix checks with explicit tool sets based on the currently registered tool names. For example:
- coding: `exec`, `exec_bg`, `exec_status`, `exec_stop`, `fs_read`, `fs_list`, `fs_write`, `fs_edit`, `fs_mkdir`, `fs_delete`, `fs_stat`
- browser: `browser_navigate`, `browser_action`, `browser_screenshot`
- knowledge: `search_knowledge`, `search_learnings`, `rag_retrieve`, `graph_traverse`, `graph_query`, `save_knowledge`, `save_learning`, `create_skill`, `list_skills`, `import_skill`, `learning_stats`, `learning_cleanup`, `librarian_pending_inquiries`, `librarian_dismiss_inquiry`

**Rationale**: Prefix matching is fragile — a tool named `execute_payment` or `execution_log` would incorrectly match the `exec` prefix. Since the runtime tool names are already explicit and stable, explicit sets provide the strongest guarantee and prevent accidental access.

### Decision 5: Reject empty agent name as orchestrator

**Choice**: In `checkRole()`, change `agentName == ""` from granting orchestrator access to returning `ErrAccessDenied` with a message indicating a system caller identity is required.

**Rationale**: An empty agent name is an ambiguous identity — it could be an uninitialized sub-agent, a test harness, or a misconfigured component. Silently treating it as the most privileged role violates the principle of least privilege. Legitimate system callers should set an explicit identity.

**Backward compatibility**: Add a package-level constant `SystemCallerName = "system"` that internal callers (like wiring code or tests) can set in context to indicate legitimate system-level access.

### Decision 6: Filter run summary injection by status

**Choice**: In `assembleRunSummarySection`, filter the summaries returned by `ListRunSummaries` to only include runs with status `running` or `paused`. Completed, failed, and stale runs are excluded.

**Rationale**: Injecting completed/failed runs into LLM context wastes token budget and can confuse the orchestrator into attempting to manage runs that are already terminal. The function is named "Active Runs" in its output header, so it should only include active runs.

**Implementation**: The filter is applied at the adapter layer (`runSummaryProviderAdapter.ListRunSummaries`) rather than at the store query, because the store may need unfiltered results for CLI and other consumers.

### Decision 7: Wire only config values with a real runtime consumer in this change

**Choice**:
- `ValidatorTimeout`: Add a `WithTimeout(d time.Duration)` option to `PEVEngine`, applied as `context.WithTimeout` around the `v.Validate()` call in `Verify()`.
- `MaxRunHistory`: Add a `PruneOldRuns(ctx context.Context, maxKeep int) error` method to `RunLedgerStore`, called after run completion events.

**Rationale**: These config values are already declared in `RunLedgerConfig` with sensible defaults and have clear runtime consumers in the current code. `PlannerMaxRetries` remains out of scope for this change because there is no planner retry loop in the live runtime to wire without reopening orchestration design.

### Decision 8: Structured run context lives in `internal/session`

**Choice**: Replace the colon-split parsing in `runIDFromSessionContext` with a typed context value stored in `internal/session`, alongside the existing session key string.

**Rationale**: `runledger` already imports `workflow` and `background` via write-through adapters. Putting the new context type in `runledger` would force `workflow/background -> runledger` imports and create an import cycle. `internal/session` is already shared by gateway, ADK, and tool middleware, so it is the correct neutral home.

**Implementation**: Add a `RunContext` struct with `SessionType`, `WorkflowID`, and `RunID` fields to `internal/session/context.go`. Set it in context when workflow/background sessions are created. The guard reads it directly instead of parsing the session key string.

## Risks / Trade-offs

- **[Risk]** Breaking change in tool profile allowlists may affect custom agents that call `run_*` tools from execution context. **Mitigation**: The `checkRole` function already rejects these calls — the profile guard was incorrectly bypassing `checkRole`. Any agent that was "working" was only working by accident and would fail at the role check layer.
- **[Risk]** Rejecting empty agent name may break test fixtures or internal wiring that does not set agent name. **Mitigation**: Introduce `SystemCallerName` constant and update all internal callers. Add test helpers that set the agent name in context.
- **[Trade-off]** Filtering run summaries at the adapter layer means the store fetches more rows than needed. **Accepted**: The limit is already 3, so at most 3 rows are fetched. The overhead of an unfiltered query on a tiny result set is negligible compared to the clarity of keeping the store interface general-purpose.
- **[Trade-off]** Using exact tool name sets requires updating the set when new tools are added. **Accepted**: Tool additions are rare and already require updates to tool registration, help text, and documentation. Adding a name to an allowlist set is minimal additional work and prevents accidental access.
