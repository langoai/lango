## 1. Orchestrator Tool Assignment

- [x] 1.1 Add `adaptTools(cfg.AdaptTool, cfg.Tools)` call after sub-agent creation in `BuildAgentTree`
- [x] 1.2 Pass adapted tools to orchestrator's `llmagent.Config.Tools` field alongside `SubAgents`

## 2. Orchestrator Instruction Update

- [x] 2.1 Restructure orchestrator instruction with "Direct Tool Usage" and "Sub-Agent Delegation" sections
- [x] 2.2 Add "NEVER invent agent names" rule and explicit agent name list to instruction

## 3. Tests

- [x] 3.1 Add `TestBuildAgentTree_OrchestratorHasToolsAndSubAgents` — verify tools are adapted for both sub-agents and orchestrator
- [x] 3.2 Add `TestBuildAgentTree_OrchestratorAdaptError` — verify orchestrator tool adaptation error is properly propagated
- [x] 3.3 Verify all existing tests still pass with the new orchestrator tool assignment

## 4. Verification

- [x] 4.1 Run `go build ./...` — no compilation errors
- [x] 4.2 Run `go test ./internal/orchestration/...` — all tests pass
