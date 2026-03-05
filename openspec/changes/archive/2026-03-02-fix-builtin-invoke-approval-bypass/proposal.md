## Why

The `toolcatalog` + `builtin_invoke` system introduced post-v0.3.0 completely breaks the multi-agent security model. The orchestrator can now invoke **any** tool directly via `builtin_invoke`, bypassing the ADK approval middleware chain. This means dangerous operations (wallet creation, payment execution, secret management) can be executed without approval and without routing through the proper sub-agent (e.g., `vault`).

## What Changes

- **Block dangerous tools in `builtin_invoke` dispatcher**: Tools with `SafetyLevel >= Dangerous` are rejected with an error directing the LLM to delegate to the appropriate sub-agent instead.
- **Remove universal tools from multi-agent orchestrator**: The orchestrator no longer receives `builtin_list`/`builtin_invoke` dispatcher tools. It must delegate all tool-requiring tasks to sub-agents, restoring the v0.3.0 "delegate only" security model.
- **Restore orchestrator prompt to delegation-only**: The orchestrator system prompt no longer mentions direct tool access, consistently instructing delegation to sub-agents.
- **Clean up orchestrator instruction builder**: Remove the `hasUniversalTools` conditional branch from `buildOrchestratorInstruction`, since the orchestrator is always tool-less in multi-agent mode.

## Capabilities

### New Capabilities

### Modified Capabilities
- `tool-catalog`: `builtin_invoke` dispatcher now blocks dangerous tools (safety >= Dangerous) from being proxy-executed
- `multi-agent-orchestration`: Orchestrator no longer receives universal tools; always delegates to sub-agents

## Impact

- `internal/toolcatalog/dispatcher.go` — safety level check added to `builtin_invoke` handler
- `internal/app/wiring.go` — universal tools removed from orchestrator config, prompt simplified
- `internal/orchestration/orchestrator.go` — universal tool adaptation code removed from `BuildAgentTree`
- `internal/orchestration/tools.go` — `buildOrchestratorInstruction` signature simplified (no `hasUniversalTools` param)
- Test files updated: `dispatcher_test.go`, `orchestrator_test.go`
