# Teammate Transient State Durability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Durably mirror built-in teammate approval-blocked state into RunLedger without changing the live read path.

**Architecture:** Add a small RunLedger-focused OpenSpec change, extend RunLedger with approval-block/unblock journal events plus snapshot fields, and attach a best-effort mirror decorator at the `AgentRunStore.UpdateProjection(...)` choke point. Keep the live control-plane model unchanged while making blocked teammate runs reconstructible from durable state after restart.

**Tech Stack:** Go, OpenSpec, Ent-backed and memory RunLedger stores, `internal/agentrt`, `internal/runledger`, `internal/eventbus`, Prometheus metrics, `go test ./...`, `go build ./...`.

---

## Source Spec

- Design spec: `internal-docs/superpowers/specs/2026-05-02-teammate-transient-state-durability-design.md`
- Archived audit source: `openspec/changes/archive/2026-05-01-dynamic-multi-agent-hard-cut/design.md`

## Scope Check

This plan intentionally covers only the approval-blocked durability gap:

- `blocked_waiting_approval`
- `blocked_reason`
- `grant_request_id`

It does **not** introduce a production `RecoveryState` writer and does **not** attempt RunLedger authoritative-read convergence.

## File Map

- Create `openspec/changes/teammate-transient-state-durability/proposal.md`: follow-up change summary and audit closure rationale.
- Create `openspec/changes/teammate-transient-state-durability/design.md`: live change design with carried-forward `recovery states` follow-up row.
- Create `openspec/changes/teammate-transient-state-durability/tasks.md`: implementation checklist.
- Create `openspec/changes/teammate-transient-state-durability/specs/run-ledger/spec.md`: primary durability delta.
- Create `openspec/changes/teammate-transient-state-durability/specs/agent-control-plane-tools/spec.md`: cross-reference delta for operator-visible blocked state durability.
- Create `openspec/changes/teammate-transient-state-durability/specs/multi-agent-orchestration/spec.md`: cross-reference delta for built-in teammate runtime durability expectations.
- Modify `internal/runledger/journal.go`: add journal event kinds and typed payload structs for approval block/unblock.
- Modify `internal/runledger/snapshot.go`: add durable snapshot fields and replay logic.
- Modify `internal/runledger/snapshot_test.go`: cover replay for approval block/unblock and older-event compatibility.
- Modify `internal/runledger/store_test.go`: cover older snapshots with empty new fields and append-hook behavior if needed.
- Create `internal/agentrt/runledger_mirror_store.go`: `AgentRunStore` decorator that mirrors approval-blocked state into RunLedger from `UpdateProjection(...)`.
- Create `internal/agentrt/runledger_mirror_store_test.go`: store-decorator tests for block, unblock, block-replace, terminal transition, and mirror failure behavior.
- Modify `internal/app/modules.go`: wrap the in-memory `AgentRunStore` with the RunLedger mirror decorator when RunLedger write-through is active.
- Modify `internal/eventbus/events.go`: add a `RunLedgerMirrorFailureEvent`.
- Modify `internal/observability/prometheus.go`: add a mirror-failure counter metric.
- Modify `internal/observability/prometheus_test.go`: verify the new metric increments.
- Modify `docs/features/multi-agent.md`: document that approval-blocked teammate runs now preserve durable blocked context in RunLedger-backed inspection.
- Modify `docs/features/run-ledger.md`: document teammate approval-blocked durability mirror and best-effort semantics.

## Commit Policy

Each task includes a suggested commit message, but do not create commits automatically while executing the plan. Let the user decide when to commit.

## Task 1: OpenSpec Change Skeleton

**Files:**
- Create: `openspec/changes/teammate-transient-state-durability/proposal.md`
- Create: `openspec/changes/teammate-transient-state-durability/design.md`
- Create: `openspec/changes/teammate-transient-state-durability/tasks.md`
- Create: `openspec/changes/teammate-transient-state-durability/specs/run-ledger/spec.md`
- Create: `openspec/changes/teammate-transient-state-durability/specs/agent-control-plane-tools/spec.md`
- Create: `openspec/changes/teammate-transient-state-durability/specs/multi-agent-orchestration/spec.md`

- [ ] **Step 1: Confirm the change name is free**

Run:

```bash
openspec list
```

Expected: no active change named `teammate-transient-state-durability`.

- [ ] **Step 2: Create the change directories**

Run:

```bash
openspec new change teammate-transient-state-durability
mkdir -p openspec/changes/teammate-transient-state-durability/specs/run-ledger
mkdir -p openspec/changes/teammate-transient-state-durability/specs/agent-control-plane-tools
mkdir -p openspec/changes/teammate-transient-state-durability/specs/multi-agent-orchestration
```

Expected: the change directory exists with the three spec subdirectories.

- [ ] **Step 3: Write `proposal.md`**

Create `openspec/changes/teammate-transient-state-durability/proposal.md` with:

```markdown
# Teammate Transient State Durability

## Why

The built-in multi-agent hard cut left one major durable visibility gap open: approval-blocked teammate state is truthful in the control-plane projection but not reconstructible from RunLedger after restart. The archived hard-cut audit marked `approval-blocked conditions` as `follow-up` for that reason.

## What Changes

This follow-up mirrors approval-blocked teammate transient state into RunLedger using journal events plus latest snapshot state. It covers `blocked_waiting_approval`, `blocked_reason`, and `grant_request_id` only. `recovery_state` remains follow-up because no production writer exists yet.

## User Impact

Operator-visible blocked teammate runs become durably inspectable later without changing the live read model. The runtime still treats RunLedger mirroring as best effort and does not fail-close if mirror writes fail.
```

- [ ] **Step 4: Write `design.md`**

Create `openspec/changes/teammate-transient-state-durability/design.md` with:

```markdown
# Design

## Archived Audit Closure

This change closes the archived `approval-blocked conditions` follow-up from `openspec/changes/archive/2026-05-01-dynamic-multi-agent-hard-cut/design.md` by cross-reference. The archived document itself remains immutable.

## Carried-Forward Follow-Ups

| Source audit item | Status in this change | Notes |
|-------------------|-----------------------|-------|
| `recovery states` | `follow-up` | No production `RecoveryState` writer exists yet, so there is no source state to mirror in this change. |
```

- [ ] **Step 5: Write `tasks.md`**

Create `openspec/changes/teammate-transient-state-durability/tasks.md` with:

```markdown
# Tasks

- [ ] Add RunLedger approval-block durability requirements
- [ ] Extend RunLedger journal and snapshot model
- [ ] Add AgentRunStore mirror decorator
- [ ] Wire the mirror into app bootstrap
- [ ] Add mirror failure metrics and tests
- [ ] Update public docs
- [ ] Run `openspec validate teammate-transient-state-durability --strict`
- [ ] Run `go build ./...`
- [ ] Run `go test ./...`
- [ ] Archive the change
```

- [ ] **Step 6: Write the primary `run-ledger` delta**

Create `openspec/changes/teammate-transient-state-durability/specs/run-ledger/spec.md` with:

```markdown
## ADDED Requirements

### Requirement: Teammate approval-blocked durability mirror
The system SHALL durably mirror built-in teammate approval-blocked state into RunLedger. The durable mirror SHALL cover `runtime_condition`, `blocked_reason`, and `grant_request_id` for approval-blocked teammate runs. This mirror uses best-effort semantics: live projection writes remain authoritative for runtime continuity, while journal plus snapshot state provide durable reconstruction.

#### Scenario: Approval-blocked teammate state is reconstructible
- **WHEN** a built-in teammate run enters `blocked_waiting_approval`
- **THEN** RunLedger SHALL append a durable approval-block journal event
- **AND** the RunLedger snapshot SHALL retain the latest blocked condition, blocked reason, and grant request ID

#### Scenario: Approval unblock clears durable blocked state
- **WHEN** a built-in teammate run leaves approval-blocked state
- **THEN** RunLedger SHALL append a durable approval-unblocked journal event
- **AND** the latest durable blocked snapshot fields SHALL be cleared

#### Scenario: Mirror failure does not fail-close runtime
- **WHEN** the durable mirror write fails
- **THEN** the live control-plane projection write SHALL still succeed
- **AND** the failure SHALL be observable through logs and metrics

#### Scenario: RunLedger disabled skips mirror silently
- **WHEN** RunLedger or write-through mirroring is disabled
- **THEN** approval-blocked mirroring SHALL be skipped
- **AND** the live control-plane projection SHALL remain the only state source
```

- [ ] **Step 7: Write the secondary cross-reference deltas**

Create `openspec/changes/teammate-transient-state-durability/specs/agent-control-plane-tools/spec.md` with:

```markdown
## ADDED Requirements

### Requirement: Durable blocked-state cross-reference
The control-plane blocked-state surface for built-in teammate runs SHALL remain aligned with the RunLedger durability mirror defined by the `run-ledger` spec.

#### Scenario: Durable mirror does not replace live projection
- **WHEN** `agent_wait` or other live control-plane readers expose approval-blocked state
- **THEN** those readers SHALL continue using the live projection path
- **AND** the RunLedger mirror SHALL serve durable reconstruction rather than replacing the live read path in this change
```

Create `openspec/changes/teammate-transient-state-durability/specs/multi-agent-orchestration/spec.md` with:

```markdown
## ADDED Requirements

### Requirement: Built-in teammate blocked-state durability
Built-in teammate approval-blocked state SHALL be durably reconstructible through the RunLedger mirror while the live runtime continues using the control-plane projection.

#### Scenario: Hard-cut audit closure is recorded by cross-reference
- **WHEN** the transient-state durability change completes
- **THEN** the archived hard-cut `approval-blocked conditions` follow-up SHALL be closed by cross-reference in the new change
- **AND** the archived `recovery states` follow-up SHALL remain open until a production writer exists
```

- [ ] **Step 8: Validate the change skeleton**

Run:

```bash
openspec validate teammate-transient-state-durability --strict
```

Expected: PASS.

- [ ] **Step 9: Suggested commit**

Suggested commit message:

```bash
docs: scaffold teammate transient state durability change
```

## Task 2: Extend RunLedger Journal and Snapshot Model

**Files:**
- Modify: `internal/runledger/journal.go`
- Modify: `internal/runledger/snapshot.go`
- Modify: `internal/runledger/snapshot_test.go`
- Modify: `internal/runledger/store_test.go`

- [ ] **Step 1: Add failing RunLedger snapshot tests**

Add these tests to `internal/runledger/snapshot_test.go`:

```go
func TestApplyEvent_TeammateApprovalBlockedUpdatesSnapshot(t *testing.T) {
	now := time.Now()
	snap := &RunSnapshot{RunID: "run-1", Notes: map[string]string{}}
	ev := JournalEvent{
		RunID:     "run-1",
		Seq:       1,
		Type:      EventTeammateApprovalBlocked,
		Timestamp: now,
		Payload: marshalPayload(TeammateApprovalBlockedPayload{
			RuntimeCondition: "blocked_waiting_approval",
			BlockedReason:    "dangerous tool requires approval",
			GrantRequestID:   "grant-run-1-exec",
		}),
	}

	require.NoError(t, applyEvent(snap, &ev))
	assert.Equal(t, "blocked_waiting_approval", snap.TeammateRuntimeCondition)
	assert.Equal(t, "dangerous tool requires approval", snap.TeammateBlockedReason)
	assert.Equal(t, "grant-run-1-exec", snap.TeammateGrantRequestID)
}

func TestApplyEvent_TeammateApprovalUnblockedClearsSnapshot(t *testing.T) {
	now := time.Now()
	snap := &RunSnapshot{
		RunID:                    "run-1",
		Notes:                    map[string]string{},
		TeammateRuntimeCondition: "blocked_waiting_approval",
		TeammateBlockedReason:    "dangerous tool requires approval",
		TeammateGrantRequestID:   "grant-run-1-exec",
	}
	ev := JournalEvent{
		RunID:     "run-1",
		Seq:       2,
		Type:      EventTeammateApprovalUnblocked,
		Timestamp: now,
		Payload: marshalPayload(TeammateApprovalUnblockedPayload{
			PreviousGrantRequestID: "grant-run-1-exec",
		}),
	}

	require.NoError(t, applyEvent(snap, &ev))
	assert.Empty(t, snap.TeammateRuntimeCondition)
	assert.Empty(t, snap.TeammateBlockedReason)
	assert.Empty(t, snap.TeammateGrantRequestID)
}
```

- [ ] **Step 2: Run the focused RunLedger tests**

Run:

```bash
go test ./internal/runledger -run 'TestApplyEvent_TeammateApprovalBlockedUpdatesSnapshot|TestApplyEvent_TeammateApprovalUnblockedClearsSnapshot' -count=1
```

Expected: FAIL because the event kinds, payload types, and snapshot fields do not exist yet.

- [ ] **Step 3: Add event kinds, payloads, and snapshot fields**

Update `internal/runledger/journal.go` with:

```go
const (
	EventTeammateApprovalBlocked   JournalEventType = "teammate_approval_blocked"
	EventTeammateApprovalUnblocked JournalEventType = "teammate_approval_unblocked"
)

type TeammateApprovalBlockedPayload struct {
	RuntimeCondition string `json:"runtime_condition"`
	BlockedReason    string `json:"blocked_reason"`
	GrantRequestID   string `json:"grant_request_id"`
}

type TeammateApprovalUnblockedPayload struct {
	PreviousGrantRequestID string `json:"previous_grant_request_id,omitempty"`
}
```

Update `internal/runledger/snapshot.go` `RunSnapshot` with:

```go
	TeammateRuntimeCondition string `json:"teammate_runtime_condition,omitempty"`
	TeammateBlockedReason    string `json:"teammate_blocked_reason,omitempty"`
	TeammateGrantRequestID   string `json:"teammate_grant_request_id,omitempty"`
```

And add these `applyEvent` cases:

```go
	case EventTeammateApprovalBlocked:
		var p TeammateApprovalBlockedPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("unmarshal teammate_approval_blocked: %w", err)
		}
		snap.TeammateRuntimeCondition = p.RuntimeCondition
		snap.TeammateBlockedReason = p.BlockedReason
		snap.TeammateGrantRequestID = p.GrantRequestID

	case EventTeammateApprovalUnblocked:
		snap.TeammateRuntimeCondition = ""
		snap.TeammateBlockedReason = ""
		snap.TeammateGrantRequestID = ""
```

- [ ] **Step 4: Add backward-compat snapshot coverage**

Add this test to `internal/runledger/store_test.go`:

```go
func TestMemoryStore_GetRunSnapshot_OlderRunsWithoutTeammateFieldsStillLoad(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	require.NoError(t, store.AppendJournalEvent(ctx, JournalEvent{
		RunID:   "run-old",
		Type:    EventRunCreated,
		Payload: marshalPayload(RunCreatedPayload{Goal: "legacy"}),
	}))

	snap, err := store.GetRunSnapshot(ctx, "run-old")
	require.NoError(t, err)
	assert.Empty(t, snap.TeammateRuntimeCondition)
	assert.Empty(t, snap.TeammateBlockedReason)
	assert.Empty(t, snap.TeammateGrantRequestID)
}
```

- [ ] **Step 5: Run the RunLedger package tests**

Run:

```bash
go test ./internal/runledger -count=1
```

Expected: PASS.

- [ ] **Step 6: Suggested commit**

Suggested commit message:

```bash
feat: extend runledger for teammate blocked durability
```

## Task 3: Add the AgentRunStore Mirror Decorator

**Files:**
- Create: `internal/agentrt/runledger_mirror_store.go`
- Create: `internal/agentrt/runledger_mirror_store_test.go`
- Modify: `internal/eventbus/events.go`
- Modify: `internal/observability/prometheus.go`
- Modify: `internal/observability/prometheus_test.go`

- [ ] **Step 1: Write the failing decorator test**

Create `internal/agentrt/runledger_mirror_store_test.go` with:

```go
type failingRunLedgerStore struct {
	runledger.RunLedgerStore
	appendErr error
}

func (s *failingRunLedgerStore) AppendJournalEvent(ctx context.Context, event runledger.JournalEvent) error {
	if s.appendErr != nil {
		return s.appendErr
	}
	return s.RunLedgerStore.AppendJournalEvent(ctx, event)
}

func TestRunLedgerMirrorStore_ApprovalBlockedAppendsJournalEvent(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:     "arun-1",
		Status: AgentRunRunning,
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-1",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-1", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "dangerous tool requires approval",
		GrantRequestID:        "grant-arun-1-exec",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-1")
	require.NoError(t, err)
	assert.Equal(t, runledger.EventTeammateApprovalBlocked, events[len(events)-1].Type)
}
```

- [ ] **Step 2: Run the focused decorator test**

Run:

```bash
go test ./internal/agentrt -run TestRunLedgerMirrorStore_ApprovalBlockedAppendsJournalEvent -count=1
```

Expected: FAIL because the decorator and helper APIs do not exist yet.

- [ ] **Step 3: Add the mirror failure event**

Add to `internal/eventbus/events.go`:

```go
const EventRunLedgerMirrorFailure = "runledger.mirror.failure"

type RunLedgerMirrorFailureEvent struct {
	Target string
	Phase  string
	RunID  string
	Error  string
}

func (e RunLedgerMirrorFailureEvent) EventName() string { return EventRunLedgerMirrorFailure }
```

- [ ] **Step 4: Add the Prometheus counter**

Extend `internal/observability/prometheus.go` with:

```go
	runLedgerMirrorFailures *prometheus.CounterVec
```

Register:

```go
		runLedgerMirrorFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "lango_runledger_mirror_failures_total",
			Help: "Total RunLedger mirror failures by target and phase.",
		}, []string{"target", "phase"}),
```

Subscribe:

```go
	eventbus.SubscribeTyped[eventbus.RunLedgerMirrorFailureEvent](bus, func(evt eventbus.RunLedgerMirrorFailureEvent) {
		e.runLedgerMirrorFailures.WithLabelValues(evt.Target, evt.Phase).Inc()
	})
```

- [ ] **Step 5: Implement the decorator**

Create `internal/agentrt/runledger_mirror_store.go` with this shape:

```go
type RunLedgerMirrorStore struct {
	base   AgentRunStore
	ledger runledger.RunLedgerStore
	bus    *eventbus.Bus
}

func NewRunLedgerMirrorStore(base AgentRunStore, ledger runledger.RunLedgerStore, bus *eventbus.Bus) *RunLedgerMirrorStore {
	return &RunLedgerMirrorStore{base: base, ledger: ledger, bus: bus}
}
```

Implement all `AgentRunStore` methods by delegating to `base`, except `UpdateProjection`.

Inside `UpdateProjection`:

1. `before, err := s.base.Get(id)`
2. `err = s.base.UpdateProjection(id, patch)`
3. `after, err := s.base.Get(id)`
4. Derive transitions:
   - `before.RuntimeCondition != blocked_waiting_approval && after.RuntimeCondition == blocked_waiting_approval` => append `EventTeammateApprovalBlocked`
   - `before.RuntimeCondition == blocked_waiting_approval && after.RuntimeCondition == ""` => append `EventTeammateApprovalUnblocked`
5. After each successful append, call `s.ledger.GetRunSnapshot(ctx, id)` best effort to refresh the cached snapshot
6. On mirror failure:
   - `logger().Warnw(...)`
   - `bus.Publish(eventbus.RunLedgerMirrorFailureEvent{...})` when `bus != nil`
   - do **not** return the mirror failure

Use `context.Background()` inside the decorator for the mirror write path.

- [ ] **Step 6: Add mirror failure and unblock tests**

Extend `internal/agentrt/runledger_mirror_store_test.go` with:

```go
func TestRunLedgerMirrorStore_ApprovalUnblockedAppendsJournalEvent(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:               "arun-2",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "dangerous tool requires approval",
		GrantRequestID:   "grant-arun-2-exec",
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-2",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-2", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionNone,
		BlockedReason:         "",
		GrantRequestID:        "",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-2")
	require.NoError(t, err)
	assert.Equal(t, runledger.EventTeammateApprovalUnblocked, events[len(events)-1].Type)
}

func TestRunLedgerMirrorStore_MirrorFailureDoesNotFailProjection(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := &failingRunLedgerStore{RunLedgerStore: runledger.NewMemoryStore(), appendErr: errors.New("boom")}

	require.NoError(t, base.Create(&AgentRun{
		ID:     "arun-3",
		Status: AgentRunRunning,
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	err := store.UpdateProjection("arun-3", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "dangerous tool requires approval",
		GrantRequestID:        "grant-arun-3-exec",
	})
	require.NoError(t, err)

	run, err := base.Get("arun-3")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, run.RuntimeCondition)

	_, getErr := ledger.GetJournalEvents(ctx, "arun-3")
	require.Error(t, getErr)
}
```

- [ ] **Step 7: Add the Prometheus test**

Extend `internal/observability/prometheus_test.go` with:

```go
func TestPrometheusExporter_RunLedgerMirrorFailureCounter(t *testing.T) {
	bus := eventbus.New()
	exp := NewPrometheusExporter()
	exp.Subscribe(bus)

	bus.Publish(eventbus.RunLedgerMirrorFailureEvent{
		Target: "agent_run_projection",
		Phase:  "append_journal",
		RunID:  "arun-1",
		Error:  "boom",
	})

	ts := httptest.NewServer(exp.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	text := string(body)
	assert.Contains(t, text, "lango_runledger_mirror_failures_total")
	assert.Contains(t, text, `target="agent_run_projection"`)
	assert.Contains(t, text, `phase="append_journal"`)
}
```

- [ ] **Step 8: Run package tests**

Run:

```bash
go test ./internal/agentrt -count=1
go test ./internal/observability -count=1
```

Expected: PASS.

- [ ] **Step 9: Suggested commit**

Suggested commit message:

```bash
feat: mirror teammate blocked state into runledger
```

## Task 4: Wire the Mirror into App Bootstrap

**Files:**
- Modify: `internal/app/modules.go`
- Modify: `internal/app/modules_test.go`

- [ ] **Step 1: Add a failing wiring test**

First extract a small helper in `internal/app/modules.go`:

```go
func newAutomationAgentRunStore(
	cfg *config.Config,
	rlv *runLedgerValues,
	bus *eventbus.Bus,
) agentrt.AgentRunStore {
	base := agentrt.NewInMemoryAgentRunStore()
	if rlv != nil && rlv.store != nil && cfg.RunLedger.Enabled && cfg.RunLedger.WriteThrough {
		return agentrt.NewRunLedgerMirrorStore(base, rlv.store, bus)
	}
	return base
}
```

Then add this test to `internal/app/modules_test.go`:

```go
func TestAutomationModule_WrapsAgentRunStoreWithRunLedgerMirrorWhenWriteThroughEnabled(t *testing.T) {
	client := testutil.TestEntClient(t)
	cfg := config.DefaultConfig()
	cfg.RunLedger.Enabled = true
	cfg.RunLedger.WriteThrough = true
	cfg.Background.Enabled = true

	boot := &bootstrap.Result{
		Storage: storage.NewFacade(nil, nil, storage.WithEntClient(client)),
	}

	runLedgerVals := &runLedgerValues{store: boot.Storage.RunLedger()}
	store := newAutomationAgentRunStore(cfg, runLedgerVals, nil)
	_, ok := store.(*agentrt.RunLedgerMirrorStore)
	require.True(t, ok)
}
```

- [ ] **Step 2: Run the focused app test**

Run:

```bash
go test ./internal/app -run TestAutomationModule_WrapsAgentRunStoreWithRunLedgerMirrorWhenWriteThroughEnabled -count=1
```

Expected: FAIL before the helper and mirror wiring exist.

- [ ] **Step 3: Compose the mirror store through the helper**

Change the automation-module wiring in `internal/app/modules.go` from:

```go
agentRunStore := agentrt.NewInMemoryAgentRunStore()
agentRunProjection := agentrt.NewAgentRunProjection(agentRunStore)
capabilityRuntime := agentrt.NewCapabilityRuntime(agentRunStore, ...)
```

to this structure:

```go
agentRunStore := newAutomationAgentRunStore(cfg, rlv, m.bus)
agentRunProjection := agentrt.NewAgentRunProjection(agentRunStore)
capabilityRuntime := agentrt.NewCapabilityRuntime(agentRunStore, ...)
```

Leave the existing background projection wiring in place.

- [ ] **Step 4: Run the app package tests**

Run:

```bash
go test ./internal/app -count=1
```

Expected: PASS.

- [ ] **Step 5: Suggested commit**

Suggested commit message:

```bash
refactor: wire runledger mirror store for teammate runtime
```

## Task 5: Update Public Docs and Close the Change

**Files:**
- Modify: `docs/features/multi-agent.md`
- Modify: `docs/features/run-ledger.md`
- Modify: `openspec/changes/teammate-transient-state-durability/tasks.md`

- [ ] **Step 1: Update `docs/features/multi-agent.md`**

Add a short note near teammate blocked-state behavior:

```markdown
When a built-in teammate enters `blocked_waiting_approval`, the live control-plane projection remains the primary read path, but the blocked condition, reason, and grant request ID are also mirrored durably into RunLedger for later inspection.
```

- [ ] **Step 2: Update `docs/features/run-ledger.md`**

Add a short section:

```markdown
## Teammate Approval-Blocked Durability

RunLedger now mirrors built-in teammate approval-blocked state using journal events plus latest snapshot fields. This mirror is best effort: live runtime writes do not fail closed if the durable mirror write fails, but failures are logged and counted.
```

- [ ] **Step 3: Mark the OpenSpec task checklist complete**

Update `openspec/changes/teammate-transient-state-durability/tasks.md` so all completed items are checked.

- [ ] **Step 4: Run final verification**

Run:

```bash
openspec validate teammate-transient-state-durability --strict
go build ./...
go test ./...
```

Expected: all commands PASS.

- [ ] **Step 5: Archive the change**

Run:

```bash
openspec archive -y teammate-transient-state-durability
```

Expected: the change moves under `openspec/changes/archive/` and main specs update.

- [ ] **Step 6: Suggested commit**

Suggested commit message:

```bash
feat: add durable mirror for teammate blocked state
```

## Self-Review

- Spec coverage:
  - approval-blocked durable mirror: Tasks 1-4
  - best-effort semantics: Tasks 3 and 5
  - archived audit cross-reference closure: Tasks 1 and 5
  - recovery-state deferral remains explicit: Tasks 1 and 5
- Placeholder scan:
  - no `TODO`, `TBD`, or empty “implement later” markers remain
  - all tasks include exact files, code shapes, commands, and expected results
- Type consistency:
  - event names use `EventTeammateApprovalBlocked` / `EventTeammateApprovalUnblocked`
  - snapshot fields use `TeammateRuntimeCondition`, `TeammateBlockedReason`, `TeammateGrantRequestID`
  - decorator type is `RunLedgerMirrorStore`
