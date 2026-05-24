# Dynamic Multi-Agent Hard Cut Design

## Goal

Replace the built-in multi-agent production path with a single dynamic teammate runtime. Built-in specialist work must no longer execute through legacy static ADK sub-agent delegation or `transfer_to_agent`. Remote A2A agents remain a separate execution model, and this design only claims shared operator-facing visibility for them where concrete wiring exists.

This change also closes three adjacent risks that matter during the cutover:

1. Capability runtime read-then-write races can leave stale `blocked_waiting_approval` projection state behind.
2. RunLedger write-through and projection behavior must remain aligned with the teammate runtime's canonical run identity chain.
3. The `agent_control` background category/channel must be treated as a first-class production path across runtime, approval, recovery, CLI/TUI, and docs.

## Scope

For this design, **built-in teammate** means any teammate name that resolves through `BuiltinTeammateTypes()` in `internal/agentrt/teammate_types.go`.

### In Scope

- Hard-cut built-in teammate execution to `agent_spawn`-driven runtime execution.
- Remove built-in production reliance on `transfer_to_agent`.
- Rewrite the 8 built-in embedded `AGENT.md` files so their escalation contract no longer targets `lango-orchestrator` through `transfer_to_agent`.
- Keep remote A2A execution distinct, and audit or document operator visibility boundaries honestly.
- Tighten capability approval state handling to reduce stale blocked projections.
- Audit and document RunLedger and `agent_control` secondary impacts.
- Update downstream artifacts that describe or expose built-in teammate execution.

### Out of Scope

- Replacing remote A2A execution with the built-in teammate runtime.
- Introducing new model-facing teammate tools beyond `agent_spawn`, `agent_wait`, and `agent_stop`.
- Reworking the broader approval UX beyond what is needed to preserve runtime correctness.
- Removing all `transfer_to_agent` support everywhere in the codebase. It remains available only for remote A2A bridging and narrow legacy interoperability boundaries.
- Claiming that remote A2A already has per-run `AgentRun` / `AgentRunProjection` visibility unless this change introduces and wires a concrete adapter.

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

The current codebase proves that remote A2A agents are loaded as ADK sub-agents and appear in orchestrator routing and listing surfaces. It does **not** prove that remote A2A requests already project into the same per-run `AgentRun` control plane as built-in teammate runs. This design therefore does not assume that behavior as an existing contract.

If shared per-run operator visibility for remote A2A is required in the same change, the implementation must add an explicit adapter and document its wiring. Otherwise the design remains honest: built-in teammate runs are hard-cut to the control plane, while remote A2A preserves its current separate execution and visibility model.

### Compatibility Boundary

`transfer_to_agent` is removed from the built-in production path. It may remain only where the runtime must bridge to remote A2A or maintain tightly bounded legacy interoperability. Built-in recovery and routing must no longer depend on it.

This boundary must be applied explicitly across the current `transfer_to_agent` inventory:

- `internal/agentregistry/defaults/*/AGENT.md` for all 8 built-in specialists: remove built-in escalation to `lango-orchestrator`
- `internal/orchestration/tools.go`: remove built-in escalation and routing language that still treats static transfer as a normal built-in path
- `internal/orchestration/orchestrator.go` / `BuildAgentTree`: retain only the root orchestration role plus remote/legacy compatibility where still needed; built-in specialists stop being production `SubAgents`
- `internal/skill/executor.go`: stop telling the model to use `transfer_to_agent('<specialist>')` for built-in specialization
- `internal/adk/agent.go`: keep hallucinated-agent recovery only for the remaining remote/legacy transfer surface, not as a primary built-in execution expectation

## Runtime Boundaries

### Orchestration

`internal/orchestration/tools.go` changes role:

- Built-in specialist definitions remain the routing and teammate-type registry.
- They no longer define production delegation targets for built-in execution.
- Their embedded prompts and the 8 built-in `AGENT.md` files no longer instruct built-in specialists to escalate via `transfer_to_agent("lango-orchestrator")`.

The root orchestrator prompt changes from "prefer `agent_spawn`" to a stronger contract:

- built-in teammate work MUST use `agent_spawn`
- `transfer_to_agent` MUST NOT be used for built-in specialists
- remote A2A routing remains separately described

Tool-name-shaped disambiguation text must also be tightened. Raw names such as `web_search` and `web_fetch` cannot be left in prompt text in a form that encourages hallucinated agent targets.

`BuildAgentTree` is not simply left untouched. Under the hard cut, it must stop wiring built-in specialists into the production ADK sub-agent tree. The retained tree, if any, exists for:

- the root orchestrator agent itself
- remote A2A sub-agents
- tightly bounded legacy interoperability that is explicitly documented

Any `transfer_to_agent("lango-orchestrator")` emitted from a built-in teammate prompt after the cut is a regression, not a valid steady-state path.

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

Projected conditions remain a separate operator-facing axis, and this change intentionally stays aligned with the existing code-level names rather than inventing new names:

- `blocked_waiting_approval`
- `blocked_waiting_message`
- `waiting_on_teammate`
- `resuming`
- `orphaned`
- `recovering`

The hard cut does not introduce a new `waiting_on_external` condition. Instead:

- `waiting_on_teammate` remains the waiting state for spawned teammate coordination
- `blocked_waiting_message` remains reserved for explicit message-waiting flows if they still exist
- `resuming` and `recovering` remain distinct operator-visible repair phases
- `orphaned` remains the inconsistency signal when run ownership or execution continuity breaks

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
- before writing a blocked projection, re-read the latest run state and re-check grant/allowance state
- if the grant or allowlist is already effective, do not persist `blocked_waiting_approval`
- after approval, `ApplyGrant()` must update both projection state and future runtime allowance checks so the next tool attempt can pass without a second artificial block

The chosen approach is optimistic recheck rather than a global execution mutex:

1. evaluate
2. fetch latest run / latest allowlist
3. re-check `hasGrant(...) || containsTool(latest.AllowedTools, toolName)`
4. only then emit the blocked projection patch

This does not require a brand-new approval model. It requires the current one to become idempotent and consistent enough for the new runtime to be the only built-in path.

This procedure narrows but does not eliminate the TOCTOU window between step 3 and step 4. A grant can still arrive after the recheck and before the blocked projection write. The remaining window is treated as benign because `ApplyGrant()` is idempotent and clears the projected condition. Success is therefore measured on the final observed run state, not on every intermediate transition.

### Recovery

Built-in teammate recovery is reframed around control-plane runs rather than static transfers:

- fail current run and surface it
- spawn a new eligible built-in teammate run
- synthesize a direct root answer from gathered evidence, analogous to the existing `RecoveryDirectAnswer` semantics

Built-in recovery must not bounce through `transfer_to_agent`.

This also reframes the current ADK hallucinated-agent retry contract. After the cut, `failed to find agent: <name>` recovery is no longer a normal built-in safeguard. It becomes a remote/legacy compatibility safeguard only.

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

The audit must produce one verdict for each audited item:

1. recorded in RunLedger with acceptable fidelity
2. not recorded, but no operator-facing consistency risk
3. not recorded and creates a consistency risk, requiring a same-change fix or an explicit follow-up issue before archive

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
- `openspec/specs/agent-routing/spec.md`

### Secondary Specs

- `openspec/specs/tool-execution-hooks/spec.md`
- `openspec/specs/tool-capability-layer/spec.md`
- `openspec/specs/run-ledger/spec.md`
- `openspec/specs/adk-architecture/spec.md`
- `openspec/specs/agent-registry/spec.md`

### Required Contract Rewrites

1. Built-in multi-agent execution must be described as spawn-based runtime execution, not static ADK delegation.
2. Built-in `transfer_to_agent` escalation requirements must be removed or limited to remote interoperability contexts.
3. `agent_spawn` must be documented as the built-in production entrypoint, not merely an optional dynamic path.
4. Capability approval must explicitly describe the runtime-emitted blocked-call path and the grant/resume contract.
5. RunLedger must at least document the durable visibility expectations that remain true after the hard cut.
6. Built-in embedded `AGENT.md` and registry-backed prompt files must stop encoding `transfer_to_agent("lango-orchestrator")` as a built-in escalation requirement.
7. ADK hallucinated-agent retry requirements must be narrowed to the remaining remote/legacy transfer surface.

## Change Boundary

This work is intended to land as **one OpenSpec change with multiple implementation slices**, not as separate independently archived changes. The reason is simple: the spec rewrite and the runtime cutover must be archived in a state where they still match each other.

## Implementation Slices

### Slice 1: Contract Rewrite

- Rewrite `multi-agent-orchestration` around built-in spawn-only production execution.
- Rewrite `agent-control-plane-tools` so built-in teammate execution is anchored on `agent_spawn`.
- Rewrite `agent-routing` / `agent-registry` expectations that still require built-in prompt files and built-in specialist definitions to escalate through `transfer_to_agent`.
- Reframe `adk-architecture` hallucinated-agent retry as remote/legacy compatibility behavior.
- Update capability-related secondary specs for blocked-call and grant wiring.
- Record RunLedger secondary impact expectations.
- At plan time, run a concrete grep inventory for `transfer_to_agent`, `lango-orchestrator`, sub-agent escalation, and `failed to find agent` patterns across all primary and secondary specs before drafting delta text.

### Slice 2: Runtime Cutover

- Remove built-in production reliance on `transfer_to_agent`.
- Convert built-in orchestration prompts and routing text to spawn-only semantics.
- Rewrite all 8 built-in embedded `AGENT.md` files to remove built-in `transfer_to_agent("lango-orchestrator")` escalation.
- Retain `BuildAgentTree` only for the root orchestrator plus remote/legacy sub-agent composition, not as the production built-in specialist execution tree.
- Remove built-in `transfer_to_agent('<specialist>')` guidance from `internal/skill/executor.go`.
- Update `internal/adk/agent.go` hallucinated-agent recovery messaging so it no longer suggests built-in transfer retries; built-in routing failures must nudge toward `agent_spawn` or direct root recovery instead.
- Verify whether `internal/skill/executor.go` can resolve `skill.Agent` against `BuiltinTeammateTypes()` at the call site.
- If that resolution is feasible, switch built-in fork guidance to `agent_spawn`-style delegation and retain transfer wording only for remote/legacy targets. If it is not feasible, treat that result as a scope decision that must be resolved before runtime cutover archive.
- Tighten tool-name disambiguation to reduce hallucinated agent targets.
- Add race mitigation in `CapabilityRuntime`.
- Audit and patch `agent_control` observability gaps.
- Verify ADK behavior when `BuildAgentTree` returns a root-only tree with no production built-in sub-agents and no remote A2A agents configured; ensure the root prompt remains coherent without sub-agent listings.

Note: `internal/skill/executor.go` defaults `skill.Agent` to `"operator"` when unset. That means most fork-style skill executions currently follow the built-in path, making this switch high-impact rather than peripheral cleanup.

### Slice 3: Operator Surfaces

- Update CLI/TUI/status wiring and language.
- Update docs and skills that still describe built-in execution as legacy-compatible first.
- Keep remote A2A execution clearly separate, and only claim shared operator visibility where concrete wiring exists.
- Add upgrade notes for users who copied or derived custom `AGENT.md` files from the embedded defaults so they can update the old built-in escalation pattern.

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
- repeated grant/block interleavings do not leave a stale `blocked_waiting_approval` condition in the final observed run state

## Risks

### Risk: Hard cut introduces broad regressions

This is not a small runtime tweak. It changes the execution contract.

Mitigations:

- require a pre-merge regression matrix that covers built-in spawn/wait/recovery, remote A2A compatibility, and hallucinated-agent regressions
- keep the change boundary single and coherent so spec/code drift does not accumulate
- require the RunLedger audit verdict before archive rather than leaving durability questions implicit

Feature-flag and shadow-mode rollout were considered but rejected for this change. The cutover rewrites system prompts and tool-surface contracts that the LLM consumes directly. Running the old and new prompt contracts side-by-side would reintroduce the exact ambiguity the hard cut is meant to remove. The mitigation is therefore quality-of-cut, not staged prompt coexistence.

### Risk: Remote and built-in semantics blur again

The mitigation is to keep execution modes explicit. Shared operator visibility must not imply shared capability or execution rules.

### Risk: Projection state diverges from durable state

The mitigation is the RunLedger secondary impact audit and targeted follow-up where durable visibility is missing.

## Success Criteria

The hard cut is successful when all of the following are true:

1. Built-in production teammate work no longer depends on `transfer_to_agent`.
2. Built-in teammate execution always enters through the control plane and background-backed runtime path.
3. Remote A2A remains operational but clearly separate in execution semantics.
4. Approval and blocked-state handling are strong enough that the regression suite observes zero stale final `blocked_waiting_approval` states after repeated grant/block interleavings for the covered test cases.
5. Operator-facing status, recovery, and documentation all reflect the new execution contract without legacy ambiguity.
