## 1. Sub-Agent Escalation Protocol (Layer 1)

- [x] 1.1 Replace `[REJECT]` text instruction with `## Escalation Protocol` section in operator agent spec
- [x] 1.2 Replace `[REJECT]` text instruction with `## Escalation Protocol` section in navigator agent spec
- [x] 1.3 Replace `[REJECT]` text instruction with `## Escalation Protocol` section in vault agent spec
- [x] 1.4 Replace `[REJECT]` text instruction with `## Escalation Protocol` section in librarian agent spec
- [x] 1.5 Replace `[REJECT]` text instruction with `## Escalation Protocol` section in automator agent spec
- [x] 1.6 Replace `[REJECT]` text instruction with `## Escalation Protocol` section in planner agent spec
- [x] 1.7 Replace `[REJECT]` text instruction with `## Escalation Protocol` section in chronicler agent spec

## 2. Orchestrator Prompt Enhancement (Layer 2)

- [x] 2.1 Add Step 0 (ASSESS) to Decision Protocol for direct response to simple conversational requests
- [x] 2.2 Replace "Rejection Handling" section with "Re-Routing Protocol" section
- [x] 2.3 Reorder Delegation Rules to prioritize direct response over delegation

## 3. Code Safety Net (Layer 3)

- [x] 3.1 Add `containsRejectPattern` function with regex-based `[REJECT]` detection in `internal/adk/agent.go`
- [x] 3.2 Add `truncate` helper function for log message preview
- [x] 3.3 Add REJECT detection and re-routing retry logic in `RunAndCollect` after successful `runAndCollectOnce`

## 4. Test Updates

- [x] 4.1 Update `TestAgentSpecs_AllHaveRejectProtocol` → `TestAgentSpecs_AllHaveEscalationProtocol` (check for `transfer_to_agent` + `lango-orchestrator`)
- [x] 4.2 Update `TestBuildAgentTree_RejectProtocolInInstructions` → `TestBuildAgentTree_EscalationProtocolInInstructions`
- [x] 4.3 Update `TestBuildAgentTree_RoutingTableInInstruction` to assert Re-Routing Protocol instead of Rejection Handling
- [x] 4.4 Add `TestBuildOrchestratorInstruction_HasAssessStep` for Step 0
- [x] 4.5 Add `TestBuildOrchestratorInstruction_HasReRoutingProtocol` for re-routing protocol
- [x] 4.6 Add `TestBuildOrchestratorInstruction_DelegationRulesOrder` for delegation rules ordering
- [x] 4.7 Add `TestContainsRejectPattern` table-driven test in `internal/adk/agent_test.go`
- [x] 4.8 Add `TestTruncate` table-driven test in `internal/adk/agent_test.go`

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/orchestration/...` passes
- [x] 5.3 `go test ./internal/adk/...` passes
