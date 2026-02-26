## 1. Config Fields

- [x] 1.1 Add `MaxTurns`, `ErrorCorrectionEnabled`, `MaxDelegationRounds` to `AgentConfig` in `internal/config/types.go`
- [x] 1.2 Add `MemoryTokenBudget`, `ReflectionConsolidationThreshold` to `ObservationalMemoryConfig` in `internal/config/types.go`

## 2. Memory Buffer

- [x] 2.1 Add `SetReflectionConsolidationThreshold(n int)` method to `Buffer` in `internal/memory/buffer.go`

## 3. Agent Constructor Options

- [x] 3.1 Add `AgentOption` type and option constructors (`WithAgentTokenBudget`, `WithAgentMaxTurns`, `WithAgentErrorFixProvider`) in `internal/adk/agent.go`
- [x] 3.2 Add `opts ...AgentOption` parameter to `NewAgent` and wire token budget, max turns, error fix provider
- [x] 3.3 Add `opts ...AgentOption` parameter to `NewAgentFromADK` and wire token budget, max turns, error fix provider

## 4. Application Wiring

- [x] 4.1 Create `buildAgentOptions(cfg, kc)` helper in `internal/app/wiring.go`
- [x] 4.2 Wire `buildAgentOptions` into single-agent path (`NewAgent` call)
- [x] 4.3 Wire `buildAgentOptions` into multi-agent path (`NewAgentFromADK` call)
- [x] 4.4 Change `MaxDelegationRounds: 5` to `cfg.Agent.MaxDelegationRounds` in orchestrator config
- [x] 4.5 Wire `MemoryTokenBudget` to `ctxAdapter.WithMemoryTokenBudget()` in both context-aware adapter paths
- [x] 4.6 Wire `ReflectionConsolidationThreshold` to `buffer.SetReflectionConsolidationThreshold()` after buffer creation

## 5. CLI

- [x] 5.1 Add `MaxTurns`, `ErrorCorrectionEnabled`, `MaxDelegationRounds` to `statusOutput` struct and table output in `internal/cli/agent/status.go`

## 6. Documentation

- [x] 6.1 Update `docs/features/multi-agent.md`: change default 5→10, add `maxDelegationRounds` config entry
- [x] 6.2 Update `docs/features/observational-memory.md`: add 2 config fields and auto-consolidation section
- [x] 6.3 Update `README.md`: add 5 new config entries to the config table

## 7. Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./internal/config/ ./internal/adk/ ./internal/memory/ ./internal/cli/agent/` passes
