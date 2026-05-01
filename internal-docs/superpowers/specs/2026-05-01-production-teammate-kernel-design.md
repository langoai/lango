# Production Teammate Kernel Design

Date: 2026-05-01

## Purpose

Renew Lango's multi-agent runtime into a production-grade dynamic teammate system while preserving the existing `agent.multiAgent=true` user-facing configuration. The new runtime should make multi-agent execution deep, flexible, controllable, observable, and recoverable instead of relying on a static tool-less orchestrator tree.

This design treats the existing specialist agents (`operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`) as teammate types. The main agent becomes a coordinator-capable agent that can answer directly or autonomously create teammates during a user turn.

## Current Context

The repository already has several pieces of a teammate system:

- `internal/agentrt/control_tools.go` defines `agent_spawn`, `agent_wait`, and `agent_stop`.
- `internal/agentrt/coordinating_executor.go` wraps execution with delegation, budget, and recovery observation.
- `internal/agentrt/agent_run_projection.go` bridges agent run state into background task projection.
- `internal/background` provides asynchronous task execution.
- `internal/cli/cockpit` and its pages expose runtime, task, approval, and status surfaces.
- `openspec/specs/multi-agent-orchestration/spec.md` defines the current static orchestrator/sub-agent model.
- `docs/features/multi-agent.md` documents the current hierarchical model and role names.

The issue is not a complete absence of mechanisms. The issue is that spawn, background execution, specialist roles, approvals, recovery, and UI visibility are not unified behind one production runtime contract.

## Goals

1. Preserve `agent.multiAgent=true` as the natural entry point for multi-agent mode.
2. Replace the internal static orchestrator model with a dynamic teammate runtime.
3. Keep existing specialist roles as teammate types with their existing product meaning.
4. Let the main agent autonomously spawn teammates when complexity, parallelism, specialization, or risk isolation makes that useful.
5. Support both in-process async teammates and future separate-process or sandboxed worker execution through a common contract.
6. Enforce role-based maximum permissions plus spawn-time least privilege.
7. Allow runtime capability escalation through policy-first approval with CLI/TUI fallback.
8. Make every run inspectable, controllable, auditable, and recoverable.

## Non-Goals

1. Do not remove the existing specialist roles.
2. Do not introduce a new user-facing setting as the primary path for teammate mode.
3. Do not require the first implementation slice to complete separate-process workers.
4. Do not redesign the whole TUI in this workstream.
5. Do not copy implementation from Claude Code. Use it only as architectural inspiration for task identity, notifications, resumability, and coordinator visibility.

## Architecture

The center of the renewal is a `Teammate Kernel`. Existing `agent.multiAgent=true` remains the public switch, but internally it routes through the kernel instead of a static orchestrator tree.

### Main Agent

The main agent owns the user conversation and acts as coordinator. It can answer directly when no teammate is needed. It can spawn teammates when work benefits from parallelism, role specialization, risk isolation, long-running execution, or narrower context.

The main agent is not a tool-less router. It is responsible for synthesis, user communication, and deciding when to create, message, wait for, cancel, or resume teammates.

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

Each type defines role prompt fragments, maximum tool scope, default budgets, concurrency limits, escalation behavior, and allowed execution modes.

### Teammate Kernel

The kernel provides a stable contract for:

- `spawn`
- `status`
- `message`
- `wait`
- `cancel`
- `resume`
- `capability_request`
- `grant`
- `audit`
- `recover`

The same run model applies whether a teammate runs in-process, in a separate worker process, or in a sandbox.

### Execution Adapters

The initial adapter should be `InProcessAdapter`, integrated with the existing background manager and turn runner where practical. A `WorkerProcessAdapter` contract should exist early so the kernel design does not assume goroutine-only execution.

Adapter choice is policy-driven:

- Fast, low-risk work can use in-process execution.
- Dangerous tool scopes, long-running implementation work, large parallel jobs, or sandbox-required tasks can be promoted to worker-process or sandboxed execution.

### Policy and Approval Layer

Permissions are the intersection of:

1. Role maximum scope.
2. Spawn-time `allowed_tools`.
3. Active grants from session, always-allow, or approval policy.
4. Runtime capability escalation decisions.

Role maximum scope is a hard upper bound. A UI approval must not grant a teammate a tool outside its role maximum scope.

If a teammate needs additional permitted scope during execution, it submits a capability request. The policy layer first evaluates existing grants and always-allow rules. If the decision is unsafe or ambiguous, the request is surfaced through CLI/TUI approval UI. The main agent observes the event but does not hide the audit trail behind natural language.

### Operational Surfaces

CLI and TUI surfaces should show teammate runs as first-class runtime objects:

- Current state.
- Teammate type.
- Parent session and parent run.
- Active tool or blocked reason.
- Requested grant and approval status.
- Partial result.
- Final result.
- Execution mode.
- Recovery action.
- Direct channel status.
- Usage and budget.

The first implementation should reuse and extend existing run projection, background task snapshots, and cockpit task/runtime panels rather than replacing them.

## Components

### `internal/agentrt/teammate`

New package for kernel contracts and state machines. It should not depend on Cobra, Bubble Tea, or concrete UI code.

Core types:

- `Run`
- `RunID`
- `Type`
- `State`
- `Instruction`
- `Mailbox`
- `Message`
- `Grant`
- `CapabilityRequest`
- `AuditEvent`
- `ExecutionMode`
- `ExecutionAdapter`
- `RecoveryDecision`

### `TeammateTypeRegistry`

Registry for built-in and future user-defined teammate types. It maps the existing specialist role definitions into teammate runtime metadata:

- Prompt/instruction fragments.
- Maximum tool scope.
- Default allowed tools.
- Default execution mode.
- Budget defaults.
- Concurrency limits.
- Escalation policy.

The existing prefix mapping remains useful, but its meaning changes from static tool partitioning to role maximum scope.

### `CoordinatorRuntime`

Main-agent-facing runtime service and tool layer:

- `teammate_spawn`
- `teammate_wait`
- `teammate_message`
- `teammate_cancel`
- `teammate_status`
- `teammate_resume`
- `teammate_request_channel`

Existing `agent_spawn`, `agent_wait`, and `agent_stop` should become compatibility wrappers over this runtime instead of separate semantics.

### `CapabilityPolicy`

Responsible for final permission decisions. It evaluates:

- Role maximum scope.
- Spawn-time allowed tool subset.
- Existing session or always-allow grants.
- Tool safety level.
- Approval policy.
- Runtime escalation requests.

It returns one of:

- `allow`
- `deny`
- `needs_user_approval`
- `outside_role_scope`

### `Mailbox` and `ChannelManager`

The default communication model is main-agent hub routing. Teammates report progress and results to the main agent through structured messages.

Scoped direct channels are available only when opened by runtime policy. A direct channel has:

- Run/team ID.
- Participants.
- Purpose.
- TTL.
- Message count limit.
- Audit correlation ID.

Permission requests and user decisions must not be resolved inside direct teammate channels.

### Projection and Surfaces

Projection adapts kernel state into existing operator surfaces:

- Background task snapshots.
- Cockpit Tasks page.
- Cockpit Runtime panel.
- CLI run inspection.
- Approval UI.

Potential CLI surface:

- `lango agent runs`
- `lango agent run <id>`
- `lango agent message <id>`
- `lango agent cancel <id>`
- `lango agent grants <id>`

The exact command names should be verified against current CLI conventions during implementation planning.

### Audit and Recovery

Every significant runtime event must produce an audit event:

- Spawn requested.
- Spawn started.
- Adapter selected.
- Message sent.
- Direct channel opened/closed.
- Tool call requested.
- Capability requested.
- Grant allowed or denied.
- Run blocked.
- Timeout.
- Cancel.
- Crash or adapter failure.
- Recovery decision.
- Partial result.
- Final result.

Recovery policy classifies failures into:

- `retry`
- `retry_with_hint`
- `reroute`
- `resume`
- `summarize_partial`
- `cancel`
- `escalate_to_user`

Cause-class retry budgets prevent infinite loops.

## Data Flow

### User Turn to Decision

When `agent.multiAgent=true`, the main agent first determines whether it can answer directly. It creates teammates when work requires specialization, parallelism, isolation, long duration, or separate context.

The main agent's spawn decision must include a reason that is stored in audit events.

### Spawn Flow

`teammate_spawn` creates a `Run` with:

- Parent session ID.
- Parent run ID, if any.
- Teammate type.
- Instruction.
- Allowed tool subset.
- Budget.
- Execution mode preference.
- Spawn reason.
- Provenance.
- Audit correlation ID.

The kernel validates role scope and budget before selecting an execution adapter.

### Work and Report Flow

Teammates emit structured progress, partial result, blocked reason, and final result messages through the mailbox. The main agent consumes these events and synthesizes user-facing responses.

Teammate results should be structured enough to preserve:

- Summary.
- Evidence.
- Artifacts.
- Changed files, when applicable.
- Verification performed.
- Remaining risks.

### Capability Escalation Flow

When a teammate needs a tool outside its current allowed subset, it emits `CapabilityRequest`.

The policy layer evaluates the request. If allowed, a grant event updates the run's effective permissions. If user approval is required, the run moves to `blocked_waiting_approval`. If denied or outside role scope, the teammate receives a structured denial and can reroute, summarize partial work, or escalate to the main agent.

### Completion and Recovery Flow

Successful completion stores final result, usage, artifacts, and audit summary.

Timeout, cancellation, crash-like adapter failure, orphaned state, blocked approval timeout, and partial result all pass through recovery policy. Recovery may retry, resume, reroute, summarize partial work, cancel, or escalate to the user.

## State Model

The kernel should support at least these states:

- `queued`
- `running`
- `blocked_waiting_approval`
- `blocked_waiting_message`
- `waiting_on_teammate`
- `resuming`
- `completed`
- `failed`
- `cancelled`
- `orphaned`
- `recovering`

State transitions must be explicit and testable. Terminal states are `completed`, `failed`, and `cancelled`. `orphaned` is recoverable, not final.

## Guardrails

### Lifecycle

Every run has timeout, cancellation token, parent session, budget, max tool calls, max messages, and max direct-channel messages.

### Spawn Discipline

The main agent cannot spawn unlimited teammates. Limits apply at:

- Per turn.
- Per session active count.
- Per teammate type.
- Per execution mode.
- Per budget class.

### Permission Safety

Role maximum scope is absolute. Spawn-time allowed tools narrow the role scope. Runtime escalation can only expand within role maximum scope.

Dangerous filesystem, exec, payment, and secret operations require policy approval and usually user approval unless covered by a valid session or always-allow grant.

### Communication Safety

Direct channels are disabled by default and must include purpose, participants, TTL, and message limits. User decisions and permission grants go through runtime approval paths.

### Context Isolation

Teammates should not receive the entire parent context by default. The spawn instruction should include only necessary context, selected files/artifacts, and explicit references. The main agent consumes structured summaries to reduce context pollution.

### Backward Compatibility

`agent.multiAgent=true` continues to enable multi-agent behavior. Existing `agent_spawn`, `agent_wait`, and `agent_stop` remain available as wrappers while documentation and prompts transition to teammate terminology.

## Testing Strategy

### Kernel Unit Tests

Test lifecycle transitions, mailbox routing, direct channel TTL and limits, grant evaluation, and audit event emission.

### Compatibility Tests

Verify existing `agent_spawn`, `agent_wait`, and `agent_stop` contracts work through the teammate kernel. Verify `agent.multiAgent=true` activates the new runtime path.

### Policy Tests

Cover role maximum scope, spawn-time allowed tools, capability escalation, always-allow grants, session grants, denial, and outside-role-scope requests.

### Recovery Tests

Reproduce timeout, cancel, adapter failure, partial result, orphaned run, blocked approval timeout, and retry budget exhaustion.

### Integration Tests

Use fake models and fake execution adapters to deterministically test a main agent spawning planner/operator/navigator teammates, receiving partial results, handling a capability request, and synthesizing a final response.

### CLI/TUI Projection Tests

Verify cockpit and CLI surfaces read the same projection and show blocked reason, requested grant, active teammate, and final result.

## Rollout Plan

### Slice 1: Kernel and Compatibility

Create the kernel package, in-memory store, fake execution adapter, audit model, state transitions, and compatibility wrappers for existing control tools.

### Slice 2: In-Process Runtime Integration

Connect the kernel to existing background manager and turn runner paths. Ensure `agent.multiAgent=true` routes through teammate runtime while preserving current user settings.

### Slice 3: Permission Escalation

Implement capability requests, policy evaluation, approval integration, grant audit events, and blocked approval states.

### Slice 4: Projection and Operator Surface

Extend run projection, cockpit Tasks/Runtime views, and CLI inspection commands to expose teammate state and blocked reasons.

### Slice 5: Worker Process Contract

Add the worker process/sandbox adapter contract and a minimal gated implementation for high-risk or long-running tasks.

## Documentation and Spec Impact

Likely affected OpenSpec capabilities:

- `agent-runtime`
- `agent-control-plane`
- `multi-agent-orchestration`
- `sub-session-isolation`
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

Documentation must describe the new dynamic teammate runtime without claiming separate-process execution is complete until the adapter is implemented.

## Acceptance Criteria

1. `agent.multiAgent=true` uses the teammate runtime path without requiring a new primary setting.
2. Existing specialist roles are available as teammate types.
3. The main agent can autonomously spawn teammates with a recorded reason.
4. Runs expose state, parentage, teammate type, execution mode, budget, blocked reason, and final result.
5. Role maximum scope and spawn-time allowed tools are both enforced.
6. Capability escalation is policy-first and surfaces user approval only when required.
7. Direct teammate channels are disabled by default and audited when opened.
8. Timeout, cancel, adapter failure, orphaned runs, and partial results converge through recovery policy.
9. Existing `agent_spawn`, `agent_wait`, and `agent_stop` remain compatible through wrappers.
10. CLI/TUI projections can show current teammate status and blocked reason from the same source of truth.

