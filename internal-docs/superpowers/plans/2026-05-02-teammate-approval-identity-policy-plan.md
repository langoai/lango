# Teammate Approval Identity Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make built-in teammate approval identity explicit by keeping `GrantRequestID` stable per `(run, tool)` while adding minimal attempt metadata that lets runtime, `agent_wait`, and RunLedger distinguish renewed approval attempts.

**Architecture:** Keep the current deterministic `GrantRequestID` shape and runtime matching behavior, then layer attempt metadata on top of it instead of rotating the logical request ID. Update the capability layer, projected `AgentRun` state, control-plane response surface, and RunLedger durable mirror together so all layers interpret repeated blocked requests consistently.

**Tech Stack:** Go, OpenSpec, `internal/agentrt`, `internal/runledger`, Cobra CLI/control-plane tool surfaces, `go test ./...`, `go build ./...`.

---

## Source Design

- OpenSpec change: `openspec/changes/teammate-approval-identity-policy/`
- Internal design: `internal-docs/superpowers/specs/2026-05-02-teammate-approval-identity-policy-design.md`

## Scope Check

This plan intentionally implements the smallest coherent slice of the approval identity policy:

- `GrantRequestID` remains stable per `(runID, toolName)`
- repeated approval blocks for the same logical request increment attempt metadata while the request remains actively blocked
- `agent_wait` exposes that metadata
- RunLedger durable mirror preserves it

This plan does **not** redesign approval UI, introduce denial persistence beyond projected state, or add a separate approval-history store.

For this slice, attempt counting follows an **active blocked-cycle** model:

- first active block for `(run, tool)` => `grant_attempt = 1`
- renewed block while that logical request remains active => increment attempt
- grant or denial clears the active-cycle attempt metadata
- a later fresh blocked cycle for the same `(run, tool)` starts again at `grant_attempt = 1`

## Identity Inventory

Current `GrantRequestID` touch points that this plan must keep consistent:

| File | Current responsibility |
|------|------------------------|
| `internal/agentrt/agent_run.go` | projected blocked-state field on `AgentRun` |
| `internal/agentrt/agent_run_store.go` | projection patch fields + projection apply logic |
| `internal/agentrt/capability_policy.go` | approval decision shape + `GrantRequestID` generation |
| `internal/agentrt/capability_runtime.go` | blocked write, rollback, deny clear, grant clear |
| `internal/agentrt/control_tools.go` | `agent_wait` response surface |
| `internal/runledger/journal.go` | approval-blocked durable payload shape |
| `internal/agentrt/runledger_mirror_store.go` | durable mirror predicate + approval event payload |

## File Map

- Modify `openspec/changes/teammate-approval-identity-policy/tasks.md`: mark implementation checklist progress.
- Modify `openspec/changes/teammate-approval-identity-policy/specs/tool-capability-layer/spec.md`: add concrete attempt-metadata requirements.
- Modify `openspec/changes/teammate-approval-identity-policy/specs/agent-control-plane-tools/spec.md`: describe `agent_wait` exposure of attempt metadata.
- Modify `openspec/changes/teammate-approval-identity-policy/specs/run-ledger/spec.md`: describe durable mirror of attempt metadata.
- Modify `internal/agentrt/agent_run.go`: add attempt metadata fields to projected `AgentRun`.
- Modify `internal/agentrt/agent_run_store.go`: add projection patch support for attempt metadata.
- Modify `internal/agentrt/agent_run_store_test.go`: cover storing/clearing attempt metadata.
- Modify `internal/agentrt/capability_policy.go`: keep stable `GrantRequestID` generation explicit and add a helper for logical request IDs.
- Modify `internal/agentrt/capability_policy_test.go`: lock stable identity behavior.
- Modify `internal/agentrt/capability_runtime.go`: increment attempt metadata when the same logical blocked request is re-issued.
- Modify `internal/agentrt/capability_runtime_test.go`: cover first attempt, renewed attempt, grant clear, and stable-ID semantics.
- Modify `internal/agentrt/control_tools.go`: expose attempt metadata through `agent_wait`.
- Modify `internal/agentrt/control_tools_test.go`: cover `grant_attempt` and `grant_state` in wait responses.
- Modify `internal/runledger/journal.go`: add optional attempt metadata to teammate approval payloads.
- Modify `internal/runledger/snapshot.go`: materialize latest attempt metadata.
- Modify `internal/runledger/snapshot_test.go`: replay tests for attempt metadata updates and clears.
- Modify `internal/runledger/store_test.go`: legacy snapshot compatibility for missing attempt metadata fields.
- Modify `internal/agentrt/runledger_mirror_store.go`: populate attempt metadata in approval-blocked / approval-unblocked journal events.
- Modify `internal/agentrt/runledger_mirror_store_test.go`: cover stable request ID + changing attempt metadata.
- Modify `docs/features/multi-agent.md`: describe stable approval identity plus attempt metadata in live teammate status.
- Modify `docs/features/run-ledger.md`: describe durable mirror of `grant_attempt` / `grant_state`.

## Commit Policy

Each task ends with a suggested commit message, but do not create commits automatically while executing the plan. Let the user decide when to commit.

## Task 1: Tighten the OpenSpec Contract

**Files:**
- Modify: `openspec/changes/teammate-approval-identity-policy/specs/tool-capability-layer/spec.md`
- Modify: `openspec/changes/teammate-approval-identity-policy/specs/agent-control-plane-tools/spec.md`
- Modify: `openspec/changes/teammate-approval-identity-policy/specs/run-ledger/spec.md`
- Modify: `openspec/changes/teammate-approval-identity-policy/tasks.md`

- [ ] **Step 1: Add the stable-ID + attempt-metadata requirement**

Update `openspec/changes/teammate-approval-identity-policy/specs/tool-capability-layer/spec.md` so the requirement reads:

```markdown
## ADDED Requirements

### Requirement: Teammate approval request identity policy
The teammate capability layer SHALL define approval identity using a stable logical `GrantRequestID` per `(runID, toolName)`. Repeated approval blocks for the same logical request SHALL reuse that `GrantRequestID` and SHALL surface renewed-attempt semantics through separate attempt metadata.

#### Scenario: First approval block initializes stable identity
- **WHEN** a built-in teammate first blocks on a dangerous in-scope tool
- **THEN** the runtime SHALL assign a stable logical `GrantRequestID`
- **AND** the runtime SHALL initialize `grant_attempt = 1`
- **AND** the runtime SHALL expose a pending-style attempt state

#### Scenario: Repeated block for the same run and tool reuses the logical identity
- **WHEN** the same built-in teammate run later blocks again on the same tool
- **THEN** the runtime SHALL reuse the same logical `GrantRequestID`
- **AND** it SHALL increment separate attempt metadata instead of rotating the logical request ID

#### Scenario: Grant or denial ends the active attempt cycle
- **WHEN** a blocked approval request is granted or denied
- **THEN** the runtime SHALL clear the active blocked projection state
- **AND** a later fresh blocked cycle for the same `(run, tool)` MAY reuse the same logical `GrantRequestID`
- **AND** that fresh cycle SHALL restart `grant_attempt` at `1`
```

- [ ] **Step 2: Add the control-plane response requirement**

Update `openspec/changes/teammate-approval-identity-policy/specs/agent-control-plane-tools/spec.md` to include:

```markdown
## ADDED Requirements

### Requirement: Approval identity is exposed consistently
The control-plane blocked-state surface for built-in teammate runs SHALL expose stable logical approval identity together with attempt metadata.

#### Scenario: agent_wait exposes logical identity and attempt metadata
- **WHEN** `agent_wait` reports an approval-blocked teammate run
- **THEN** the response SHALL include `grant_request_id`
- **AND** the response SHALL expose attempt metadata sufficient to distinguish the first attempt from a renewed attempt of the same logical request
- **AND** `grant_attempt` SHALL be at least `1` whenever the run is currently `blocked_waiting_approval`
```

- [ ] **Step 3: Add the durable mirror requirement**

Update `openspec/changes/teammate-approval-identity-policy/specs/run-ledger/spec.md` to include:

```markdown
## ADDED Requirements

### Requirement: Durable mirror preserves approval identity semantics
The RunLedger durable mirror for built-in teammate approval blocking SHALL preserve both the stable logical `grant_request_id` and the latest attempt metadata for that logical request.

#### Scenario: Durable snapshot reflects renewed attempt without rotating request ID
- **WHEN** a built-in teammate approval-blocked request is re-issued for the same run and tool
- **THEN** the durable mirror SHALL preserve the same logical `grant_request_id`
- **AND** the latest durable snapshot SHALL reflect the new attempt metadata
```

- [ ] **Step 4: Update the OpenSpec task checklist**

Replace `openspec/changes/teammate-approval-identity-policy/tasks.md` with:

```markdown
# Tasks

- [ ] Add stable approval identity requirements to OpenSpec
- [ ] Extend AgentRun projection with approval attempt metadata
- [ ] Keep GrantRequestID stable while incrementing renewed attempts
- [ ] Expose approval attempt metadata through `agent_wait`
- [ ] Mirror approval attempt metadata into RunLedger
- [ ] Update docs
- [ ] Run `openspec validate teammate-approval-identity-policy --strict`
- [ ] Run `go build ./...`
- [ ] Run `go test ./...`
- [ ] Archive the change
```

- [ ] **Step 5: Validate the spec change**

Run:

```bash
openspec validate teammate-approval-identity-policy --strict
```

Expected: PASS.

- [ ] **Step 6: Suggested commit**

Suggested commit message:

```bash
docs: tighten teammate approval identity contract
```

## Task 2: Add Attempt Metadata to Projected AgentRun State

**Files:**
- Modify: `internal/agentrt/agent_run.go`
- Modify: `internal/agentrt/agent_run_store.go`
- Modify: `internal/agentrt/agent_run_store_test.go`

- [ ] **Step 1: Write failing store tests for attempt metadata**

Add these tests to `internal/agentrt/agent_run_store_test.go`:

```go
func TestInMemoryAgentRunStore_UpdateProjection_StoresGrantAttemptMetadata(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "run-1",
		Status: AgentRunRunning,
	}))

	err := store.UpdateProjection("run-1", RunProjectionPatch{
		ApplyGrantAttempt: true,
		ApplyGrantState:   true,
		GrantAttempt:      2,
		GrantState:        "pending",
	})
	require.NoError(t, err)

	run, err := store.Get("run-1")
	require.NoError(t, err)
	assert.Equal(t, 2, run.GrantAttempt)
	assert.Equal(t, "pending", run.GrantState)
}

func TestInMemoryAgentRunStore_UpdateProjection_ClearsGrantAttemptMetadata(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:           "run-2",
		Status:       AgentRunRunning,
		GrantAttempt: 3,
		GrantState:   "pending",
	}))

	err := store.UpdateProjection("run-2", RunProjectionPatch{
		ApplyGrantAttempt: true,
		ApplyGrantState:   true,
		GrantAttempt:      0,
		GrantState:        "",
	})
	require.NoError(t, err)

	run, err := store.Get("run-2")
	require.NoError(t, err)
	assert.Equal(t, 0, run.GrantAttempt)
	assert.Empty(t, run.GrantState)
}
```

- [ ] **Step 2: Run the focused store tests**

Run:

```bash
go test ./internal/agentrt -run 'TestInMemoryAgentRunStore_UpdateProjection_StoresGrantAttemptMetadata|TestInMemoryAgentRunStore_UpdateProjection_ClearsGrantAttemptMetadata' -count=1
```

Expected: FAIL because the fields and patch flags do not exist yet.

- [ ] **Step 3: Extend AgentRun and RunProjectionPatch**

Update `internal/agentrt/agent_run.go`:

```go
	GrantAttempt     int
	GrantState       string
```

Update `internal/agentrt/agent_run_store.go`:

```go
	ApplyGrantAttempt      bool
	ApplyGrantState        bool
	GrantAttempt           int
	GrantState             string
```

And in `UpdateProjection(...)` add:

```go
	if patch.ApplyGrantAttempt {
		run.GrantAttempt = patch.GrantAttempt
	}
	if patch.ApplyGrantState {
		run.GrantState = patch.GrantState
	}
```

- [ ] **Step 4: Run the focused store tests again**

Run:

```bash
go test ./internal/agentrt -run 'TestInMemoryAgentRunStore_UpdateProjection_StoresGrantAttemptMetadata|TestInMemoryAgentRunStore_UpdateProjection_ClearsGrantAttemptMetadata' -count=1
```

Expected: PASS.

- [ ] **Step 5: Suggested commit**

Suggested commit message:

```bash
feat: extend agent run projection with grant attempt metadata
```

## Task 3: Keep GrantRequestID Stable While Tracking Renewed Attempts

**Files:**
- Modify: `internal/agentrt/capability_policy.go`
- Modify: `internal/agentrt/capability_policy_test.go`
- Modify: `internal/agentrt/capability_runtime.go`
- Modify: `internal/agentrt/capability_runtime_test.go`

- [ ] **Step 1: Write failing policy/runtime tests**

Add this helper expectation to `internal/agentrt/capability_policy_test.go`:

```go
func TestCapabilityPolicy_DangerousInScopeToolReusesStableGrantRequestID(t *testing.T) {
	policy := CapabilityPolicy{}

	first := policy.Evaluate(CapabilityRequest{
		RunID:        "run-1",
		TeammateType: "operator",
		ToolName:     "exec",
		ToolSafety:   agent.SafetyLevelDangerous,
	})
	second := policy.Evaluate(CapabilityRequest{
		RunID:        "run-1",
		TeammateType: "operator",
		ToolName:     "exec",
		ToolSafety:   agent.SafetyLevelDangerous,
	})

	assert.Equal(t, "grant-run-1-exec", first.GrantRequestID)
	assert.Equal(t, first.GrantRequestID, second.GrantRequestID)
}
```

Add this runtime test to `internal/agentrt/capability_runtime_test.go`:

```go
func TestCapabilityRuntime_RepeatedBlockedRequestIncrementsAttemptWithoutRotatingGrantRequestID(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-repeat",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(string) agent.SafetyLevel {
		return agent.SafetyLevelDangerous
	})

	call := toolchain.BlockedToolCall{
		ToolName:    "exec",
		BlockReason: dynamicAllowedToolsBlockReason,
	}

	require.NoError(t, rt.HandleBlockedToolCall("arun-repeat", call))
	first, err := store.Get("arun-repeat")
	require.NoError(t, err)

	require.NoError(t, rt.HandleBlockedToolCall("arun-repeat", call))
	second, err := store.Get("arun-repeat")
	require.NoError(t, err)

	assert.Equal(t, first.GrantRequestID, second.GrantRequestID)
	assert.Equal(t, 1, first.GrantAttempt)
	assert.Equal(t, 2, second.GrantAttempt)
	assert.Equal(t, "pending", second.GrantState)
}
```

- [ ] **Step 2: Run the focused tests**

Run:

```bash
go test ./internal/agentrt -run 'TestCapabilityPolicy_DangerousInScopeToolReusesStableGrantRequestID|TestCapabilityRuntime_RepeatedBlockedRequestIncrementsAttemptWithoutRotatingGrantRequestID' -count=1
```

Expected: FAIL because attempt metadata is not set yet.

- [ ] **Step 3: Make stable request identity explicit in the policy**

In `internal/agentrt/capability_policy.go`, add:

```go
func grantRequestID(runID, toolName string) string {
	return fmt.Sprintf("grant-%s-%s", runID, toolName)
}
```

And replace:

```go
GrantRequestID: fmt.Sprintf("grant-%s-%s", req.RunID, req.ToolName),
```

with:

```go
GrantRequestID: grantRequestID(req.RunID, req.ToolName),
```

- [ ] **Step 4: Increment attempt metadata in the runtime**

In `internal/agentrt/capability_runtime.go`, inside the `CapabilityDecisionNeedsApproval` path:

1. Read the latest run first
2. Compare `latest.GrantRequestID` with `decision.GrantRequestID`
3. Set:

```go
attempt := 1
if latest.GrantRequestID == decision.GrantRequestID && latest.GrantAttempt > 0 {
	attempt = latest.GrantAttempt + 1
}
```

4. Include in the blocked projection patch:

```go
ApplyGrantAttempt: true,
ApplyGrantState:   true,
GrantAttempt:      attempt,
GrantState:        "pending",
```

5. The same `CapabilityDecisionNeedsApproval` branch must treat attempt/state changes as approval-block replacements for mirror purposes. Anywhere that derives `changedBlockedState` must include:

```go
changedBlockedState := before.BlockedReason != after.BlockedReason ||
	before.GrantRequestID != after.GrantRequestID ||
	before.GrantAttempt != after.GrantAttempt ||
	before.GrantState != after.GrantState
```

6. In the post-write rollback path and deny path, clear or finalize the new fields together with the old blocked fields. Both patches must include:

```go
ApplyGrantAttempt: true,
GrantAttempt:      0,
ApplyGrantState:   true,
```

For the post-write rollback path:

```go
GrantState: "",
```

For the deny path:

```go
GrantState: "denied",
```

7. In `ApplyGrant(...)`, clear the transient blocked projection and set:

```go
ApplyGrantAttempt: true,
GrantAttempt:      0,
ApplyGrantState:   true,
GrantState:        "granted",
```

This plan intentionally uses active blocked-cycle semantics rather than cross-cycle monotonic attempt counting. Do **not** rotate `GrantRequestID`.

- [ ] **Step 5: Run the focused tests again**

Run:

```bash
go test ./internal/agentrt -run 'TestCapabilityPolicy_DangerousInScopeToolReusesStableGrantRequestID|TestCapabilityRuntime_RepeatedBlockedRequestIncrementsAttemptWithoutRotatingGrantRequestID' -count=1
```

Expected: PASS.

- [ ] **Step 6: Run the full agentrt package**

Run:

```bash
go test ./internal/agentrt -count=1
```

Expected: PASS.

- [ ] **Step 7: Suggested commit**

Suggested commit message:

```bash
feat: keep stable grant request ids across renewed attempts
```

## Task 4: Expose Attempt Metadata Through agent_wait

**Files:**
- Modify: `internal/agentrt/control_tools.go`
- Modify: `internal/agentrt/control_tools_test.go`

- [ ] **Step 1: Write failing agent_wait tests**

Add this test to `internal/agentrt/control_tools_test.go`:

```go
func TestAgentWait_TimeoutIncludesGrantAttemptMetadata(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "wait-attempt",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "dangerous tool requires approval",
		GrantRequestID:   "grant-wait-attempt-exec",
		GrantAttempt:     2,
		GrantState:       "pending",
	}))

	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	waitTool := findControlTool(t, BuildControlTools(cp), "agent_wait")

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-attempt",
		"timeout":  0,
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "grant-wait-attempt-exec", m["grant_request_id"])
	assert.Equal(t, 2, m["grant_attempt"])
	assert.Equal(t, "pending", m["grant_state"])
}
```

- [ ] **Step 2: Run the focused control tool test**

Run:

```bash
go test ./internal/agentrt -run TestAgentWait_TimeoutIncludesGrantAttemptMetadata -count=1
```

Expected: FAIL because `agentRunResponse(...)` does not expose the new fields yet.

- [ ] **Step 3: Extend `agentRunResponse(...)`**

In `internal/agentrt/control_tools.go`, inside the non-terminal branch, add:

```go
		if run.GrantAttempt > 0 {
			resp["grant_attempt"] = run.GrantAttempt
		}
		if run.GrantState != "" {
			resp["grant_state"] = run.GrantState
		}
```

- [ ] **Step 4: Run the focused control tool test again**

Run:

```bash
go test ./internal/agentrt -run TestAgentWait_TimeoutIncludesGrantAttemptMetadata -count=1
```

Expected: PASS.

- [ ] **Step 5: Suggested commit**

Suggested commit message:

```bash
feat: expose teammate approval attempt metadata in agent wait
```

## Task 5: Mirror Attempt Metadata Into RunLedger

**Files:**
- Modify: `internal/runledger/journal.go`
- Modify: `internal/runledger/snapshot.go`
- Modify: `internal/runledger/snapshot_test.go`
- Modify: `internal/runledger/store_test.go`
- Modify: `internal/agentrt/runledger_mirror_store.go`
- Modify: `internal/agentrt/runledger_mirror_store_test.go`

- [ ] **Step 1: Write failing RunLedger replay tests**

Add this test to `internal/runledger/snapshot_test.go`:

```go
func TestApplyEvent_TeammateApprovalBlockedUpdatesAttemptMetadata(t *testing.T) {
	snap := &RunSnapshot{RunID: "run-1", Notes: map[string]string{}}
	ev := JournalEvent{
		RunID: "run-1",
		Seq:   1,
		Type:  EventTeammateApprovalBlocked,
		Payload: marshalPayload(TeammateApprovalBlockedPayload{
			RuntimeCondition: "blocked_waiting_approval",
			BlockedReason:    "dangerous tool requires approval",
			GrantRequestID:   "grant-run-1-exec",
			GrantAttempt:     2,
			GrantState:       "pending",
		}),
	}

	require.NoError(t, applyEvent(snap, &ev))
assert.Equal(t, 2, snap.TeammateGrantAttempt)
assert.Equal(t, "pending", snap.TeammateGrantState)
}
```

- [ ] **Step 2: Run the focused RunLedger test**

Run:

```bash
go test ./internal/runledger -run TestApplyEvent_TeammateApprovalBlockedUpdatesAttemptMetadata -count=1
```

Expected: FAIL because the payload and snapshot fields do not exist yet.

- [ ] **Step 3: Extend payloads and snapshots**

In `internal/runledger/journal.go`:

```go
type TeammateApprovalBlockedPayload struct {
	RuntimeCondition string `json:"runtime_condition"`
	BlockedReason    string `json:"blocked_reason,omitempty"`
	GrantRequestID   string `json:"grant_request_id,omitempty"`
	GrantAttempt     int    `json:"grant_attempt,omitempty"`
	GrantState       string `json:"grant_state,omitempty"`
}
```

In `internal/runledger/snapshot.go` add:

```go
	TeammateGrantAttempt   int    `json:"teammate_grant_attempt,omitempty"`
	TeammateGrantState     string `json:"teammate_grant_state,omitempty"`
```

Update replay to set and clear these fields alongside the existing teammate blocked fields.

Implement the clear helper explicitly as:

```go
func clearTeammateApprovalState(snap *RunSnapshot) {
	snap.TeammateRuntimeCondition = ""
	snap.TeammateBlockedReason = ""
	snap.TeammateGrantRequestID = ""
	snap.TeammateGrantAttempt = 0
	snap.TeammateGrantState = ""
}
```

- [ ] **Step 4: Populate mirror payloads from the decorator**

In `internal/agentrt/runledger_mirror_store.go`, when appending blocked/unblocked events, include:

```go
GrantAttempt: after.GrantAttempt,
GrantState:   after.GrantState,
```

And ensure the replacement predicate itself treats attempt-only updates as a fresh approval-block event:

```go
changedBlockedState := before.BlockedReason != after.BlockedReason ||
	before.GrantRequestID != after.GrantRequestID ||
	before.GrantAttempt != after.GrantAttempt ||
	before.GrantState != after.GrantState
```

And ensure unblocked events clear them through snapshot replay.

- [ ] **Step 5: Add decorator and legacy-compat tests**

Add to `internal/agentrt/runledger_mirror_store_test.go`:

```go
func TestRunLedgerMirrorStore_RepeatedBlockedRequestPreservesStableIDAndUpdatesAttemptMetadata(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:               "arun-repeat",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "old",
		GrantRequestID:   "grant-arun-repeat-exec",
		GrantAttempt:     1,
		GrantState:       "pending",
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-repeat",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-repeat", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		ApplyGrantAttempt:     true,
		ApplyGrantState:       true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "new",
		GrantRequestID:        "grant-arun-repeat-exec",
		GrantAttempt:          2,
		GrantState:            "pending",
	})
	require.NoError(t, err)

	snap, err := ledger.GetRunSnapshot(ctx, "arun-repeat")
	require.NoError(t, err)
	assert.Equal(t, "grant-arun-repeat-exec", snap.TeammateGrantRequestID)
assert.Equal(t, 2, snap.TeammateGrantAttempt)
assert.Equal(t, "pending", snap.TeammateGrantState)
}
```

Add an attempt-only replacement regression test too:

```go
func TestRunLedgerMirrorStore_RepeatedBlockedRequest_AttemptOnlyUpdateStillAppendsEvent(t *testing.T) {
	ctx := context.Background()
	base := NewInMemoryAgentRunStore()
	ledger := runledger.NewMemoryStore()

	require.NoError(t, base.Create(&AgentRun{
		ID:               "arun-repeat-attempt-only",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "same",
		GrantRequestID:   "grant-arun-repeat-attempt-only-exec",
		GrantAttempt:     1,
		GrantState:       "pending",
	}))

	store := NewRunLedgerMirrorStore(base, ledger, nil)
	require.NoError(t, ledger.AppendJournalEvent(ctx, runledger.JournalEvent{
		RunID:   "arun-repeat-attempt-only",
		Type:    runledger.EventRunCreated,
		Payload: json.RawMessage(`{"goal":"teammate"}`),
	}))

	err := store.UpdateProjection("arun-repeat-attempt-only", RunProjectionPatch{
		ApplyRuntimeCondition: true,
		ApplyBlockedReason:    true,
		ApplyGrantRequestID:   true,
		ApplyGrantAttempt:     true,
		ApplyGrantState:       true,
		RuntimeCondition:      AgentRunConditionBlockedWaitingApproval,
		BlockedReason:         "same",
		GrantRequestID:        "grant-arun-repeat-attempt-only-exec",
		GrantAttempt:          2,
		GrantState:            "pending",
	})
	require.NoError(t, err)

	events, err := ledger.GetJournalEvents(ctx, "arun-repeat-attempt-only")
	require.NoError(t, err)
	assert.Equal(t, runledger.EventTeammateApprovalBlocked, events[len(events)-1].Type)

	snap, err := ledger.GetRunSnapshot(ctx, "arun-repeat-attempt-only")
	require.NoError(t, err)
	assert.Equal(t, "grant-arun-repeat-attempt-only-exec", snap.TeammateGrantRequestID)
	assert.Equal(t, 2, snap.TeammateGrantAttempt)
	assert.Equal(t, "pending", snap.TeammateGrantState)
}
```

Also extend legacy snapshot compatibility tests in `internal/runledger/store_test.go` with concrete code so missing attempt metadata fields still load as zero values:

```go
func TestRunSnapshot_JSONUnmarshalLegacySnapshotMissingGrantAttemptMetadata(t *testing.T) {
	legacy := []byte(`{
		"run_id":"run-legacy",
		"status":"running",
		"notes":{"k":"v"},
		"teammate_runtime_condition":"blocked_waiting_approval",
		"teammate_blocked_reason":"dangerous tool requires approval",
		"teammate_grant_request_id":"grant-run-legacy-exec",
		"last_journal_seq":7
	}`)

	var snap RunSnapshot
	require.NoError(t, json.Unmarshal(legacy, &snap))

	assert.Equal(t, "grant-run-legacy-exec", snap.TeammateGrantRequestID)
	assert.Equal(t, 0, snap.TeammateGrantAttempt)
	assert.Empty(t, snap.TeammateGrantState)
}
```

- [ ] **Step 6: Run the RunLedger and agentrt packages**

Run:

```bash
go test ./internal/runledger -count=1
go test ./internal/agentrt -count=1
```

Expected: PASS.

- [ ] **Step 7: Suggested commit**

Suggested commit message:

```bash
feat: mirror teammate approval attempt metadata into runledger
```

## Task 6: Update Public Docs and Close the Change

**Files:**
- Modify: `docs/features/multi-agent.md`
- Modify: `docs/features/run-ledger.md`

- [ ] **Step 1: Document stable logical approval identity in multi-agent docs**

Add a short paragraph near the `agent_wait` section in `docs/features/multi-agent.md`:

```markdown
For built-in teammate approval blocking, `grant_request_id` identifies the stable logical blocked request for a given `(run, tool)`. If the same request is surfaced again later, the request ID stays stable and any renewed-attempt semantics appear through separate metadata such as `grant_attempt` or `grant_state`.
```

- [ ] **Step 2: Document durable attempt metadata in RunLedger docs**

Add a short paragraph near the teammate approval-blocked section in `docs/features/run-ledger.md`:

```markdown
The teammate approval-blocked durable mirror preserves both the stable logical `grant_request_id` and the latest attempt metadata derived from approval-block journal events. Repeated attempts for the same logical blocked request do not require rotating the request ID.
```

- [ ] **Step 3: Run final verification**

Run:

```bash
openspec validate teammate-approval-identity-policy --strict
go build ./...
go test ./...
```

Expected: all PASS.

- [ ] **Step 4: Archive the change**

Run:

```bash
openspec archive -y teammate-approval-identity-policy
```

Expected: archive succeeds and main specs are synced.

- [ ] **Step 5: Suggested commit**

Suggested commit message:

```bash
docs: finalize teammate approval identity policy
```
