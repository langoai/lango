## Why

The ADK agent layer suffers from unbounded tool-calling loops, inefficient session history handling, and uncontrolled memory growth in long-running sessions. These issues degrade response quality, waste tokens, and can cause the agent to spin indefinitely on errors without self-correcting.

## What Changes

- Add turn-limit enforcement to `Agent.Run()` to prevent unbounded tool-calling loops (default 25 turns)
- Implement learning-based self-correction: agent retries with a known fix when a tool error matches a high-confidence learning pattern
- Add token-budget-aware history truncation configurable per model family (Claude 100K, Gemini 200K, GPT-4o 64K)
- Cache truncated history and converted events with `sync.Once` for O(1) repeated access instead of O(n) recomputation
- Add memory token budgeting to `ContextAwareModelAdapter` — reflections prioritized over observations within a configurable budget (default 4000 tokens)
- Auto-trigger meta-reflection in `memory.Buffer` when reflections accumulate past a threshold (default 5) to prevent unbounded growth
- Raise learning engine auto-apply confidence threshold from 0.5 to 0.7 to reduce false-positive fix applications
- Scope learning confidence boosts to exact tool triggers to prevent unrelated learnings from being boosted
- Increase default orchestrator delegation rounds from 5 to 10 with round-budget guidance in the orchestrator prompt
- Fix two test assertion mismatches (`p2p_test.go` command name extraction, `orchestrator_test.go` delegation round string)

## Capabilities

### New Capabilities
- `agent-turn-limit`: Enforcement of maximum tool-calling turns per agent run to prevent infinite loops
- `agent-self-correction`: Learning-based error correction that retries failed operations with known fixes
- `model-aware-token-budget`: Per-model-family token budgeting for session history truncation

### Modified Capabilities
- `learning-engine`: Raise confidence threshold for auto-apply from 0.5 to 0.7; scope success boosts to exact tool triggers
- `observational-memory`: Add token budgeting to memory section assembly; auto-trigger meta-reflection on accumulation
- `multi-agent-orchestration`: Increase default delegation rounds from 5 to 10 with budget guidance prompt

## Impact

- **Core packages**: `internal/adk/` (agent.go, context_model.go, session_service.go, state.go)
- **Memory**: `internal/memory/buffer.go` (auto meta-reflection)
- **Learning**: `internal/learning/engine.go` (confidence threshold, scoped boosts)
- **Orchestration**: `internal/orchestration/orchestrator.go`, `tools.go` (delegation rounds, prompt)
- **Tests**: `internal/cli/p2p/p2p_test.go`, `internal/orchestration/orchestrator_test.go` (assertion fixes)
- **Dependencies**: `go.mod` / `go.sum` updates
