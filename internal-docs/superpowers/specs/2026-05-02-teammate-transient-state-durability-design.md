# Teammate Transient State Durability Design

## Goal

Close the highest-value durable visibility gap left by the built-in multi-agent hard cut by mirroring approval-blocked teammate transient state into RunLedger.

This change covers:

- `blocked_waiting_approval`
- `blocked_reason`
- `grant_request_id`

The goal is **mirror only**. This change does not turn RunLedger into the authoritative read path for teammate runtime status. It makes the current operator-visible approval-blocked state reconstructible from durable state after restart or replay.

## Scope

### In Scope

- Durable mirror for approval-blocked teammate state
- Journal events for state transitions
- Snapshot fields for latest transient state
- Best-effort mirror semantics with explicit drift logging and metrics

### Out of Scope

- Replacing `AgentRunStore` or `AgentRunProjection` as the primary live read model
- Rewriting `agent_wait` or CLI/TUI to read from RunLedger first
- Approval UI redesign
- Broader recovery-engine redesign
- Introducing the first production `RecoveryState` writer
- Durable mirror for `recovery_state` before a production writer exists

## Problem

The hard-cut audit closed most built-in teammate runtime gaps, but two operator-visible state families remained marked as follow-up in the archived change:

- approval-blocked conditions
- recovery states

Today `approval-blocked` state is visible through the control-plane projection, but it is not durably mirrored in a way that allows a teammate run's blocked state to be reconstructed from RunLedger after restart.

`recovery_state` is a different problem. The archived audit is explicit that no production writer currently persists `AgentRun.RecoveryState`. That means there is no reliable source state to mirror yet. This change therefore narrows scope deliberately: it closes the approval-blocked durability gap now, and leaves `recovery_state` in follow-up until a production writer exists.

That leaves an operator-visible truth gap:

- live surfaces can say why a teammate is blocked
- durable state cannot fully explain the same run later

## Target Architecture

### Mirror Strategy

This change uses a hybrid mirror model:

1. **Journal events** record important transient-state transitions
2. **RunLedger snapshot state** reflects the latest known teammate transient state

This split is deliberate:

- events preserve the change trail
- snapshots make current state easy to read

### Event Types

The change introduces or formalizes two transition classes:

- `teammate_approval_blocked`
- `teammate_approval_unblocked`

Each event should carry only the state needed to reconstruct operator-visible teammate runtime status, such as:

- `run_id`
- `runtime_condition`
- `blocked_reason`
- `grant_request_id`
- timestamp

Event kinds should be added alongside the existing `JournalEventType` constants, and their payloads should use typed structs rather than ad hoc maps.

### Snapshot Fields

The RunLedger snapshot should retain the latest values for:

- current `runtime_condition`
- current `blocked_reason`
- current `grant_request_id`

The snapshot is not a separate source of truth. It is a durable mirror of the operator-visible teammate transient state.

This change assumes snapshot extension is handled through `RunSnapshot` plus `snapshot_data` JSON expansion, not through new Ent columns by default. If implementation discovers a need for new durable columns, that should be called out explicitly as a separate schema decision with migration requirements.

## Wiring

### Mirror Choke Point

This change should mirror state from the existing control-plane projection choke point rather than from scattered logic-level emitters.

The current production writers for teammate transient state already converge on `AgentRunStore.UpdateProjection(...)`. That is the narrowest and most reliable place to attach the mirror:

- no emitter duplication
- future projection writers automatically join the same mirror path
- unblock semantics can be derived from state delta rather than per-caller branching
- post-write reconciliation remains compatible because the mirror sees the actual final projection transition

Recommended structure:

1. keep `AgentRunStore.UpdateProjection(...)` as the control-plane write path
2. wrap or decorate that store with a RunLedger mirror adapter
3. derive journal events and snapshot updates from the incoming projection patch plus before/after state

With this structure:

- `blocked_waiting_approval` transition to active state emits `teammate_approval_blocked`
- any transition from approval-blocked state to clear state emits `teammate_approval_unblocked`
- the mirror logic is state-delta based, not caller-name based

### Mirror Failure Policy

RunLedger mirroring is **best effort** in this change.

If the mirror write fails:

- the control-plane projection write still succeeds
- the mirror failure is logged
- the mirror failure increments an observable metric counter
- drift remains observable for operators and future audit

This is the right trade-off for a mirror-only change. Runtime continuity is more important than failing closed on durability while RunLedger is still not the authoritative live read path.

When RunLedger is disabled or write-through is unavailable, mirroring is silently skipped and the live control-plane projection remains the only state source.

## Read Model Boundary

This change intentionally stops short of read-path convergence.

That means:

- `AgentRunStore` and `AgentRunProjection` remain the live status model
- RunLedger gains enough durable state to explain blocked teammate runs later
- future work may move `agent_wait`, CLI, or TUI to a stronger RunLedger-backed read model, but this change does not

This is also forward-compatible with the existing `run-ledger` authoritative-read direction. The new snapshot fields are intended to become usable by a future read-path convergence change, but that future flip is not part of this work.

## OpenSpec Impact

### Primary Spec

- `openspec/specs/run-ledger/spec.md`

### Secondary Specs

- `openspec/specs/agent-control-plane-tools/spec.md`
- `openspec/specs/multi-agent-orchestration/spec.md`

## Implementation Waves

### Wave 1: Contract Closure

- Define teammate transient state durability in `run-ledger`
- Close the archived audit gap for approval-blocked conditions
- Explicitly defer `recovery_state` durability until a production writer exists
- Record best-effort mirror semantics explicitly
- Record that archived verdict closure happens via cross-reference in the new change's design, not by mutating the archived design in place
- Record the schema decision:
  - new journal event kinds + typed payload structs are added
  - snapshot fields are extended through `RunSnapshot` / `snapshot_data`
  - new Ent columns are not assumed by default

### Wave 2: Runtime Mirror Wiring

- Mirror approval block transitions
- Mirror approval unblock transitions
- Attach mirroring at the `AgentRunStore.UpdateProjection(...)` choke point
- Keep drift observable when mirror writes fail

## Testing

### Unit Tests

- blocked transition appends the correct journal event
- grant/unblock transition appends the correct journal event
- snapshot fields reflect the latest transient state

### Integration Tests

- a built-in teammate run that becomes approval-blocked and then unblocked can be reconstructed from RunLedger

### Regression Tests

- RunLedger mirror failure does not fail-close the runtime path
- mirror failure produces observable drift logging and metrics
- older teammate runs created before this change continue to load with empty approval-blocked durability fields rather than failing snapshot reads

## Success Criteria

1. `blocked_waiting_approval`, `blocked_reason`, and `grant_request_id` are visible in the durable RunLedger snapshot for teammate runs
2. approval block/unblock transitions are present in the RunLedger journal trail
3. mirror write failure does not abort the live runtime path, but it is observable through logs and metrics
4. the archived hard-cut audit verdict for `approval-blocked conditions` is closed by cross-reference in the new change and becomes `recorded under best-effort mirror semantics`
5. the archived hard-cut audit verdict for `recovery states` remains `follow-up` until a production `RecoveryState` writer exists
