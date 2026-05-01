## Context

This change defines the production contract for a dynamic in-process teammate runtime without widening the model-facing tool surface. The runtime must coexist with existing multi-agent orchestration, the background manager, and `RunLedger`, while preserving current remote A2A behavior in v1.

## Source Of Truth

`AgentRun.ID`, the background task ID, and the `RunLedger` run ID remain the canonical identity chain for a teammate run. The system does not mint a second runtime-local identifier for spawned teammates.

`AgentRun` stores the control-plane projection that operators and tools read back through `agent_wait` and status surfaces. The background manager owns in-process execution and lifecycle transitions. `ChildSession` owns execution isolation and parent-child session linkage. `RunLedger` mirrors durable state for inspection and recovery, rather than becoming an alternate source of truth.

This lets the runtime keep one stable identity across spawn, execution, waiting, approval pauses, cancellation, and completion.

## V1 Tool Surface

The model-facing tool surface remains exactly `agent_spawn`, `agent_wait`, and `agent_stop`. v1 does not introduce `teammate_*` tools or `agent_message`.

The runtime grows behind those existing tools by enriching spawn projection, background execution, and capability policy. This keeps prompt surface area small and avoids creating parallel lifecycle APIs before the in-process runtime is proven.

## Teammate Types

Existing specialist roles become teammate types for built-in dynamic spawning. The canonical built-in teammate registry is `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. A built-in teammate type maps to a role-defined maximum tool scope, execution prompt shape, and runtime defaults, and spawn-time validation, prompt defaults, and role max scope MUST derive from that built-in registry.

Spawn-time `allowed_tools` may narrow that role max scope but may never expand it. This gives callers a least-privilege way to create focused teammates without creating ad hoc roles or bypassing orchestration policy.

## Capability Escalation

Capability escalation starts when a tool call is blocked by `DynamicAllowedTools`. Runtime interception must preserve the structured blocked-call metadata emitted by the hook layer so policy can evaluate the attempted tool, the active allowlist, and the teammate's role max scope without reparsing human text.

Approval can only be requested for tools that are inside the teammate role's max scope. Approval never overrides the role ceiling. When an in-scope tool is approved, the runtime extends the run's `AllowedTools` projection so the next execution context includes the granted tool. Out-of-scope tools are denied directly and remain denied even if an operator would otherwise approve them.

## Compatibility

`transfer_to_agent` remains a legacy ADK static fallback, not the primary production teammate path. v1 keeps it available for:
- legacy static sub-agent fallback
- specialist re-routing when a legacy flow is already in progress
- existing remote A2A paths

This change does not redefine remote A2A execution. It adds the in-process dynamic teammate runtime beside the legacy path and preserves current routing compatibility while the production runtime stabilizes. In status surfaces, `dynamic-v1` means that the built-in in-process teammate path is configured and available under multi-agent mode; it does not imply exclusivity over legacy fallback or remote A2A coexistence.
