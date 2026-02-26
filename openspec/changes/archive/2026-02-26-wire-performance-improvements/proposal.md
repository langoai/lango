## Why

Branch `claude/improve-lango-performance-JzdXK` introduced 6 core performance features (agent turn limits, error correction, token budgets, delegation round config, memory token budget, reflection consolidation threshold) but none are wired into the application layer. The builder methods exist but are never called, config fields are missing, and the orchestrator's `MaxDelegationRounds` is hardcoded to `5` despite the default changing to `10`.

## What Changes

- Add 5 new config fields: `agent.maxTurns`, `agent.errorCorrectionEnabled`, `agent.maxDelegationRounds`, `observationalMemory.memoryTokenBudget`, `observationalMemory.reflectionConsolidationThreshold`
- Add `SetReflectionConsolidationThreshold` setter to memory `Buffer`
- Add `AgentOption` functional options pattern to `adk.NewAgent` / `adk.NewAgentFromADK`
- Wire all 6 features in `app/wiring.go` (token budget, max turns, error fix provider, delegation rounds, memory token budget, reflection threshold)
- Update CLI `agent status` to display new performance fields
- Update documentation (multi-agent.md, observational-memory.md, README.md) with new config entries

## Capabilities

### New Capabilities

_None — this change wires existing internal capabilities to the config/CLI/docs layer._

### Modified Capabilities

- `config-types`: Add 5 new config fields to `AgentConfig` and `ObservationalMemoryConfig`
- `agent-turn-limit`: Wire `maxTurns` config to agent constructor via `AgentOption`
- `agent-self-correction`: Wire `errorCorrectionEnabled` config to agent constructor
- `model-aware-token-budget`: Wire token budget to agent via `AgentOption` at construction time
- `multi-agent-orchestration`: Use `maxDelegationRounds` from config instead of hardcoded `5`
- `observational-memory`: Add `memoryTokenBudget` and `reflectionConsolidationThreshold` config wiring
- `cli-agent-inspection`: Add MaxTurns, ErrorCorrection, DelegationRounds to status output

## Impact

- **Config**: `internal/config/types.go` — 5 new fields across 2 structs
- **Memory**: `internal/memory/buffer.go` — 1 new setter method
- **ADK**: `internal/adk/agent.go` — new `AgentOption` type, modified constructor signatures
- **Wiring**: `internal/app/wiring.go` — 4 wiring locations updated
- **CLI**: `internal/cli/agent/status.go` — 3 new output fields
- **Docs**: `docs/features/multi-agent.md`, `docs/features/observational-memory.md`, `README.md`
