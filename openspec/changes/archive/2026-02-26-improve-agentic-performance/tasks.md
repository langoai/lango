## 1. Agent Turn Limit

- [x] 1.1 Add `ErrorFixProvider` interface and `defaultMaxTurns` constant to `internal/adk/agent.go`
- [x] 1.2 Add `maxTurns` and `errorFixProvider` fields to `Agent` struct
- [x] 1.3 Implement `WithMaxTurns(n)` and `WithErrorFixProvider(p)` builder methods
- [x] 1.4 Wrap `runner.Run()` iterator in `Agent.Run()` with turn-counting and limit enforcement
- [x] 1.5 Implement `hasFunctionCalls(event)` helper for detecting tool-calling events

## 2. Agent Self-Correction

- [x] 2.1 Add learning-based retry logic to `RunAndCollect` — retry with correction message when ErrorFixProvider returns a fix
- [x] 2.2 Log learned fix application and retry failure at appropriate levels

## 3. Model-Aware Token Budget

- [x] 3.1 Implement `ModelTokenBudget(modelName)` in `internal/adk/state.go` with per-family budgets
- [x] 3.2 Add `tokenBudget` field to `SessionServiceAdapter` and `WithTokenBudget()` builder
- [x] 3.3 Propagate token budget through `SessionServiceAdapter.Create/Get/getOrCreate` to `SessionAdapter`
- [x] 3.4 Pass token budget from `SessionAdapter.Events()` to `EventsAdapter`

## 4. Event History Caching

- [x] 4.1 Add `sync.Once` lazy caching for `truncatedHistory()` in `EventsAdapter`
- [x] 4.2 Refactor `At(i)` to build and cache full event list on first call instead of iterating per-call

## 5. Memory Token Budgeting

- [x] 5.1 Add `memoryTokenBudget` field and `WithMemoryTokenBudget()` to `ContextAwareModelAdapter`
- [x] 5.2 Implement budget-aware `assembleMemorySection` — reflections first, then observations fill remaining budget
- [x] 5.3 Use `memory.EstimateTokens()` for per-item token counting

## 6. Auto Meta-Reflection

- [x] 6.1 Add `reflectionConsolidationThreshold` field to `memory.Buffer` (default 5)
- [x] 6.2 Add meta-reflection trigger in `Buffer.process()` when reflections >= threshold
- [x] 6.3 Call `ReflectOnReflections` and log result

## 7. Learning Engine Hardening

- [x] 7.1 Raise `autoApplyConfidenceThreshold` from 0.5 to 0.7 in `internal/learning/engine.go`
- [x] 7.2 Update `GetFixForError` and `handleError` to use new threshold constant
- [x] 7.3 Scope `handleSuccess` confidence boost to exact `"tool:<name>"` trigger match

## 8. Orchestration Delegation Rounds

- [x] 8.1 Change default `MaxDelegationRounds` from 5 to 10 in `internal/orchestration/orchestrator.go`
- [x] 8.2 Add round-budget management guidance to orchestrator prompt in `tools.go`
- [x] 8.3 Restructure prompt so delegation rules no longer contain inline round limit

## 9. Test Fixes

- [x] 9.1 Fix `TestNewP2PCmd_Structure` — use `strings.Fields(sub.Use)[0]` for command name extraction
- [x] 9.2 Fix `TestBuildOrchestratorInstruction_ContainsRoutingTable` — update assertion to match actual format string
