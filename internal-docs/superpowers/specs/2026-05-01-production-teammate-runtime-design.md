# Production Teammate Runtime Design

Date: 2026-05-01
Revision: 2

## Purpose

Renew Lango's multi-agent runtime into a production-grade dynamic teammate system while preserving the existing `agent.multiAgent=true` user-facing configuration. The runtime should make multi-agent execution deep, flexible, controllable, observable, and recoverable without duplicating the agent runtime assets that already exist.

This design treats the existing specialist agents (`operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`) as teammate types. The main agent becomes a coordinator-capable agent that can answer directly or autonomously create teammates during a user turn.

The implementation direction is not "add a parallel teammate subsystem." The direction is "reframe the existing `agentrt`, background, RunLedger, and child-session machinery into one production teammate contract."

## Current Context

The repository already has several pieces of the target runtime:

- `internal/agentrt/control_tools.go` defines `agent_spawn`, `agent_wait`, and `agent_stop`.
- `internal/agentrt/agent_run.go` defines `AgentRun`, `AgentRunStatus`, `ChildSession`, and `AllowedTools`.
- `internal/agentrt/agent_run_projection.go` unifies `AgentRun.ID` with the background task ID through `background.Projection`.
- `internal/background/manager.go` provides in-process async execution, timeout, cancellation, retry, task snapshots, and projection hooks.
- `internal/agentrt/coordinating_executor.go` wraps execution with delegation, budget, and recovery observation.
- `internal/agentrt/recovery.go` provides `RecoveryPolicy`, `CauseClass`, per-class retry limits, and backoff.
- `internal/agentrt/delegation_guard.go` provides circuit breaker state for agents and providers.
- `internal/session/child.go` defines `ChildSession` and `ChildSessionStore`.
- `openspec/specs/sub-session-isolation/spec.md` defines isolated child-session behavior, summary-only merge, and raw child history isolation.
- `openspec/specs/run-ledger/spec.md` defines RunLedger as the Task OS durable execution engine.
- `internal/cli/cockpit` and `internal/cli/agent` expose runtime, task, status, trace, graph, and agent inspection surfaces.
- `openspec/specs/multi-agent-orchestration/spec.md` currently defines the static tool-less orchestrator/sub-agent model.
- `docs/features/multi-agent.md` documents the current hierarchical model and role names.

The issue is not a complete absence of mechanisms. The issue is that spawn, background execution, specialist roles, approvals, recovery, child sessions, RunLedger, and UI visibility are not unified behind one production teammate contract.

## Goals

1. Preserve `agent.multiAgent=true` as the natural entry point for multi-agent mode.
2. Replace the internal static tool-less orchestrator contract with a dynamic teammate runtime contract.
3. Keep existing specialist roles as teammate types with their existing product meaning.
4. Let the main agent autonomously spawn teammates when complexity, parallelism, specialization, or risk isolation makes that useful.
5. Use the existing in-process background execution path for v1.
6. Enforce role-based maximum permissions plus spawn-time least privilege.
7. Allow runtime capability escalation through policy-first approval with CLI/TUI fallback.
8. Make every run inspectable, controllable, auditable, and recoverable.
9. Extend existing `agentrt`, background, RunLedger, and child-session assets instead of duplicating them.

## Non-Goals

1. Do not remove the existing specialist roles.
2. Do not introduce a new user-facing setting as the primary path for teammate mode.
3. Do not create a new parallel `internal/agentrt/teammate` package in v1 unless implementation planning proves the existing `agentrt` package cannot hold the extension cleanly.
4. Do not define or implement a separate-process worker adapter in v1.
5. Do not redesign the whole TUI in this workstream.
6. Do not expose both `agent_*` and `teammate_*` LLM tool families at the same time.
7. Do not copy implementation from Claude Code. Use it only as architectural inspiration for task identity, notifications, resumability, and coordinator visibility.

## Spec Rewrite Scope

The current `multi-agent-orchestration` spec is not a small downstream impact. It is the primary contract that must be rewritten before code changes switch `agent.multiAgent=true` to the dynamic model.

### Deprecated Requirements

These current requirements are incompatible with the dynamic teammate model and should be deprecated in the rewrite wave:

| Existing requirement area | Reason |
| --- | --- |
| Orchestrator has `Tools: nil` | The main agent is coordinator-capable and may answer directly or use control tools. |
| Orchestrator MUST delegate all tool-requiring tasks | The new runtime allows autonomous teammate spawning but does not force static transfer for every tool-requiring request. |
| Static `BuildAgentTree` as the core multi-agent shape | Teammates are runtime-created work units, not only prebuilt ADK sub-agents. |
| Static `RoleToolSet` partition as execution authority | Prefix mapping becomes role maximum scope, not final runtime assignment. |
| `transfer_to_agent` as the only handoff primitive | Background run identity, child sessions, and control tools become the teammate execution primitive. |

### Reframed Requirements

These requirements remain valuable but must be reframed:

| Existing requirement area | New framing |
| --- | --- |
| Specialist roles | Built-in teammate types with role prompts, maximum tool scope, default budget, and isolation policy. |
| Prefix partitioning | Role maximum scope and default tool affinity. Spawn-time `AllowedTools` narrows it. |
| Capability descriptions | Teammate type descriptions used by main-agent routing and operator surfaces. |
| Remote A2A agents | Candidate teammate providers or remote teammate types, after the in-process path is stable. |
| Re-routing protocol | Recovery policy plus main-agent synthesis and reroute decisions. |
| Event author identity | Teammate run and child-session authorship must remain traceable in events and summaries. |

### New Requirements

The rewrite wave should introduce requirements for:

- Dynamic teammate run creation under `agent.multiAgent=true`.
- Main-agent direct answer and spawn decision protocol.
- Spawn reason audit.
- Role maximum scope plus spawn-time `AllowedTools`.
- Capability request and policy-first approval.
- ChildSession isolation for teammate runs.
- RunLedger/background projection as durable task authority.
- Teammate run projection for CLI/TUI inspection.
- Recovery behavior for timeout, cancel, blocked approval, partial result, and orphaned runs.

This rewrite should be its own OpenSpec wave before implementation. It is the contract boundary for every later slice.

## Source-Of-Truth Boundaries

The production teammate runtime has layered authority. No single new `Run` type should replace all existing state.

| Layer | Authority | Existing assets | Runtime role |
| --- | --- | --- | --- |
| Durable task authority | RunLedger when enabled, otherwise background task lifecycle plus projection | `internal/runledger`, `background.Projection`, `background.Manager` | Canonical task identity, durable state, journal/snapshot reads, retry/resume foundation. |
| Agent run control projection | Agent runtime control state | `AgentRun`, `AgentRunStore`, `AgentRunProjection` | Teammate identity, requested type, instruction, child session key, allowed tools, cancel function, result/error projection. |
| Context isolation | Child session store | `ChildSession`, `ChildSessionStore`, `ChildSessionServiceAdapter`, `StructuredSummarizer` | Isolated teammate transcript, summary-only merge, discard note, raw child history exclusion. |
| Policy layer | Coordinator runtime policy | `RecoveryPolicy`, `DelegationGuard`, approval policy, access-control hooks | Spawn discipline, role scope, grants, recovery decisions, circuit breaker decisions. |
| Operator projection | CLI/TUI read model | Cockpit pages, runtime tracker, `internal/cli/agent`, background snapshots | Human-readable run state, blocked reason, grant request, result, audit and trace inspection. |

The teammate runtime should sit above these assets as orchestration policy and projection glue. It should not create a second durable run store unless RunLedger integration explicitly requires a new projection table.

## Asset Inventory And Migration

The implementation should extend existing symbols first.

| New concept | Existing symbol or location | Action | Notes |
| --- | --- | --- | --- |
| Teammate run ID | `AgentRun.ID`, `background.Task.ID`, RunLedger `run_id` | Keep and extend | Preserve the existing canonical ID unification. |
| Teammate status | `AgentRunStatus`, `background.Status`, RunLedger snapshot status | Extend via fields/projection before adding many enum states | Avoid a premature 11-state enum unless producer/consumer wiring requires it. |
| Teammate type | `AgentRun.RequestedAgent`, agent registry definitions | Extend/rename later | `RequestedAgent` can represent teammate type in v1. A later rename can follow after spec rewrite. |
| Spawn instruction | `AgentRun.Instruction`, background prompt | Keep | Spawn reason and role metadata can be added around it. |
| Tool subset | `AgentRun.AllowedTools`, DynamicAllowedTools context | Keep and enforce | Add role maximum scope validation before storing. |
| Context isolation | `AgentRun.ChildSession`, `ChildSessionStore` | Keep and require | Every isolated teammate run should carry the child session key. |
| In-process execution | `background.Manager`, `AgentRunProjection` | Keep | This is the v1 execution path. |
| Recovery | `RecoveryPolicy`, `CauseClass`, `DelegationGuard` | Extend | Add teammate-aware outcomes only where missing. |
| Audit event | Event bus, RunLedger journal, trace events | Decide in implementation plan | The design requires audit events; the storage target must be specified per slice. |
| Operator read model | Cockpit Tasks/Runtime pages, `lango agent status/list/trace/graph/metrics` | Extend | Add runs/grants only after CLI command names are locked. |
| Direct messaging | No current `agent_message`; spec excludes it from initial tools | Add later | Direct teammate channels are not v1. |
| Worker process | None | Separate spike | Do not define v1 interfaces for this yet. |

## Architecture

### Main Agent

The main agent owns the user conversation and acts as coordinator. It can answer directly when no teammate is needed. It can spawn teammates when work benefits from parallelism, role specialization, risk isolation, long-running execution, or narrower context.

The main agent is not a tool-less router. It is responsible for synthesis, user communication, and deciding when to create, wait for, cancel, or recover teammate runs.

### Teammate Types

Existing specialist agents become teammate types:

- `planner`: planning, decomposition, and strategy; no direct tools by default.
- `operator`: local execution, file operations, and skill execution.
- `navigator`: browser and web navigation work.
- `vault`: cryptography, secrets, payment, wallet, and signing work.
- `librarian`: knowledge search, RAG, graph traversal, learning, and skill management.
- `automator`: background, cron, workflow, and scheduled work.
- `chronicler`: memory, observations, reflections, and recall.
- `ontologist`: ontology types, entities, facts, conflicts, and ingestion.

Each type defines role prompt fragments, maximum tool scope, default budgets, concurrency limits, escalation behavior, and isolation policy. In v1, this metadata should live near existing agent registry and `agentrt` control-plane code rather than in a separate runtime package.

### Runtime Contract

The teammate runtime contract is a policy contract over existing assets:

- Spawn creates or updates `AgentRun`, registers the canonical ID with projection, and submits background work.
- Status reads from `AgentRunStore`, background snapshots, and RunLedger where configured.
- Wait polls the existing run store until terminal status or timeout.
- Cancel uses existing cancellation hooks.
- Recovery uses `RecoveryPolicy` and `DelegationGuard`.
- Context isolation uses `ChildSessionStore`.
- Permission narrowing uses `AgentRun.AllowedTools` and DynamicAllowedTools enforcement.
- Capability escalation adds a policy and approval layer around allowed-tool expansion.

### V1 Execution Model

V1 is in-process only. It uses `background.Manager` and the existing runner path. Worker-process and sandboxed teammates are deferred to a later spike after the in-process source-of-truth model is stable.

This is an honest scope boundary. The design should not promise a worker-process contract in v1.

### Policy And Approval Layer

Permissions are the intersection of:

1. Role maximum scope.
2. Spawn-time `allowed_tools`.
3. Active grants from session, always-allow, or approval policy.
4. Runtime capability escalation decisions.

Role maximum scope is a hard upper bound. A UI approval must not grant a teammate a tool outside its role maximum scope.

If a teammate needs additional permitted scope during execution, it submits a capability request. The policy layer first evaluates existing grants and always-allow rules. If the decision is unsafe or ambiguous, the request is surfaced through CLI/TUI approval UI. The main agent observes the event but does not hide the audit trail behind natural language.

### Operational Surfaces

CLI and TUI surfaces should show teammate runs as first-class runtime objects:

- Current status.
- Teammate type.
- Parent session and child session.
- Active tool or blocked reason.
- Requested grant and approval status.
- Partial result.
- Final result.
- Recovery action.
- Usage and budget.

The implementation should reuse and extend existing run projection, background task snapshots, `internal/cli/agent`, and cockpit task/runtime panels.

## LLM Tool Surface Decision

V1 exposes one LLM-facing control-tool family: existing `agent_spawn`, `agent_wait`, and `agent_stop`.

The `teammate_*` names remain product/design language until the runtime is stable. They should not be exposed to the model alongside `agent_*` tools.

| Tool surface | V1 decision | Reason |
| --- | --- | --- |
| `agent_spawn` | Keep and extend internally | Existing spec and tests already define it. |
| `agent_wait` | Keep | Existing polling contract remains valid. |
| `agent_stop` | Keep | Existing cancellation contract remains valid. |
| `agent_message` | Do not add in v1 | Existing spec explicitly excludes it from initial tools. |
| `teammate_spawn` and related tools | Do not expose in v1 | Avoid duplicate routing surface and model confusion. |

If a later release renames the public model-facing tools, it should use a separate alias-then-deprecate change with prompts, docs, and tests updated together.

## State Projection Model

The runtime should not jump directly from 5 `AgentRunStatus` values to an 11-value enum. V1 should preserve the current terminal model and add projection fields only where producers and consumers are known.

### Existing Base Status

| Base state | Existing producer | Existing consumer | Meaning |
| --- | --- | --- | --- |
| `spawned` | `agent_spawn`, `AgentRunProjection.SyncTask(Pending)` | `agent_wait`, CLI/TUI projections | Run exists but background execution has not started. |
| `running` | `AgentRunProjection.SyncTask(Running)` | `agent_wait`, runtime tracker, operator surfaces | Teammate is executing. |
| `completed` | `AgentRunProjection.SyncTask(Done)` | `agent_wait`, result readers, summaries | Teammate finished successfully. |
| `failed` | `AgentRunProjection.SyncTask(Failed)` | recovery and operator surfaces | Teammate failed with error. |
| `cancelled` | `AgentRunStore.Cancel`, `AgentRunProjection.SyncTask(Cancelled)` | `agent_wait`, operator surfaces | User/runtime cancelled the run. |

### Projected Runtime Conditions

These conditions can be represented as fields or derived view state before adding enum values:

| Condition | Producer | Consumer | Trigger to clear |
| --- | --- | --- | --- |
| `blocked_waiting_approval` | Capability policy creates approval request | Approval UI, main agent event observer, CLI/TUI run view | Grant or denial decision. |
| `blocked_waiting_message` | Future direct-message or clarification request flow | Main agent or operator surface | Message received or run cancelled. |
| `waiting_on_teammate` | Parent run records child dependency | Main agent, run projection | Child terminal state or cancellation. |
| `resuming` | Recovery or user resume command | Background manager/runner, operator surface | New attempt starts running or fails. |
| `orphaned` | Reconciliation job detects run without active task/session | Recovery policy, operator surface | Recovery decision records resume, fail, or cancel. |
| `recovering` | Recovery policy begins action | Main agent, trace, operator surface | Recovery action completes. |

Each projected condition needs an explicit storage location during implementation planning. Candidate fields include `BlockedReason`, `WaitingOnRunID`, `RecoveryState`, `ExecutionMode`, `BudgetSnapshot`, and `GrantRequestID`. If these become durable columns, the OpenSpec change must include schema and migration requirements.

## Data Flow

### User Turn To Decision

When `agent.multiAgent=true`, the main agent first determines whether it can answer directly. It creates teammates when work requires specialization, parallelism, isolation, long duration, or separate context.

The main agent's spawn decision must include a reason that is stored in audit or trace events.

### Spawn Flow

`agent_spawn` remains the v1 model-facing tool. It should create or extend an `AgentRun` with:

- Parent session ID.
- Parent run ID, if any.
- Teammate type in `RequestedAgent`.
- Instruction.
- Allowed tool subset.
- Spawn reason.
- Child session key when isolation is active.
- Audit correlation ID or trace ID.

The spawn path validates role maximum scope and budget before registering the ID with `AgentRunProjection` and submitting the in-process background task.

### Work And Report Flow

Teammates execute in isolated child sessions when isolation is active. Raw child messages remain out of parent persistence. The main agent receives structured summaries, progress events, and final results through existing event/trace/projection channels.

Teammate results should preserve:

- Summary.
- Evidence.
- Artifacts.
- Changed files, when applicable.
- Verification performed.
- Remaining risks.

### Capability Escalation Flow

When a teammate needs a tool outside its current allowed subset but inside role maximum scope, it emits a capability request.

The policy layer evaluates the request. If allowed, a grant event updates the run's effective permissions. If user approval is required, the projected condition becomes `blocked_waiting_approval`. If denied or outside role scope, the teammate receives a structured denial and can reroute, summarize partial work, or escalate to the main agent.

### Completion And Recovery Flow

Successful completion stores final result, usage, artifacts, and summary.

Timeout, cancellation, adapter failure, orphaned state, blocked approval timeout, and partial result all pass through recovery policy. Recovery may retry, retry with hint, reroute, summarize partial work, cancel, or escalate to the user.

## Guardrails

### Lifecycle

Every run has timeout, cancellation token, parent session, child session when isolated, budget, max tool calls, and max messages.

### Spawn Discipline

The main agent cannot spawn unlimited teammates. Limits apply at:

- Per turn.
- Per session active count.
- Per teammate type.
- Per budget class.

### Permission Safety

Role maximum scope is absolute. Spawn-time allowed tools narrow the role scope. Runtime escalation can only expand within role maximum scope.

Dangerous filesystem, exec, payment, and secret operations require policy approval and usually user approval unless covered by a valid session or always-allow grant.

### Communication Safety

Direct teammate channels are not v1. When introduced later, they must include purpose, participants, TTL, and message limits. User decisions and permission grants must go through runtime approval paths.

### Context Isolation

Teammates should not receive the entire parent context by default. The spawn instruction should include only necessary context, selected files/artifacts, and explicit references. ChildSession provides isolated writes and summary-only merge/discard behavior.

### Backward Compatibility

`agent.multiAgent=true` continues to enable multi-agent behavior. Existing `agent_spawn`, `agent_wait`, and `agent_stop` remain the model-facing control tools while documentation and prompts transition to teammate terminology.

## Testing Strategy

### Spec Rewrite Tests

The first OpenSpec wave should update `multi-agent-orchestration` and related specs before implementation changes. Tests should fail if code still assumes the old tool-less orchestrator contract after the rewrite wave is applied.

### Asset Mapping Tests

Test that the spawn path preserves canonical ID unification across `AgentRun`, background task ID, and RunLedger projection when enabled.

### Compatibility Tests

Verify existing `agent_spawn`, `agent_wait`, and `agent_stop` contracts keep working after the teammate policy layer is added.

### Policy Tests

Cover role maximum scope, spawn-time allowed tools, capability escalation, always-allow grants, session grants, denial, and outside-role-scope requests.

### Child-Session Tests

Verify isolated teammate runs use child sessions, preserve summary-only merge, discard raw child history on failure, and retain classified incomplete causes.

### Recovery Tests

Reproduce timeout, cancel, background failure, partial result, orphaned run, blocked approval timeout, and retry budget exhaustion.

### CLI/TUI Projection Tests

Verify cockpit and CLI surfaces read the same projection and show blocked reason, requested grant, active teammate, and final result.

## Rollout Plan

### Slice A: OpenSpec Rewrite

Rewrite `multi-agent-orchestration` from static tool-less orchestrator to dynamic teammate runtime. Classify old requirements as deprecated, reframed, or replaced. Update related specs only where the contract boundary requires it.

No runtime code changes should switch behavior before this contract wave is accepted.

### Slice B: Existing Asset Mapping And In-Process Control Path

Extend `AgentRun`, `AgentRunProjection`, and the existing `agent_*` tools as needed so the teammate policy layer can use the current in-process background execution path. Preserve the current model-facing tool names.

### Slice C: Main-Agent Prompt And Dynamic Spawn Policy

Update the main-agent multi-agent prompt so `agent.multiAgent=true` means coordinator-capable main agent plus dynamic teammate spawning. Ensure spawn decisions include a recorded reason and respect role maximum scope.

### Slice D: Capability Escalation And Approval

Implement capability requests, policy evaluation, approval integration, grant audit events, and blocked approval projection.

### Slice E: Operator Projection

Extend run projection, cockpit Tasks/Runtime views, and CLI inspection commands to expose teammate type, blocked reason, grant request, child session, and final result.

### Separate Spike: Worker Process And Sandbox Execution

Investigate worker-process or sandboxed teammate execution after the in-process teammate runtime is stable. This spike should define whether a new execution adapter interface is necessary.

## Documentation And Spec Impact

Primary OpenSpec wave:

- `multi-agent-orchestration`

Secondary impacted specs:

- `agent-control-plane-tools`
- `agent-runtime`
- `sub-session-isolation`
- `run-ledger`
- `agent-turn-tracing`
- `approval-flow`
- `approval-policy`
- `cockpit-pages`
- `cockpit-status-page`
- `tui-runtime-visibility`
- `cli-agent-inspection`

Likely affected public docs:

- `docs/features/multi-agent.md`
- CLI docs for agent/runtime inspection, once command names are finalized.

Documentation must describe the new dynamic teammate runtime without claiming worker-process execution exists.

## Acceptance Criteria

1. `agent.multiAgent=true` uses the dynamic teammate runtime path without requiring a new primary setting.
2. The `multi-agent-orchestration` spec is rewritten before behavior changes depend on the new model.
3. Existing specialist roles are available as teammate types.
4. The main agent can autonomously spawn teammates with a recorded reason.
5. Existing `agent_spawn`, `agent_wait`, and `agent_stop` remain the only v1 model-facing control tools.
6. `AgentRun`, background task ID, and RunLedger projection preserve canonical identity.
7. ChildSession remains the context isolation primitive for teammate execution.
8. Role maximum scope and spawn-time `AllowedTools` are both enforced.
9. Capability escalation is policy-first and surfaces user approval only when required.
10. Timeout, cancel, background failure, orphaned runs, and partial results converge through recovery policy.
11. CLI/TUI projections can show current teammate status and blocked reason from the same source of truth.
12. Worker-process and sandboxed teammate execution remain outside v1 and are handled as a separate spike.

