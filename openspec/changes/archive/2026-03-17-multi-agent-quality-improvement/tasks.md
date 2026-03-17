## 1. Orchestrator Prompt Fixes

- [x] 1.1 Remove Diagnostics section from buildOrchestratorInstruction
- [x] 1.2 Change unmatched tools message from "handle directly" to "route to best-matching agent"
- [x] 1.3 Replace "partial answer" guidance with completion-first policy
- [x] 1.4 Add Output Awareness section to orchestrator instruction

## 2. AgentSpec Routing Enhancements

- [x] 2.1 Add ExampleRequests and Disambiguation fields to AgentSpec struct
- [x] 2.2 Add ExampleRequests and Disambiguation fields to routingEntry struct
- [x] 2.3 Update buildRoutingEntry to propagate new fields
- [x] 2.4 Replace ambiguous single-word keywords with compound keywords for all 7 agents
- [x] 2.5 Set ExampleRequests (3-5 examples) for all 7 agents
- [x] 2.6 Set Disambiguation strings for all 7 agents

## 3. Routing Table Rendering

- [x] 3.1 Replace individual tool name listing with tool count in routing table
- [x] 3.2 Render ExampleRequests as bulleted list in routing table
- [x] 3.3 Render Disambiguation as "When NOT this agent" in routing table
- [x] 3.4 Add Disambiguation Rules section to orchestrator instruction
- [x] 3.5 Add Complexity Analysis phase (SIMPLE/COMPOUND/COMPLEX) to Decision Protocol
- [x] 3.6 Strengthen Re-Routing Protocol with failed agent tracking and consecutive failure fallback

## 4. Universal Tool Distribution

- [x] 4.1 Update PartitionTools to collect tool_output_ prefix tools separately
- [x] 4.2 Distribute tool_output_ tools to all non-empty agent sets (excluding planner)
- [x] 4.3 Update PartitionToolsDynamic with same universal distribution logic

## 5. Output Handling in Sub-Agents

- [x] 5.1 Define outputHandlingSection constant
- [x] 5.2 Add Output Handling section to operator, navigator, vault, librarian, automator, chronicler instructions in agentSpecs
- [x] 5.3 Verify planner instruction does NOT include Output Handling

## 6. Dynamic Turn Budget

- [x] 6.1 Add delegation tracking variables (delegationCount, uniqueAgents, plannerInvolved, budgetExpanded)
- [x] 6.2 Implement budget expansion logic (1.5x on planner OR 3+ delegations OR 2+ unique agents)
- [x] 6.3 Replace single wrap-up turn with configurable wrap-up budget (1 default, 3 expanded)
- [x] 6.4 Add expansion logging with session, old/new max, unique agents, delegation count

## 7. Prompt Override File Sync

- [x] 7.1 Replace [REJECT] patterns with transfer_to_agent escalation in 7 IDENTITY.md files
- [x] 7.2 Add Output Handling section to 6 non-planner IDENTITY.md files
- [x] 7.3 Add Output Handling section to 6 non-planner AGENT.md files
- [x] 7.4 Verify planner IDENTITY.md and AGENT.md do NOT have Output Handling

## 8. Tests

- [x] 8.1 Update existing orchestrator tests for new keyword/routing format
- [x] 8.2 Add universal tool_output_ distribution tests (PartitionTools and PartitionToolsDynamic)
- [x] 8.3 Add ExampleRequests and Disambiguation tests
- [x] 8.4 Add disambiguation rules, complexity analysis, output awareness tests
- [x] 8.5 Add dynamic budget expansion condition tests
- [x] 8.6 Add wrap-up budget mechanics tests
- [x] 8.7 Add non-planner Output Handling presence test

## 9. Build Verification

- [x] 9.1 go build ./... passes
- [x] 9.2 go test ./internal/orchestration/... passes
- [x] 9.3 go test ./internal/adk/... passes
- [x] 9.4 No [REJECT] text in any prompt file
- [x] 9.5 tool_output_get distributed to all tool-bearing agents except planner
