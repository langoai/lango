## 1. Golden Test Suite (Unit 1 — prerequisite)

- [x] 1.1 Add golden test: user/model/tool/function message round-trip via `eventToMessage()` → store → `EventsAdapter.All()` restoration
- [x] 1.2 Add golden test: streaming partial → final deduplication (verify done event does not duplicate partial text)
- [x] 1.3 Add golden test: orphaned tool response handling (delta with no preceding start → dropped)
- [x] 1.4 Add golden test: delegation event preservation across save/restore cycle
- [x] 1.5 Add golden test: isolated-agent child history segregation (child messages not in parent history)
- [x] 1.6 Add golden test: thought_signature round-trip (preserved through save → restore → provider bridge)
- [x] 1.7 Add regression test: FunctionResponse role correction (`state.go:246-257`) — role="user" stored by ADK → corrected to "tool" on restore
- [x] 1.8 Add regression test: Get() auto-create for non-existent session and auto-renew for expired session
- [x] 1.9 Run `go test ./internal/adk/... -v -count=1` — all tests green

## 2. Session Event Converter Consolidation (Unit 2+3)

- [x] 2.1 Extract inline parts assembly block in `EventsAdapter.All()` (`state.go:282-360+`) into named function `buildEventParts(msg internal.Message, lastAssistantToolCalls []internal.ToolCall) ([]*genai.Part, types.MessageRole)`
- [x] 2.2 Identify shared conversion logic between `eventToMessage()` (`session_service.go:299-370`) and the extracted `buildEventParts()` — document shared vs direction-specific responsibilities
- [x] 2.3 Extract shared FunctionCall field mapping (ID, Name, Args/Input, Thought, ThoughtSignature) into converter helper
- [x] 2.4 Extract shared FunctionResponse field mapping (ID, Name, Response/Output) into converter helper
- [x] 2.5 Update `eventToMessage()` to use shared converter helpers
- [x] 2.6 Update `buildEventParts()` to use shared converter helpers
- [x] 2.7 Verify bug fix #1 (role correction) is preserved — existing + golden tests pass
- [x] 2.8 Verify bug fix #4 (thought_signature) is preserved at `state.go:303-304` and `session_service.go:326-327`
- [x] 2.9 Add contract-deviation comment to `SessionServiceAdapter.Get()` documenting auto-create/renew behavior differs from ADK `session.Service.Get()` contract
- [x] 2.10 Extract child session tracking logic (`session_service.go:194-272`) into named function `trackChildSession()`
- [x] 2.11 Run `go test ./internal/adk/... -v -count=1` — all tests green including golden tests from Unit 1

## 3. Context Model Prompt Pipeline Split (Unit 4)

- [x] 3.1 Create `internal/adk/context_retrieval.go` — move parallel retrieval orchestration (knowledge + RAG + memory + runSummary errgroup) from `context_model.go`
- [x] 3.2 Create `internal/adk/context_assembly.go` — move prompt section combination logic from `context_model.go`
- [x] 3.3 Slim `context_model.go` to GenerateContent entry point + ContextAwareModelAdapter struct definition
- [x] 3.4 Verify retrieval order and token budget calculation unchanged — existing tests pass
- [x] 3.5 Run `go test ./internal/adk/... -v -count=1` — all tests green

## 4. ADK Plugin Integration Spike (Unit 5)

- [x] 4.1 Add `WithPlugins(...*plugin.Plugin)` option to agent creation in `internal/adk/agent.go`
- [x] 4.2 Wire PluginConfig into `runner.Config` at `agent.go:125-129` (NewAgent) and `agent.go:167-171` (NewAgentStreaming)
- [x] 4.3 Create `internal/adk/plugin.go` — spike `OnEventCallback` mapped to event observation logging
- [x] 4.4 Spike `BeforeToolCallback` mapped to `SecurityFilterHook` — verify tool execution can be blocked by returning non-nil
- [x] 4.5 Spike `AfterToolCallback` mapped to `WithLearning` — verify tool name, args, result, error are accessible
- [x] 4.6 Document parity gap table: each toolchain middleware classified as "movable" or "must remain"
- [x] 4.7 Document key gap: ADK callbacks are agent-level, Lango middleware is per-tool
- [x] 4.8 Write decision document: which middlewares to move (if any) and which must stay in toolchain
- [x] 4.9 Run `go test ./internal/adk/... -v -count=1` — all tests green (spike code must not break existing)

## 5. MCPToolset Parity Spike (Unit 6)

- [x] 5.1 Create `internal/adk/mcp_spike_test.go` — test-only evaluation file
- [x] 5.2 Construct ADK `mcptoolset.New()` with test MCP server transport
- [x] 5.3 Evaluate naming contract: can `mcp__{server}__{tool}` naming be preserved? Document pass/fail
- [x] 5.4 Evaluate approval path: can `RequireConfirmationProvider` express always-allow/payment/owner policies? Document pass/fail
- [x] 5.5 Evaluate safety metadata: can per-tool safety level be propagated? Document pass/fail
- [x] 5.6 Evaluate output truncation: can maxOutputTokens be applied? Document pass/fail
- [x] 5.7 Evaluate event publication: can tool events reach event bus via ADK callbacks? Document pass/fail
- [x] 5.8 Produce 5-condition pass/fail summary table with adoption recommendation
- [x] 5.9 Run `go test ./internal/adk/... -v -count=1` — all tests green

## 6. ModelAdapter Streaming Cleanup (Unit 7)

- [x] 6.1 Define state machine states for `toolCallAccumulator`: `Idle`, `Receiving(index/id)`, `Complete`
- [x] 6.2 Refactor `add()` method to use state transitions instead of `hasAny`/`lastIndex` tracking
- [x] 6.3 Map orphaned delta handling (bug fix #2 at `model.go:54-61`) to "delta received in Idle state → drop with warning"
- [x] 6.4 Ensure OpenAI index-based correlation works through state machine
- [x] 6.5 Ensure Anthropic ID/Name-based correlation works through state machine
- [x] 6.6 Verify `done()` output ordering and empty-name filtering preserved
- [x] 6.7 Run `go test ./internal/adk/... -v -count=1` — all tests green including streaming regression tests
