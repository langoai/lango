# Mission Lifecycle Wave 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use `superpowers:subagent-driven-development` for each task. Every task must go through implementation, spec-compliance review, and code-quality review before the next task starts.

**Goal:** Land Wave 2 durable mission lifecycle so Mission Control no longer relies on runtime projection alone. Durable mission rows, mission history, execution links, and coarse decision state must exist in the real app, not only in tests.

**Architecture:** Keep `RunLedger` as execution truth and add a separate durable mission layer in Ent. `MissionService` is the single writer for mission latest state, mission history, and mission-execution links. Mission Control becomes durable-first, but still overlays unmatched runtime work and live activity/approval signals. Wave 2 is only complete if there are real write paths for:

1. direct mission start from Mission Control
2. accepting a proposed mission from Mission Control
3. execution-link attachment when mission-bound work creates a background task or run
4. approval attribution that drives durable `waiting_decision`

**Tech Stack:** Go, Ent, Bubble Tea, `internal/mission`, `internal/storage`, `internal/app`, `internal/toolchain`, `internal/background`, `internal/runledger`, OpenSpec, Zensical docs

## Non-Goals

- Do not make `TaskEntry` the authoritative mission checklist model in this wave.
- Do not auto-create durable missions for every background task or run.
- Do not persist a full durable approval queue; Wave 2 stores only coarse mission decision summary.
- Do not make standalone `lango chat` a mission-native surface yet. Wave 2 direct mission creation is scoped to Mission Control.

## File Map

### Worker A: OpenSpec / Docs / Public Truth

- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/proposal.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/design.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/tasks.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/specs/mission-control-tui/spec.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/specs/agent-control-plane-tools/spec.md`
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

### Worker B: Durable Mission Persistence / App Boundary

- Create: `internal/ent/schema/mission.go`
- Create: `internal/ent/schema/mission_state_history.go`
- Create: `internal/ent/schema/mission_execution_link.go`
- Create: `internal/mission/store.go`
- Create: `internal/mission/store_test.go`
- Create: `internal/mission/service.go`
- Create: `internal/mission/service_test.go`
- Modify: `internal/storage/facade.go`
- Modify: `internal/appinit/module.go`
- Create: `internal/app/modules_mission.go`
- Modify: `internal/app/modules.go`
- Modify: `internal/app/modules_runledger.go`
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/modules_test.go`
- Modify: `internal/app/app_test.go`

### Worker C: Mission Write Paths / Approval / Durable Read Surface

- Modify: `internal/ctxkeys/ctxkeys.go`
- Modify: `internal/approval/approval.go`
- Modify: `internal/toolchain/mw_approval.go`
- Modify: `internal/toolchain/middleware_test.go`
- Modify: `internal/background/tools.go`
- Create: `internal/background/tools_test.go`
- Modify: `internal/runledger/tools.go`
- Modify: `internal/runledger/tools_test.go`
- Modify: `internal/agentrt/control_tools.go`
- Modify: `internal/agentrt/control_tools_test.go`
- Modify: `internal/cli/chat/chat.go`
- Modify: `internal/cli/chat/chat_test.go`
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/deps_test.go`
- Modify: `internal/cli/cockpit/cockpit.go`
- Modify: `internal/cli/cockpit/cockpit_test.go`
- Modify: `internal/cli/cockpit/learning_buffer.go`
- Modify: `internal/cli/cockpit/learning_buffer_test.go`
- Modify: `internal/cli/cockpit/missioncontrol_types.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector_test.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol_test.go`
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`

## Task Breakdown

### Task 1: Create the Wave 2 OpenSpec Change

**Owner:** Worker A

**Files:**
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/proposal.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/design.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/tasks.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/specs/mission-control-tui/spec.md`
- Create: `openspec/changes/2026-05-03-mission-lifecycle-wave-two/specs/agent-control-plane-tools/spec.md`

- [ ] **Step 1: Create the change skeleton**

Write proposal/design/tasks that match the approved Wave 2 direction:

- hybrid storage
- separate durable `mission_id`
- durable mission latest-state row plus append-only state history
- `MissionExecutionLink` as the only durable execution relationship truth
- durable rows begin at `prepared`, not `proposed`

- [ ] **Step 2: Add change-local Mission Control delta**

The `mission-control-tui` delta must state:

- durable mission rows are the primary read source
- unmatched runtime work remains visible until linked or dismissed
- direct mission start and proposal acceptance are real app write paths
- `waiting_decision` is durable coarse state, not a durable approval queue

- [ ] **Step 3: Add change-local agent-control delta**

The `agent-control-plane-tools` delta must state:

- `TaskEntry` remains lightweight in this wave
- Wave 2 does not promote task tracking into the durable mission checklist model
- mission-aware execution linkage is attached at execution creation sites, not by retrofitting all task tracking into mission truth

- [ ] **Step 4: Validate the change**

Run:

```bash
openspec validate 2026-05-03-mission-lifecycle-wave-two --strict
```

Expected:

```text
Change '2026-05-03-mission-lifecycle-wave-two' is valid
```

### Task 2: Add Ent Schemas for Durable Missions

**Owner:** Worker B

**Files:**
- Create: `internal/ent/schema/mission.go`
- Create: `internal/ent/schema/mission_state_history.go`
- Create: `internal/ent/schema/mission_execution_link.go`

- [ ] **Step 1: Add `Mission` schema**

Implement fields equivalent to:

```go
field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable()
field.String("session_key").NotEmpty()
field.String("title").NotEmpty()
field.Text("description").Optional().Nillable()
field.Enum("status").Values("prepared", "active", "waiting_decision", "blocked", "done", "cancelled").Default("prepared")
field.String("source_kind").NotEmpty()
field.String("source_ref").Optional().Nillable()
field.String("current_blocked_reason").Optional().Nillable()
field.String("current_decision_kind").Optional().Nillable()
field.Text("current_decision_summary").Optional().Nillable()
field.Time("created_at").Default(time.Now).Immutable()
field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now)
field.Time("completed_at").Optional().Nillable()
```

- [ ] **Step 2: Add `Mission` indexes**

Wave 2 reads missions by session and recent activity. Add at least:

```go
index.Fields("session_key", "updated_at")
index.Fields("status")
index.Fields("source_kind", "source_ref")
```

- [ ] **Step 3: Add `MissionStateHistory` schema**

Implement fields equivalent to:

```go
field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable()
field.UUID("mission_id", uuid.UUID{})
field.Int64("seq")
field.Enum("from_status").Values("prepared", "active", "waiting_decision", "blocked", "done", "cancelled").Optional().Nillable()
field.Enum("to_status").Values("prepared", "active", "waiting_decision", "blocked", "done", "cancelled")
field.String("reason").Optional().Nillable()
field.String("actor_kind").NotEmpty()
field.String("actor_ref").Optional().Nillable()
field.String("execution_kind").Optional().Nillable()
field.String("execution_ref").Optional().Nillable()
field.String("decision_kind").Optional().Nillable()
field.Text("decision_summary").Optional().Nillable()
field.JSON("payload", map[string]any{}).Optional()
field.Time("created_at").Default(time.Now).Immutable()
```

Add indexes:

```go
index.Fields("mission_id", "seq").Unique()
index.Fields("mission_id", "created_at")
```

- [ ] **Step 4: Add `MissionExecutionLink` schema**

Implement fields equivalent to:

```go
field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable()
field.UUID("mission_id", uuid.UUID{})
field.Enum("execution_kind").Values("runledger_run", "task_os_execution")
field.String("execution_ref").NotEmpty()
field.Enum("link_role").Values("primary", "followup", "retry", "research", "draft", "handoff").Default("primary")
field.Time("created_at").Default(time.Now).Immutable()
```

Indexes required:

```go
index.Fields("mission_id", "execution_kind", "execution_ref").Unique()
index.Fields("execution_kind", "execution_ref")
index.Fields("mission_id", "link_role")
```

- [ ] **Step 5: Generate Ent code**

Run:

```bash
go generate ./internal/ent/...
```

Expected:

```text
[no error output]
```

### Task 3: Build the Mission Store

**Owner:** Worker B

**Files:**
- Create: `internal/mission/store.go`
- Create: `internal/mission/store_test.go`

- [ ] **Step 1: Define the store interface**

Keep the interface centered on Wave 2 reads and writes:

```go
type Store interface {
	CreateMission(ctx context.Context, in CreateMissionInput) (*Mission, error)
	GetMission(ctx context.Context, missionID string) (*Mission, error)
	ListMissionsBySession(ctx context.Context, sessionKey string, limit int) ([]*Mission, error)
	TransitionMission(ctx context.Context, in TransitionMissionInput) (*Mission, error)
	AppendExecutionLink(ctx context.Context, in AppendExecutionLinkInput) error
	ListExecutionLinks(ctx context.Context, missionID string) ([]*ExecutionLink, error)
	FindMissionByExecution(ctx context.Context, executionKind, executionRef string) (*Mission, error)
}
```

- [ ] **Step 2: Implement create/get/list**

Use `ent.Client` directly and preserve repo conventions:

- create latest row
- fetch by UUID/string conversion
- list by `session_key` and `updated_at desc`

- [ ] **Step 3: Implement transition with per-mission `seq` history append**

Transition path must:

- load latest mission
- reject invalid transitions
- append one `MissionStateHistory`
- update latest row atomically

- [ ] **Step 4: Implement execution-link writes and reverse lookup**

Support:

- append unique mission execution link
- list links by mission
- reverse lookup by `(execution_kind, execution_ref)`

- [ ] **Step 5: Add store tests**

Cover:

- create + get + list ordering
- transition appends `seq` history
- duplicate execution link rejected
- reverse lookup by execution reference

Run:

```bash
go test ./internal/mission -run 'Test(Store|Mission)' -count=1
```

Expected:

```text
ok
```

### Task 4: Build the Mission Service

**Owner:** Worker B

**Files:**
- Create: `internal/mission/service.go`
- Create: `internal/mission/service_test.go`

- [ ] **Step 1: Define service inputs**

Concrete inputs should include:

```go
type StartMissionInput struct {
	SessionKey  string
	Title       string
	Description string
	SourceKind  string
	SourceRef   string
	StartActive bool
}

type AcceptProposalInput struct {
	SessionKey  string
	SourceKind  string
	SourceRef   string
	Title       string
	Description string
}
```

- [ ] **Step 2: Implement direct user-start and proposal-accept paths**

Rules:

- direct user-start creates the first durable mission row
- accepted proposal creates the first durable mission row
- no durable `proposed` mission row exists before acceptance
- initial state is `prepared` or `active`

- [ ] **Step 3: Implement durable decision and blocker transitions**

Rules:

- only one coarse durable decision marker lives on the mission row:
  `current_decision_kind`, `current_decision_summary`
- `waiting_decision` does not point to a durable approval queue row
- denied or timed-out approval keeps the mission in a decision-needed state until new direction arrives

- [ ] **Step 4: Implement execution-link helpers**

Add service methods for:

- attaching `runledger_run` and `task_os_execution`
- resolving mission by execution reference
- refreshing mission latest state from execution outcomes when the link already exists

- [ ] **Step 5: Add service tests**

Cover:

- user-start creates `prepared` mission
- accepted proposal creates first durable mission row
- `waiting_decision` stores descriptive decision summary
- execution link attachment is idempotent
- invalid transitions rejected

Run:

```bash
go test ./internal/mission -run 'TestService' -count=1
```

Expected:

```text
ok
```

### Task 5: Wire Mission Persistence Through Storage and App

**Owner:** Worker B

**Files:**
- Modify: `internal/storage/facade.go`
- Modify: `internal/appinit/module.go`
- Create: `internal/app/modules_mission.go`
- Modify: `internal/app/modules.go`
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/modules_test.go`
- Modify: `internal/app/app_test.go`

- [ ] **Step 1: Extend the storage facade**

Keep bootstrap boundary discipline intact:

- `storage.Facade` must expose a mission accessor wired from `WithEntClient`
- mission bootstrap must not reach around the facade for raw DB or raw `ent.Client`

- [ ] **Step 2: Add `ProvidesMission` and `missionModule`**

Follow existing `appinit` conventions:

- add a new `ProvidesMission` key
- create `missionValues`
- build `MissionStore` and `MissionService` inside `missionModule`
- depend on the storage-backed mission accessor, not direct DB handles
- keep the module nil-safe: when durable mission storage is unavailable, leave mission wiring disabled instead of creating a second non-durable truth

- [ ] **Step 3: Populate `App` fields**

Expose:

```go
MissionStore   mission.Store
MissionService *mission.Service
```

Mirror the existing `RunLedger` and automation patterns in `populateAppFields`.

- [ ] **Step 4: Wire mission-backed adapters into the app boundary**

When `MissionService` exists, app wiring must expose the pieces later tasks need:

- approval lifecycle observer passed through the real `toolchain.WithApproval(...)` composition site in `internal/app/app.go`
- execution-link adapter passed through the real `background.BuildTools(...)` and `runledger.BuildTools(...)` call sites in app/module wiring
- nil-storage behavior must keep these adapters absent, with `App.MissionStore` and `App.MissionService` left nil

- [ ] **Step 5: Add wiring tests**

Cover:

- facade mission accessor exists when `WithEntClient` is used
- module enabled path yields `MissionStore` and `MissionService`
- missing storage keeps mission wiring disabled and nil-safe
- `populateAppFields` copies them into `App`
- app-level composition tests verify the approval observer and execution-link adapters are actually wired at `app.New(...)` and module call sites, not only present on intermediate structs

Run:

```bash
go test ./internal/app ./internal/storage -run 'Test.*Mission|TestWithEntClient' -count=1
```

Expected:

```text
ok
```

### Task 6: Add Real Mission Creation Paths in Mission Control

**Owner:** Worker C

**Files:**
- Modify: `internal/ctxkeys/ctxkeys.go`
- Modify: `internal/cli/chat/chat.go`
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/cockpit.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol_test.go`
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`

- [ ] **Step 1: Add mission runtime binding helpers**

Use `internal/ctxkeys` for dependency-light context propagation:

- mission binding (`mission_id` only; session identity continues to flow through the existing `internal/session` context path)
- optional execution-link adapter interface for mission-aware execution creators
- optional approval-attribution metadata carrier for mission-bound approvals

- [ ] **Step 2: Add a page-controlled chat submit hook**

Do not fork chat submission logic. Add a narrow `ChatModel` API so Mission Control can:

- inspect pending composer input
- create a mission before dispatch
- submit through the existing shared `TurnRunner` path with a decorated parent context

Standalone `lango chat` remains unchanged in this wave.

- [ ] **Step 3: Start durable missions from Mission Control composer**

When Mission Control composer submits a top-level request:

- create a durable mission via `MissionService.StartMission(...)`
- decorate the turn context with the new mission binding plus mission-backed adapters
- refresh local page state immediately after submission
- durable-row rendering itself lands in Task 8, once the projector reads durable missions first

- [ ] **Step 4: Accept proposed missions from Mission Control**

Selected `MissionKindProposed` rows must have a real accept path:

- carry enough source metadata in `MissionView` to call `AcceptProposal(...)` without re-deriving from display text
- accept the suggestion through `MissionService.AcceptProposal(...)`
- create the durable mission row
- dismiss the accepted suggestion from `LearningSuggestionBuffer`
- refresh local page state; durable-row rendering is finalized in Task 8

- [ ] **Step 5: Add Mission Control write-path tests**

Cover:

- composer submission creates a durable mission before turn dispatch
- chat page submission does not implicitly create a mission
- accepting a proposed mission creates a durable row and removes the transient overlay entry

Run:

```bash
go test ./internal/cli/cockpit/pages ./cmd/lango -run 'TestMissionControl|TestRunCockpit' -count=1
```

Expected:

```text
ok
```

### Task 7: Attribute Approvals and Execution Creation to Missions

**Owner:** Worker C

**Files:**
- Modify: `internal/approval/approval.go`
- Modify: `internal/toolchain/mw_approval.go`
- Modify: `internal/toolchain/middleware_test.go`
- Modify: `internal/app/app.go`
- Modify: `internal/background/tools.go`
- Create: `internal/background/tools_test.go`
- Modify: `internal/app/modules.go`
- Modify: `internal/runledger/tools.go`
- Modify: `internal/runledger/tools_test.go`
- Modify: `internal/app/modules_runledger.go`
- Modify: `internal/agentrt/control_tools.go`
- Modify: `internal/agentrt/control_tools_test.go`

- [ ] **Step 1: Extend approval metadata for mission-aware flows**

Add optional fields to `approval.ApprovalRequest` for coarse attribution only:

- `MissionID`
- `ExecutionKind`
- `ExecutionRef`

These fields are optional and must not break existing approval providers.

- [ ] **Step 2: Add approval lifecycle observation without a layer violation**

`toolchain` must not import `mission` directly. Add an injected observer interface in the lower layer so the app can supply a mission-backed adapter at the real `WithApproval(...)` composition site in `internal/app/app.go`.

Required behavior:

- when approval is requested for a mission-bound turn, mark the mission `waiting_decision`
- when approval is granted, clear the decision marker and return to `active`
- when approval is denied or times out, keep the mission in a decision-needed state with updated summary

- [ ] **Step 3: Attach execution links at creation sites**

When a mission-bound turn creates runtime work:

- `bg_submit` attaches `task_os_execution`
- `run_create` attaches `runledger_run`
- `agent_spawn` / `AgentControlPlane.Submitter.Submit(...)` preserves the mission binding so spawned background work attaches to the same mission

Use the mission binding from context and the injected execution-link adapter. Do not add a second truth source outside `MissionExecutionLink`.

- [ ] **Step 4: Add attribution tests**

Cover:

- mission-bound approval request transitions mission to `waiting_decision`
- approval grant clears durable decision marker
- denial/timeout leaves mission in a coherent decision-needed state
- `bg_submit` attaches a `task_os_execution` link
- `run_create` attaches a `runledger_run` link
- mission-bound spawned child work stays linked through the `agent_spawn` submitter path

Run:

```bash
go test ./internal/toolchain ./internal/background ./internal/runledger ./internal/agentrt -run 'TestWithApproval|TestBuildTools|TestAgentSpawn' -count=1
```

Expected:

```text
ok
```

### Task 8: Rebase Mission Control on Durable Missions

**Owner:** Worker C

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/missioncontrol_types.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector.go`
- Modify: `internal/cli/cockpit/missioncontrol_projector_test.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `cmd/lango/main.go`
- Modify: `cmd/lango/main_test.go`

- [ ] **Step 1: Add a durable mission reader dependency**

Expose a cockpit read surface like:

```go
type MissionReader interface {
	ListMissionsBySession(ctx context.Context, sessionKey string, limit int) ([]*mission.Mission, error)
	ListExecutionLinks(ctx context.Context, missionID string) ([]*mission.ExecutionLink, error)
}
```

- [ ] **Step 2: Make durable missions the first read source**

Read order:

1. durable mission rows
2. linked execution summaries
3. unmatched runtime overlays

- [ ] **Step 3: Preserve Wave 1 runtime visibility**

If runtime work exists but is not yet linked to a durable mission:

- keep it visible only when it belongs to the current cockpit session
- mark it as unmatched runtime overlay
- do not silently hide it
- use `TaskSnapshot.OriginSession` as the primary session filter so one cockpit session does not render another session's background work

- [ ] **Step 4: Add projector tests**

Cover:

- durable missions render first
- linked run/task enriches durable mission
- unmatched runtime work still renders
- durable `waiting_decision` plus live approval overlay remain coherent
- loading and degraded states still work when mission reader is unavailable
- runtime wiring injects the mission reader through `runCockpit` / `cmd/lango/main.go`

Run:

```bash
go test ./internal/cli/cockpit ./cmd/lango -run 'TestMissionControl(Projector|DurableMission|UnmatchedRuntime)|TestRunCockpit' -count=1
```

Expected:

```text
ok
```

### Task 9: Update Public Docs and Complete OpenSpec

**Owner:** Worker A

**Files:**
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

- [ ] **Step 1: Audit landed Wave 2 behavior before writing docs**

Verify in code:

- durable mission row exists
- mission history exists
- mission-execution links exist
- Mission Control can start missions and accept proposals
- Mission Control reads durable missions first while keeping unmatched runtime visibility

- [ ] **Step 2: Update public docs**

Document only landed behavior:

- durable mission identity
- latest mission state + history
- execution links
- decision summary semantics
- direct mission start / proposal acceptance in Mission Control
- task tracking remains lightweight and non-authoritative

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

- `openspec validate 2026-05-03-mission-lifecycle-wave-two --strict`
- `openspec verify 2026-05-03-mission-lifecycle-wave-two --strict`
- archive the change so delta specs merge into main specs

### Task 10: Final Verification and Wave Review

**Owner:** Main agent

- [ ] **Step 1: Run focused suites**

Run:

```bash
go test ./internal/mission ./internal/app ./internal/storage ./internal/toolchain ./internal/background ./internal/runledger ./internal/cli/cockpit ./cmd/lango -count=1
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

- [ ] **Step 3: Run final Wave 2 review**

Before claiming Wave 2 complete:

- request final spec-compliance review against the Wave 2 change
- request final code-quality review against the landed implementation
- fix all Critical and Major findings

- [ ] **Step 4: Archive and record final state**

After reviews and verification succeed:

- archive the Wave 2 OpenSpec change
- confirm the main specs were updated
- record the final commit SHA in the work log

## Acceptance Summary

Wave 2 is complete only when all of the following are true:

- Mission Control can create a durable mission from real user input
- Mission Control can accept a transient proposed mission into a durable row
- approval requests can move a mission into durable `waiting_decision`
- mission-bound background tasks or runs can attach `MissionExecutionLink` rows
- Mission Control reads durable missions first without hiding unmatched runtime work
- all new app/storage boundaries are wired through `storage.Facade` and `appinit`
- build, tests, docs build, OpenSpec validate/verify/archive all succeed
