## Why

In multi-agent mode, the orchestrator LLM hallucinates agent names (e.g., `browser_agent`, `filesystem_agent`) instead of using actual sub-agent names (`executor`, `researcher`, `planner`, `memory-manager`). This causes "failed to find agent" errors because ADK's `transfer_to_agent` requires exact name matching. The root cause is that the orchestrator has no tools of its own, forcing all tasks through sub-agent delegation. The LLM infers fake agent names from tool name prefixes (`browser_*`, `fs_*`).

## What Changes

- Give the orchestrator all tools directly (via `llmagent.Config.Tools`), in addition to keeping sub-agents
- Update orchestrator instruction to guide direct tool usage for simple tasks and sub-agent delegation for complex tasks
- Simple single-step tasks are handled directly by the orchestrator without delegation round-trips
- Complex multi-step tasks still leverage specialized sub-agents

## Capabilities

### New Capabilities

_(none — this is a behavioral fix to an existing capability)_

### Modified Capabilities

- `multi-agent-orchestration`: The orchestrator now receives all tools directly in addition to sub-agents. Simple tasks are handled without delegation; complex tasks still delegate to sub-agents. The orchestrator instruction explicitly lists valid agent names and prohibits inventing new ones.

## Impact

- `internal/orchestration/orchestrator.go` — orchestrator creation gains `Tools` field + updated instruction
- `internal/orchestration/orchestrator_test.go` — new tests verifying orchestrator has tools and sub-agents simultaneously
- No changes to `internal/app/wiring.go` (already passes full tool set)
- No changes to `internal/orchestration/tools.go` (partition logic unchanged)
