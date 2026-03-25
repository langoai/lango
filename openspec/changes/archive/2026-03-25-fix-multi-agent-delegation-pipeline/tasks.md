## Tasks

All tasks are COMPLETE (implementation preceded spec creation).

### Fix A: transfer_to_agent guard exception
- [x] Add `isPureTransferToAgentCall()` helper (`internal/adk/agent.go`)
- [x] Add guard exception at line 331 (`internal/adk/agent.go`)
- [x] Add 5 test cases: nil, pure, mixed, regular, text-only (`internal/adk/agent_test.go`)

### Fix B: convertMessages FunctionResponse split
- [x] Add split logic for role=="tool" && FunctionResponse >= 2 (`internal/adk/model.go`)
- [x] Add `buildToolResponseMessage()` helper (`internal/adk/model.go`)
- [x] Add 3 tests: multi-split, single-unchanged, calls-stay-merged (`internal/adk/state_test.go`)

### Fix C: closeDanglingParentToolCalls OriginAuthor
- [x] Refactor `danglingToolCalls()` to return `danglingCall` with OriginAuthor (`internal/adk/session_service.go`)
- [x] Use OriginAuthor in closure messages with fallback + warning (`internal/adk/session_service.go`)
- [x] Add diagnostic logging: count, origin_authors, call_ids (`internal/adk/session_service.go`)
- [x] Add author preservation test (`internal/adk/session_service_test.go`)
- [x] Add multi-author dangling test (`internal/adk/session_service_test.go`)

### Fix D: Recovery CauseOrchestratorDirectTool → Escalate
- [x] Add CauseOrchestratorDirectTool → RecoveryEscalate in Decide() (`internal/agentrt/recovery.go`)
- [x] Add error classification diagnostic logging (`internal/agentrt/coordinating_executor.go`)
- [x] Add test case (`internal/agentrt/recovery_test.go`)

### Fix E: TUI stdlib logger redirect
- [x] Add `log.SetOutput(logFile)` after logging.Init() (`cmd/lango/main.go`)
- [x] Add `defer logFile.Close()` for cleanup (`cmd/lango/main.go`)

### Fix F: Orphan error message improvement
- [x] Update repairOrphanedToolCalls error content (`internal/provider/openai/openai.go`)
