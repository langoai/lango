# Multi-Agent Coworking Slice 5 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use `superpowers:subagent-driven-development` for each task. Every task must go through implementation, spec-compliance review, and code-quality review before the next task starts.

**Goal:** Land the first practical slice of Slice 5 by making mission-linked local coworking visible in Mission Control. The user should be able to see collaborating agents, recent handoffs, local blocked-on-teammate / blocked-on-approval states, and mission-linked budget/recovery signals without pretending full external team UX already exists.

**Architecture:** Slice 5 remains projection-first. It adds a collaboration projection over mission-linked local runtime signals:

1. mission execution links
2. linked `AgentRun` runtime state
3. recent delegation edges derived from `TurnTraceStore`, then attached only when mission attribution is provable
4. mission-attributed live budget/recovery signals from an app-level in-memory runtime buffer

External P2P team state remains out of the primary first slice.

Canonical source contract for the first slice:

- participant / blocked / waiting state: linked `AgentRun` and linked `RunLedger`
- handoff edges: `TurnTraceStore` session traces, then filtered through mission attribution
- budget/recovery: app-level mission-attributed runtime buffer built from live EventBus signals
- `reviewing`: linked `RunLedger` `verify_pending` / orchestrator-review-needed state only

**Tech Stack:** Go, Bubble Tea, EventBus, `internal/collabview`, `internal/app`, `internal/cli/cockpit`, OpenSpec, Zensical docs

## Scope Guardrails

- first slice is **mission-linked local coworking** only
- session-level delegation/budget/recovery must not be over-attributed to a mission
- external P2P team coordination remains secondary in this slice
- no new durable collaboration table in this slice
- no cockpit controls for team formation, role editing, or conflict resolution in this slice

## File Map

### Worker A: OpenSpec / Docs / Public Truth

- Create: `openspec/changes/multi-agent-coworking-slice-five/proposal.md`
- Create: `openspec/changes/multi-agent-coworking-slice-five/design.md`
- Create: `openspec/changes/multi-agent-coworking-slice-five/tasks.md`
- Create: `openspec/changes/multi-agent-coworking-slice-five/specs/mission-control-tui/spec.md`
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

### Worker B: Collaboration Projection Domain / App Wiring

- Create: `internal/collabview/types.go`
- Create: `internal/collabview/projector.go`
- Create: `internal/collabview/projector_test.go`
- Create: `internal/app/bridge_collaboration_runtime.go`
- Create: `internal/app/bridge_collaboration_runtime_test.go`
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/app_test.go`

### Worker C: Mission Control Collaboration Surface

- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/deps_test.go`
- Modify: `internal/cli/cockpit/missioncontrol_types.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector_test.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol_test.go`
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`

## Task Breakdown

### Task 1: Create the Slice 5 OpenSpec Change

**Owner:** Worker A

**Files:**
- Create: `openspec/changes/multi-agent-coworking-slice-five/proposal.md`
- Create: `openspec/changes/multi-agent-coworking-slice-five/design.md`
- Create: `openspec/changes/multi-agent-coworking-slice-five/tasks.md`
- Create: `openspec/changes/multi-agent-coworking-slice-five/specs/mission-control-tui/spec.md`

- [ ] **Step 1: Create the change skeleton**

The change must capture:

- mission-linked local coworking
- collaboration projection, not a new durable model
- local handoff / participant / blocked / budget / recovery visibility
- external P2P team UX remaining secondary

- [ ] **Step 2: Add a Mission Control delta**

The `mission-control-tui` delta must state:

- mission rows can show collaboration context
- handoffs and local collaboration state are mission-linked, not raw session noise
- external team management is not part of this slice

- [ ] **Step 3: Validate the change**

Run:

```bash
openspec validate multi-agent-coworking-slice-five --strict
```

Expected:

```text
Change 'multi-agent-coworking-slice-five' is valid
```

### Task 2: Add the Collaboration Projection Domain

**Owner:** Worker B

**Files:**
- Create: `internal/collabview/types.go`
- Create: `internal/collabview/projector.go`
- Create: `internal/collabview/projector_test.go`

- [ ] **Step 1: Define collaboration types**

Add explicit types such as:

- `CollaborationState`
- `ParticipantView`
- `HandoffEdge`
- `BudgetSignal`
- `RecoverySignal`
- `CollaborationView`

- [ ] **Step 2: Implement mission-linked attribution rules**

The projector must only attach collaboration signals to a mission when attribution is provable through a mission-linked local execution.

- delegation edges require mission-linked execution attribution
- budget/recovery signals require mission-linked local execution attribution
- otherwise the signal stays out of mission collaboration view

- [ ] **Step 3: Add collaboration-state derivation**

First-slice states:

- `solo`
- `delegating`
- `waiting_on_teammate`
- `reviewing`
- `blocked_on_approval`
- `recovering`

`reviewing` may only come from:

- a mission-linked local execution in `verify_pending` / orchestrator-review-needed state
- not from Slice 4 loop/follow-up state in this first slice

- [ ] **Step 4: Add projection tests**

Cover:

- mission-linked delegation attribution
- non-attributable session-level delegation ignored
- participant extraction
- blocked_on_approval / waiting_on_teammate / recovering state
- review-state source rules
- budget/recovery only when attribution is provable

Run:

```bash
go test ./internal/collabview -count=1
```

Expected:

```text
ok
```

### Task 3: Wire Collaboration Readers Into the App

**Owner:** Worker B

**Files:**
- Create: `internal/app/bridge_collaboration_runtime.go`
- Create: `internal/app/bridge_collaboration_runtime_test.go`
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/app_test.go`

- [ ] **Step 1: Expose narrow collaboration readers**

Expose only the narrow local collaboration readers Mission Control needs:

- mission execution-link reader
- linked `AgentRun` reader
- mission-linked delegation source backed by `TurnTraceStore`
- mission-attributed local budget/recovery source backed by an app-level in-memory runtime buffer

- [ ] **Step 2: Keep external team data secondary**

Do not expose full P2P team-control surfaces as a first-slice Mission Control dependency here.

- [ ] **Step 3: Add an app-level collaboration runtime bridge**

The bridge should:

- subscribe to live budget/recovery signals
- keep a small in-memory buffer
- only emit mission-attributed records when attribution is provable

- [ ] **Step 4: Add wiring tests**

Cover:

- collaboration readers are present when local sources exist
- external team-only surfaces are not implied by default
- budget/recovery runtime bridge only stores mission-attributed records

Run:

```bash
go test ./internal/app -run 'Test.*Collaboration' -count=1
```

Expected:

```text
ok
```

### Task 4: Add Collaboration Context To Mission Control

**Owner:** Worker C

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/deps_test.go`
- Modify: `internal/cli/cockpit/missioncontrol_types.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector_test.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol_test.go`
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`

- [ ] **Step 1: Add collaboration read dependencies to cockpit**

Cockpit deps should accept only the narrow collaboration readers required by the collaboration projector.

- [ ] **Step 2: Surface collaboration context without replacing the mission board**

Mission Control should add compact collaboration context such as:

- participant summary
- recent handoff summary
- collaboration-state hint
- budget/recovery hint when attributable

- [ ] **Step 3: Keep collaboration compact in the first slice**

Do not build a separate collaboration dashboard yet.

- [ ] **Step 4: Add Mission Control collaboration tests**

Cover:

- participants render from linked local signals
- handoff summary only appears when attributable
- blocked_on_approval / waiting_on_teammate / recovering hints
- no external-team overstatement

Run:

```bash
go test ./internal/cli/cockpit ./internal/cli/cockpit/pages ./cmd/lango -run 'TestMissionControl(Collaboration|Handoff)|TestRunCockpit' -count=1
```

Expected:

```text
ok
```

### Task 5: Update Public Docs And Complete OpenSpec

**Owner:** Worker A

**Files:**
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

- [ ] **Step 1: Audit landed Slice 5 behavior before editing docs**

Verify in code:

- collaboration projection exists
- local mission-linked signals are the first-slice source
- external P2P team UX is still secondary

- [ ] **Step 2: Update docs**

Document only landed behavior:

- which collaboration state is visible
- how local teammates appear in Mission Control
- what external team functionality is still out of scope

- [ ] **Step 3: Build docs**

Run:

```bash
.venv/bin/zensical build
```

Expected:

```text
Build finished
```

- [ ] **Step 4: Complete OpenSpec workflow**

After implementation and verification:

- `openspec validate multi-agent-coworking-slice-five --strict`
- archive the change

### Task 6: Final Verification And Slice Review

**Owner:** Main agent

- [ ] **Step 1: Run focused suites**

Run:

```bash
go test ./internal/collabview ./internal/app ./internal/cli/cockpit ./internal/cli/cockpit/pages ./cmd/lango -count=1
```

Expected:

```text
ok
```

- [ ] **Step 2: Run repository-wide verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
openspec validate --changes --strict
```

Expected:

```text
[all commands succeed]
```

- [ ] **Step 3: Run final Slice 5 review**

Before claiming Slice 5 complete:

- request final spec-compliance review
- request final code-quality review
- fix all Critical and Major findings

- [ ] **Step 4: Archive and record final state**

After review and verification succeed:

- archive the Slice 5 change
- confirm main specs were updated
- record the final commit SHA in the work log
