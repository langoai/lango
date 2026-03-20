## 1. Resume Control Flow Fix (gateway/server.go)

- [x] 1.1 Hoist `confirmResume && resumeRunId != ""` check before `DetectResumeIntent` in `handleChatMessage`
- [x] 1.2 Add explicit `return` after successful resume confirmation broadcast
- [x] 1.3 Replace `context.Background()` with timeout-bounded `s.shutdownCtx`-derived context for `FindCandidates` and `Resume` calls
- [x] 1.4 Replace hardcoded `time.Hour` in `NewResumeManager` with `s.config.RunLedger.StaleTTL` (requires config access in server)

## 2. Session Key Structured Context

- [x] 2.1 Define `RunContext` struct with `SessionType`, `WorkflowID`, `RunID` fields and context key in `internal/session/`
- [x] 2.2 Add `RunContextFromContext(ctx) *RunContext` and `WithRunContext(ctx, rc) context.Context` helpers
- [x] 2.3 Set `RunContext` in context when workflow sessions are created (`internal/workflow/`)
- [x] 2.4 Set `RunContext` in context when background sessions are created (`internal/background/`)
- [x] 2.5 Rewrite `runIDFromSessionContext` in `tool_profile_guard.go` to read `RunContext` instead of colon-split parsing

## 3. Tool Profile Authority Model (tool_profile_guard.go)

- [x] 3.1 Define explicit tool name sets: `orchestratorOnlyRunTools`, `executionRunTools`, `anyRoleRunTools`
- [x] 3.2 Replace `strings.HasPrefix(toolName, "run_") → return true` with role-aware set lookup in `toolAllowedForProfiles`
- [x] 3.3 Replace `strings.HasPrefix(toolName, "exec")` with exact tool names for coding profile (`exec`, `exec_bg`, `exec_status`, `exec_stop`)
- [x] 3.4 Replace `strings.HasPrefix(toolName, "fs_")` with exact filesystem tool names for coding profile
- [x] 3.5 Replace `strings.HasPrefix(toolName, "browser_")` with exact browser tool names for browser profile
- [x] 3.6 Replace all remaining knowledge prefix checks with exact tool names (`search_knowledge`, `search_learnings`, `rag_retrieve`, `graph_traverse`, `graph_query`, `save_knowledge`, `save_learning`, `create_skill`, `list_skills`, `import_skill`, `learning_stats`, `learning_cleanup`, `librarian_pending_inquiries`, `librarian_dismiss_inquiry`)
- [x] 3.7 Verify supervisor profile already uses exact names (`run_read`, `run_active`, `run_note`)

## 4. Empty Agent Name Security (tools.go)

- [x] 4.1 Define `SystemCallerName` constant (e.g., `"system"`) in `internal/runledger/`
- [x] 4.2 Change `checkRole` to reject `agentName == ""` with `ErrAccessDenied` instead of treating as orchestrator
- [x] 4.3 Update internal callers (wiring, tests) to set explicit agent name in context when system-level access is needed

## 5. Run Summary Filtering (adk/context_model.go)

- [x] 5.1 Add status filter in `runSummaryProviderAdapter.ListRunSummaries` to include only `running` and `paused` runs
- [x] 5.2 Verify `assembleRunSummarySection` header ("Active Runs") matches filtered content

## 6. Config Value Wiring

- [x] 6.1 Add `WithTimeout(d time.Duration)` option to `PEVEngine` and apply `context.WithTimeout` around `v.Validate()` in `Verify()`
- [x] 6.2 Wire `config.RunLedger.ValidatorTimeout` to `PEVEngine.WithTimeout()` at app init
- [x] 6.3 Add `PruneOldRuns(ctx context.Context, maxKeep int) error` to `RunLedgerStore` interface
- [x] 6.4 Implement `PruneOldRuns` in `EntStore` (delete oldest completed/failed runs exceeding limit)
- [x] 6.5 Implement `PruneOldRuns` in `MemoryStore` (for test parity)
- [x] 6.6 Call `PruneOldRuns` after run completion/failure events with `config.RunLedger.MaxRunHistory`

## 7. Tests

- [x] 7.1 Add test: confirmed resume executes without `DetectResumeIntent` and returns immediately
- [x] 7.2 Add test: `RunContext` struct round-trips through context correctly
- [x] 7.3 Add test: `runIDFromSessionContext` reads from `RunContext` instead of string parsing
- [x] 7.4 Add test: execution agent denied `run_create` by profile guard (not just `checkRole`)
- [x] 7.5 Add test: `execute_payment` does not match coding profile
- [x] 7.6 Add test: empty agent name returns `ErrAccessDenied` for orchestrator-only tools
- [x] 7.7 Add test: `SystemCallerName` grants orchestrator access
- [x] 7.8 Add test: run summary adapter filters out completed/failed runs
- [x] 7.9 Add test: `PEVEngine.Verify` respects `WithTimeout` deadline
- [x] 7.10 Add test: `PruneOldRuns` removes oldest terminal runs, preserves active runs

## 8. Verification

- [x] 8.1 Run `go build ./...` and confirm no build errors
- [x] 8.2 Run `go test ./internal/runledger/...` and confirm all tests pass
- [x] 8.3 Run `go test ./internal/gateway/...` and confirm all tests pass
- [x] 8.4 Run `go test ./internal/adk/...` and confirm all tests pass
- [x] 8.5 Run `go test ./...` and confirm no regressions

## 9. Downstream Updates

- [x] 9.1 Update CLI help text if tool access behavior descriptions changed
- [x] 9.2 Update README.md RunLedger section to reflect authority model changes
- [x] 9.3 Update any existing documentation referencing `run_*` blanket access
