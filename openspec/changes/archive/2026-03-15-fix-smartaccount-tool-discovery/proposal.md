## Why

Lango Agent cannot discover smart account tools (12 tools) when `smartAccount.enabled=false` or `payment.enabled=false`. The agent repeatedly reports only seeing `fs_`, `exec_`, `skill_`, `transfer_to_agent` and tells users to "restart the server" — it has no way to diagnose why tools are missing. Additionally, the `lango account` CLI guard is missing from `blockLangoExec`, and the orchestrator prompt lacks diagnostic guidance.

## What Changes

- Add `lango account` guard to `blockLangoExec` so the agent is redirected to built-in smart account tools instead of attempting CLI subprocess execution.
- Register a disabled `smartaccount` category in the tool catalog when the subsystem is not initialized, so `builtin_list` shows the category with its disabled status and config hint.
- Enhance `initSmartAccount()` log messages with actionable config hints (`smartAccount.enabled`, `payment.enabled`).
- Add `builtin_health` diagnostic tool to the dispatcher, enabling the agent to self-diagnose enabled/disabled categories and required config keys.
- Add a Diagnostics section to the orchestrator prompt instructing it to use `builtin_list` or `builtin_health` when tools appear missing.

## Capabilities

### New Capabilities
- `tool-health-diagnostics`: Agent self-diagnosis tool (`builtin_health`) that reports enabled/disabled categories with config hints

### Modified Capabilities
- `tool-catalog`: Add `builtin_health` to dispatcher, register disabled categories
- `smart-account`: Add `lango account` exec guard, improve init logging with config hints
- `multi-agent-orchestration`: Add diagnostics section to orchestrator prompt

## Impact

- `internal/app/tools.go` — `blockLangoExec` guard list
- `internal/app/app.go` — disabled category registration in `New()`
- `internal/app/wiring_smartaccount.go` — log messages with config hints
- `internal/toolcatalog/dispatcher.go` — new `builtin_health` tool
- `internal/orchestration/tools.go` — orchestrator prompt diagnostics
- Tests: `dispatcher_test.go`, `tools_test.go`, `orchestrator_test.go`
