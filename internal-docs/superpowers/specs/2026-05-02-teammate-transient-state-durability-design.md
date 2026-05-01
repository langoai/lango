# Teammate Transient State Durability Design

## Goal

Close the durable visibility gap left by the built-in multi-agent hard cut by mirroring operator-visible teammate transient state into RunLedger.

This change covers:

- `blocked_waiting_approval`
- `blocked_reason`
- `grant_request_id`
- `recovery_state`

The goal is **mirror only**. This change does not turn RunLedger into the authoritative read path for teammate runtime status. It makes the current operator-visible transient state reconstructible from durable state after restart or replay.

## Scope

### In Scope

- Durable mirror for approval-blocked teammate state
- Durable mirror for teammate recovery state
- Journal events for state transitions
- Snapshot fields for latest transient state
- Best-effort mirror semantics with explicit drift logging

### Out of Scope

- Replacing `AgentRunStore` or `AgentRunProjection` as the primary live read model
- Rewriting `agent_wait` or CLI/TUI to read from RunLedger first
- Approval UI redesign
- Broader recovery-engine redesign

## Problem

The hard-cut audit closed most built-in teammate runtime gaps, but two operator-visible state families remained marked as follow-up in the archived change:

- approval-blocked conditions
- recovery states

Today these are visible through the control-plane projection, but they are not durably mirrored in a way that allows a teammate run's blocked or recovering state to be reconstructed from RunLedger after restart.

That leaves an operator-visible truth gap:

- live surfaces can say why a teammate is blocked or recovering
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

The change introduces or formalizes three transition classes:

- `teammate_approval_blocked`
- `teammate_approval_unblocked`
- `teammate_recovery_state_changed`

Each event should carry only the state needed to reconstruct operator-visible teammate runtime status, such as:

- `run_id`
- `runtime_condition`
- `blocked_reason`
- `grant_request_id`
- `recovery_state`
- timestamp

### Snapshot Fields

The RunLedger snapshot should retain the latest values for:

- current `runtime_condition`
- current `blocked_reason`
- current `grant_request_id`
- current `recovery_state`

The snapshot is not a separate source of truth. It is a durable mirror of the operator-visible teammate transient state.

## Wiring

### Emitters

This change should mirror state from the places that already mutate control-plane teammate projection state rather than inventing a parallel state machine.

Primary emitters:

1. `CapabilityRuntime.HandleBlockedToolCall()`
   - when `blocked_waiting_approval` is projected
   - append `teammate_approval_blocked`
   - update latest durable blocked snapshot fields

2. `CapabilityRuntime.ApplyGrant()`
   - when approval-blocked state is cleared
   - append `teammate_approval_unblocked`
   - clear the durable blocked snapshot fields

3. teammate recovery-state writer
   - when `recovery_state` changes
   - append `teammate_recovery_state_changed`
   - update latest durable recovery snapshot field

### Mirror Failure Policy

RunLedger mirroring is **best effort** in this change.

If the mirror write fails:

- the control-plane projection write still succeeds
- the mirror failure is logged
- drift remains observable for operators and future audit

This is the right trade-off for a mirror-only change. Runtime continuity is more important than failing closed on durability while RunLedger is still not the authoritative live read path.

## Read Model Boundary

This change intentionally stops short of read-path convergence.

That means:

- `AgentRunStore` and `AgentRunProjection` remain the live status model
- RunLedger gains enough durable state to explain blocked/recovering teammate runs later
- future work may move `agent_wait`, CLI, or TUI to a stronger RunLedger-backed read model, but this change does not

## OpenSpec Impact

### Primary Spec

- `openspec/specs/run-ledger/spec.md`

### Secondary Specs

- `openspec/specs/agent-control-plane-tools/spec.md`
- `openspec/specs/multi-agent-orchestration/spec.md`

## Implementation Waves

### Wave 1: Contract Closure

- Define teammate transient state durability in `run-ledger`
- Close the archived audit gap for:
  - approval-blocked conditions
  - recovery states
- Record best-effort mirror semantics explicitly

### Wave 2: Runtime Mirror Wiring

- Mirror approval block transitions
- Mirror approval unblock transitions
- Mirror recovery-state transitions
- Keep drift observable when mirror writes fail

## Testing

### Unit Tests

- blocked transition appends the correct journal event
- grant/unblock transition appends the correct journal event
- recovery-state transition appends the correct journal event
- snapshot fields reflect the latest transient state

### Integration Tests

- a built-in teammate run that becomes approval-blocked and then unblocked can be reconstructed from RunLedger
- a built-in teammate run that changes recovery state can be reconstructed from RunLedger

### Regression Tests

- RunLedger mirror failure does not fail-close the runtime path
- mirror failure produces observable drift logging

## Success Criteria

1. `blocked_waiting_approval`, `blocked_reason`, and `grant_request_id` are visible in the durable RunLedger snapshot for teammate runs
2. approval block/unblock transitions are present in the RunLedger journal trail
3. `recovery_state` is durably mirrored for teammate runs
4. mirror write failure does not abort the live runtime path, but it is observable
5. the archived hard-cut audit verdicts for:
   - approval-blocked conditions
   - recovery states
   change from `follow-up` to `recorded`
