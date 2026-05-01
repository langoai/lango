## Why

The current `agent.multiAgent=true` path assumes a static, tool-less orchestrator and fixed specialist layout. That creates inconsistent behavior once the runtime needs dynamic teammate creation, least-privilege tool scope per spawned teammate, a stable background identity chain, capability escalation handling, and operator-visible runtime state.

Today the system lacks one coherent production contract for:
- dynamic teammate creation from the existing control-plane tools
- spawn-time narrowing of tool scope without violating role boundaries
- a canonical identity chain across `AgentRun`, background task state, and `RunLedger`
- structured blocked-tool metadata that survives hook interception
- operator-facing visibility into whether the dynamic teammate runtime is active

## What Changes

- This change artifact defines the production contract first; later implementation and archive tasks land the runtime behind that contract
- Define an in-process dynamic teammate runtime for multi-agent execution while keeping the model-facing tool surface limited to `agent_spawn`, `agent_wait`, and `agent_stop`
- Extend the `AgentRun` projection so spawned teammates carry runtime, identity, condition, and capability-escalation state needed by the existing background manager and `RunLedger`
- Require validation of built-in teammate `allowed_tools` against role max scope at spawn time
- Preserve structured blocked-call metadata through hook interception so capability escalation remains policy-driven instead of string-parsed
- Route `DynamicAllowedTools` blocked calls into a capability policy that distinguishes in-scope approval from out-of-scope denial
- Expose dynamic runtime availability through CLI agent inspection without changing v1 remote A2A routing behavior

## User-Facing Impact

- `lango agent status` exposes whether teammate runtime `dynamic-v1` is configured and available for built-in teammates under multi-agent mode; legacy static fallback and remote A2A may still coexist beside that available path
- Cockpit/TUI projection work is explicitly deferred; this change only defines the production contract and CLI-visible status surface
- Remote A2A behavior remains on the existing v1 routing path, preserving current compatibility expectations
