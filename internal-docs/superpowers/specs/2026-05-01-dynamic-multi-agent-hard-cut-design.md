# Dynamic Multi-Agent Hard Cut Design

## Goal

Replace the built-in multi-agent production path with a single dynamic teammate runtime. Built-in specialist work must no longer execute through legacy static ADK sub-agent delegation or `transfer_to_agent`. Remote A2A agents remain a separate execution model, but they continue to appear in the shared operator-facing control plane.

This change also closes three adjacent risks that matter during the cutover:

1. Capability runtime read-then-write races can leave stale `blocked_waiting_approval` projection state behind.
2. RunLedger write-through and projection behavior must remain aligned with the teammate runtime's canonical run identity chain.
3. The `agent_control` background category/channel must be treated as a first-class production path across runtime, approval, recovery, CLI/TUI, and docs.

## Scope

### In Scope

- Hard-cut built-in teammate execution to `agent_spawn`-driven runtime execution.
- Remove built-in production reliance on `transfer_to_agent`.
- Keep remote A2A execution distinct, while preserving shared operator visibility.
- Tighten capability approval state handling to reduce stale blocked projections.
- Audit and document RunLedger and `agent_control` secondary impacts.
- Update downstream artifacts that describe or expose built-in teammate execution.

### Out of Scope

- Replacing remote A2A execution with the built-in teammate runtime.
- Introducing new model-facing teammate tools beyond `agent_spawn`, `agent_wait`, and `agent_stop`.
- Reworking the broader approval UX beyond what is needed to preserve runtime correctness.
- Removing all `transfer_to_agent` support everywhere in the codebase. It remains available only for remote A2A bridging and narrow legacy interoperability boundaries.

## Target Architecture

### Built-In Teammate Runtime

Built-in teammate execution is always:

`root orchestrator decision -> agent_spawn -> AgentRunStore/Projection -> background.Manager submit -> ChildSession isolated run -> CapabilityRuntime/approval -> RunLedger write-through -> agent_wait/status/operator surfaces`

Built-in specialists stop being production ADK sub-agents. They become canonical teammate types resolved from the built-in registry:

- `operator`
- `navigator`
- `vault`
- `librarian`
- `automator`
- `planner`
- `chronicler`
- `ontologist`

`RequestedAgent` is no longer merely a loose hint for built-in teammates. It becomes the teammate type selector that chooses the built-in runtime contract, prompt defaults, role maximum scope, and operator labeling.

### Remote A2A Runtime

Remote A2A agents remain a distinct execution model. They are not absorbed into the built-in teammate runtime, and they do not inherit built-in capability gating semantics such as `DynamicAllowedTools` or built-in role max scope.

However, remote A2A runs continue to project into the shared control plane so that operators can inspect status, recovery, and parent/child relationships through common surfaces.

### Compatibility Boundary

`transfer_to_agent` is removed from the built-in production path. It may remain only where the runtime must bridge to remote A2A or maintain tightly bounded legacy interoperability. Built-in recovery and routing must no longer depend on it.

## Runtime Boundaries

### Orchestration

`internal/orchestration/tools.go` changes role:

- Built-in specialist definitions remain the routing and teammate-type registry.
- They no longer define production delegation targets for built-in execution.
- Their prompts no longer instruct built-in specialists to escalate via `transfer_to_agent("lango-orchestrator")`.

The root orchestrator prompt changes from "prefer `agent_spawn`" to a stronger contract:

- built-in teammate work MUST use `agent_spawn`
- `transfer_to_agent` MUST NOT be used for built-in specialists
- remote A2A routing remains separately described

Tool-name-shaped disambiguation text must also be tightened. Raw names such as `web_search` and `web_fetch` cannot be left in prompt text in a form that encourages hallucinated agent targets.

### Agent Runtime

`internal/agentrt/control_tools.go` becomes the only production entrypoint for built-in teammate creation. The built-in runtime is defined by:

- `AgentRun`
- `AgentRunProjection`
- `CapabilityRuntime`
- `background.Manager`
- `ChildSession`

The runtime lifecycle is explicit:

- `spawned`
- `running`
- terminal: `completed | failed | cancelled`

Projected conditions remain a separate operator-facing axis:

- `blocked_waiting_approval`
- `waiting_on_external`
- `recovering`

This keeps base execution state stable while allowing operator-visible waiting and recovery conditions to surface without fake terminal transitions.

### Capability Approval and Race Handling

The built-in runtime keeps the existing policy-first approval model:

- blocked tool call
- policy evaluation
- projection update
- approval/grant
- resumed tool allowance

But the stale projection race must be reduced. `HandleBlockedToolCall()` currently reads decision state and then writes projection state in separate steps. During the gap, another goroutine may grant the tool and leave a stale blocked condition behind.

The runtime contract for the hard cut is:

- evaluate the blocked call
- before writing a blocked projection, re-check grant/allowance state
- if the grant or allowlist is already effective, do not persist `blocked_waiting_approval`
- after approval, `ApplyGrant()` must update both projection state and future runtime allowance checks so the next tool attempt can pass without a second artificial block

This does not require a brand-new approval model. It requires the current one to become idempotent and consistent enough for the new runtime to be the only built-in path.

### Recovery

Built-in teammate recovery is reframed around control-plane runs rather than static transfers:

- fail current run and surface it
- spawn a new eligible built-in teammate run
- synthesize a direct root answer from available evidence

Built-in recovery must not bounce through `transfer_to_agent`.

## Canonical Identity and RunLedger

The hard cut preserves the existing canonical identity chain:

- `AgentRun.ID`
- background task ID
- RunLedger run ID

These three remain the same logical run identifier for built-in teammate execution.

RunLedger is not a secondary afterthought here. The write-through and projection layer remains responsible for durable run inspection and recovery visibility. The design therefore requires an explicit secondary impact audit covering:

- teammate spawn submission
- run status transitions
- projection sync markers
- approval-blocked conditions
- recovery states

If any built-in teammate state that operators depend on exists only in the control-plane projection and not in the durable RunLedger view, that mismatch must be called out and resolved or explicitly deferred.

## `agent_control` Category and Channel

The hard cut promotes `agent_control` from an implementation detail to a first-class production execution category for built-in teammates.

The audit must verify that `agent_control` is correctly propagated through:

- background submission origin/category
- recovery and status surfaces
- approval and blocked-state handling
- CLI/TUI runtime views
- public docs that describe teammate execution

The purpose is not to rename the channel unless necessary. The purpose is to prevent silent observability gaps after built-in execution is fully moved behind that path.

## Spec Impact

### Primary Specs

- `openspec/specs/multi-agent-orchestration/spec.md`
- `openspec/specs/agent-control-plane-tools/spec.md`

### Secondary Specs

- `openspec/specs/tool-execution-hooks/spec.md`
- `openspec/specs/tool-capability-layer/spec.md`
- `openspec/specs/run-ledger/spec.md`

### Required Contract Rewrites

1. Built-in multi-agent execution must be described as spawn-based runtime execution, not static ADK delegation.
2. Built-in `transfer_to_agent` escalation requirements must be removed or limited to remote interoperability contexts.
3. `agent_spawn` must be documented as the built-in production entrypoint, not merely an optional dynamic path.
4. Capability approval must explicitly describe the runtime-emitted blocked-call path and the grant/resume contract.
5. RunLedger must at least document the durable visibility expectations that remain true after the hard cut.

## Implementation Waves

### Wave 1: Contract Rewrite

- Rewrite `multi-agent-orchestration` around built-in spawn-only production execution.
- Rewrite `agent-control-plane-tools` so built-in teammate execution is anchored on `agent_spawn`.
- Update capability-related secondary specs for blocked-call and grant wiring.
- Record RunLedger secondary impact expectations.

### Wave 2: Runtime Cutover

- Remove built-in production reliance on `transfer_to_agent`.
- Convert built-in orchestration prompts and routing text to spawn-only semantics.
- Tighten tool-name disambiguation to reduce hallucinated agent targets.
- Add race mitigation in `CapabilityRuntime`.
- Audit and patch `agent_control` observability gaps.

### Wave 3: Operator Surfaces

- Update CLI/TUI/status wiring and language.
- Update docs and skills that still describe built-in execution as legacy-compatible first.
- Keep remote A2A execution clearly separate while preserving unified operator visibility.

## Testing and Verification

### Unit Tests

- built-in routing produces `agent_spawn`-driven execution rather than `transfer_to_agent`
- tool-name-shaped requests cannot produce built-in `transfer_to_agent` targets
- blocked tool -> grant -> retry no longer leaves a stale blocked projection when grant is already effective

### Integration Tests

- `AgentRun.ID == background task ID == RunLedger run ID` remains true for built-in teammates
- built-in teammate spawn/wait/approval/recovery works end-to-end
- remote A2A paths still work while remaining distinct from built-in capability enforcement

### Regression Tests

- `failed to find agent: web_search` class regression
- `agent_control` visibility across status and recovery surfaces
- public documentation and CLI/TUI labels match the actual production execution path

## Risks

### Risk: Hard cut introduces broad regressions

This is not a small runtime tweak. It changes the execution contract. The mitigation is a single coherent change with explicit wave sequencing, not a long-lived mixed model.

### Risk: Remote and built-in semantics blur again

The mitigation is to keep execution modes explicit. Shared operator visibility must not imply shared capability or execution rules.

### Risk: Projection state diverges from durable state

The mitigation is the RunLedger secondary impact audit and targeted follow-up where durable visibility is missing.

## Success Criteria

The hard cut is successful when all of the following are true:

1. Built-in production teammate work no longer depends on `transfer_to_agent`.
2. Built-in teammate execution always enters through the control plane and background-backed runtime path.
3. Remote A2A remains operational but clearly separate in execution semantics.
4. Approval and blocked-state handling are consistent enough that stale blocked projections are materially reduced.
5. Operator-facing status, recovery, and documentation all reflect the new execution contract without legacy ambiguity.
