# Proactive Agent Behavior Slice 3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use `superpowers:subagent-driven-development` for each task. Every task must go through implementation, spec-compliance review, and code-quality review before the next task starts.

**Goal:** Land the first practical slice of Slice 3 by turning existing learning suggestions into transient proposals with automatic low-risk preparation, then letting Mission Control accept those prepared proposals into durable missions without losing the prepared context.

**Architecture:** Slice 3 keeps durable mission truth in the existing `mission` layer and adds a separate transient proposal layer. The first slice is intentionally narrow:

1. use `LearningSuggestionEvent` as the only active producer
2. create transient proposal records with explicit lifecycle
3. automatically generate a deterministic prepared brief from source-native evidence
4. show prepared proposals in Mission Control
5. accept a prepared proposal into a durable mission while preserving that prepared brief

This slice does **not** yet enable librarian-gap proposals or runtime-failure proposals, and it does **not** yet launch generic proposal-owned background executions.

**Tech Stack:** Go, Bubble Tea, EventBus, `internal/proposal`, `internal/app`, `internal/cli/cockpit`, OpenSpec, Zensical docs

## Scope Guardrails

- `proposed` remains transient; no durable `Mission` row is created before acceptance
- only `LearningSuggestionEvent` is an active producer in this slice
- preparation is deterministic and source-native, not broad heuristic agent work
- no filesystem mutation, external messaging, payment, calendar confirmation, or dangerous command execution in preparation
- no proposal-owned background executions yet; therefore no `proposal_id` execution-link promotion logic in this slice
- librarian and runtime-failure proposal producers stay explicitly deferred until their adapters exist

## File Map

### Worker A: OpenSpec / Docs / Public Truth

- Create: `openspec/changes/proactive-agent-behavior-slice-three/proposal.md`
- Create: `openspec/changes/proactive-agent-behavior-slice-three/design.md`
- Create: `openspec/changes/proactive-agent-behavior-slice-three/tasks.md`
- Create: `openspec/changes/proactive-agent-behavior-slice-three/specs/mission-control-tui/spec.md`
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

### Worker B: Proposal Domain / App Wiring

- Create: `internal/proposal/types.go`
- Create: `internal/proposal/registry.go`
- Create: `internal/proposal/registry_test.go`
- Create: `internal/proposal/preparer.go`
- Create: `internal/proposal/preparer_test.go`
- Create: `internal/proposal/service.go`
- Create: `internal/proposal/service_test.go`
- Create: `internal/app/modules_proposal.go`
- Modify: `internal/appinit/module.go`
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/modules_test.go`
- Modify: `internal/app/app_test.go`

### Worker C: Mission Control Proposal Surface / Acceptance

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

### Task 1: Create the Slice 3 OpenSpec Change

**Owner:** Worker A

**Files:**
- Create: `openspec/changes/proactive-agent-behavior-slice-three/proposal.md`
- Create: `openspec/changes/proactive-agent-behavior-slice-three/design.md`
- Create: `openspec/changes/proactive-agent-behavior-slice-three/tasks.md`
- Create: `openspec/changes/proactive-agent-behavior-slice-three/specs/mission-control-tui/spec.md`

- [ ] **Step 1: Create the change skeleton**

The change must capture:

- transient proposal model
- learning suggestion as the only active producer in this slice
- deterministic prepared brief generation
- Mission Control rendering prepared proposals
- durable mission acceptance preserving prepared context

- [ ] **Step 2: Add a Mission Control delta**

The `mission-control-tui` delta must state:

- proposed missions are now backed by a transient proposal registry rather than raw learning-buffer rows
- proposals can move through `suggested`, `preparing`, and `prepared`
- accepting a prepared proposal creates a durable mission and preserves the prepared brief
- this slice does not yet enable generic proposal-owned execution or librarian/runtime-failure proposal producers

- [ ] **Step 3: Validate the change**

Run:

```bash
openspec validate proactive-agent-behavior-slice-three --strict
```

Expected:

```text
Change 'proactive-agent-behavior-slice-three' is valid
```

### Task 2: Add the Transient Proposal Domain

**Owner:** Worker B

**Files:**
- Create: `internal/proposal/types.go`
- Create: `internal/proposal/registry.go`
- Create: `internal/proposal/registry_test.go`

- [ ] **Step 1: Define proposal types**

Add an explicit transient proposal model with:

- `Proposal`
- `ProposalStatus`
- `PreparedBrief`
- `ProposalSource`

Required statuses in this slice:

- `suggested`
- `preparing`
- `prepared`
- `dismissed`
- `accepted`
- `expired`

- [ ] **Step 2: Implement session-scoped proposal registry**

The registry should provide:

- upsert from producer event
- get by proposal ID
- list by session
- mark prepared
- dismiss
- accept
- prune expired

The first slice may keep this in-memory, but it must be deterministic and concurrency-safe.

- [ ] **Step 3: Add registry tests**

Cover:

- create/update proposal from same source
- session scoping
- dismiss and expiration behavior
- accept transitions proposal out of the visible active set

Run:

```bash
go test ./internal/proposal -run 'Test(Proposal|Registry)' -count=1
```

Expected:

```text
ok
```

### Task 3: Add Deterministic Proposal Preparation

**Owner:** Worker B

**Files:**
- Create: `internal/proposal/preparer.go`
- Create: `internal/proposal/preparer_test.go`
- Create: `internal/proposal/service.go`
- Create: `internal/proposal/service_test.go`

- [ ] **Step 1: Define a deterministic preparer contract**

The preparer in this slice must stay source-native and low-risk.

It should accept a learning-suggestion proposal source and produce a `PreparedBrief` containing:

- source summary
- reason / rationale
- suggested acceptance effect
- any supporting deterministic evidence already available from the source event

- [ ] **Step 2: Implement learning-suggestion preparation**

Rules:

- no LLM call
- no filesystem writes
- no background job launch
- no broad heuristic inference

This is a deterministic preparation pass over existing source fields.

- [ ] **Step 3: Add a proposal service and mutation surface**

The UI must not mutate the registry directly. Add an explicit `ProposalService` that owns:

- create or update proposal from producer input
- transition `suggested -> preparing -> prepared`
- dismiss proposal
- accept proposal
- prune expired proposals

This service is the write boundary the cockpit can call later.

- [ ] **Step 4: Make lifecycle timing explicit**

This slice uses deterministic preparation, so the lifecycle is:

1. producer event creates or updates `suggested`
2. service immediately transitions to `preparing`
3. preparer generates the deterministic brief
4. service transitions to `prepared`

Expiration and pruning rules:

- registry read paths should prune expired proposals opportunistically
- the proposal module may also prune on producer writes

- [ ] **Step 5: Add preparer and service tests**

Cover:

- learning suggestion becomes a stable prepared brief
- empty or partial source fields degrade gracefully
- preparation never mutates external state
- proposal service performs `suggested -> preparing -> prepared`
- dismiss and accept use the explicit service instead of direct registry mutation

Run:

```bash
go test ./internal/proposal -run 'Test(Preparer|ProposalService)' -count=1
```

Expected:

```text
ok
```

### Task 4: Wire Proposal Registry Into the App

**Owner:** Worker B

**Files:**
- Create: `internal/app/modules_proposal.go`
- Modify: `internal/appinit/module.go`
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/modules_test.go`
- Modify: `internal/app/app_test.go`

- [ ] **Step 1: Add proposal app values**

Expose app-level proposal components:

- `ProposalRegistry`
- `ProposalPreparer`
- `ProposalService`

- [ ] **Step 2: Add `proposalModule`**

The module should:

- subscribe to `LearningSuggestionEvent`
- create or update transient proposals through `ProposalService`
- run deterministic preparation through `ProposalService`
- keep proposal state session-scoped

- [ ] **Step 3: Keep non-learning producers disabled**

Librarian and runtime-failure producers must stay explicitly off in this slice. Do not add fallback heuristics.

- [ ] **Step 4: Add wiring tests**

Cover:

- proposal module enabled path
- event subscription creates or updates a proposal
- app exposes `ProposalService` as the write surface
- no librarian/runtime-failure producer activation in this slice

Run:

```bash
go test ./internal/app -run 'Test.*Proposal' -count=1
```

Expected:

```text
ok
```

### Task 5: Rebase Mission Control Proposed Lane on Proposal Registry

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

- [ ] **Step 1: Add a proposal reader dependency**

Cockpit deps should expose a read surface for active proposals by session.

- [ ] **Step 2: Add a proposal mutation dependency**

Cockpit deps should also expose a write surface for:

- dismiss proposal
- accept proposal

- [ ] **Step 3: Render proposals from the registry**

Mission Control should stop treating the learning buffer as the primary proposal source.

It should render transient proposals with:

- title
- reason
- preparation status
- prepared brief summary
- source metadata

- [ ] **Step 4: Keep the learning buffer only as compatibility fallback if needed**

If any compatibility fallback remains, it must be clearly secondary and not the main source of truth.

- [ ] **Step 5: Add projector tests**

Cover:

- prepared proposals render with prepared status
- session-scoped proposal listing
- fallback behavior remains honest when proposal registry is unavailable

Run:

```bash
go test ./internal/cli/cockpit ./cmd/lango -run 'TestMissionControl(Proposal|Projector)|TestRunCockpit' -count=1
```

Expected:

```text
ok
```

### Task 6: Accept Prepared Proposals Into Durable Missions Without Losing Context

**Owner:** Worker C

**Files:**
- Modify: `internal/cli/cockpit/pages/missioncontrol.go`
- Modify: `internal/cli/cockpit/pages/missioncontrol_test.go`

- [ ] **Step 1: Preserve prepared context on acceptance**

When the user accepts a prepared proposal:

- call `ProposalService.AcceptProposal(...)` first so transient proposal ownership changes through the write boundary
- then call `MissionService.AcceptProposal(...)`
- move the prepared brief into the durable mission description or equivalent durable user-facing field
- remove the proposal from the active proposal lane through the proposal service, not direct registry mutation

- [ ] **Step 2: Keep acceptance deterministic**

This slice does not attach proposal-owned execution refs because the slice does not launch generic preparation executions yet.

- [ ] **Step 3: Add acceptance tests**

Cover:

- prepared brief survives acceptance into durable mission data
- proposal disappears from the active proposed lane
- proposal dismiss path also uses the explicit proposal service
- no durable mission row is created before acceptance

Run:

```bash
go test ./internal/cli/cockpit/pages -run 'TestMissionControl(Accept|Proposal)' -count=1
```

Expected:

```text
ok
```

### Task 7: Update Public Docs and Complete OpenSpec

**Owner:** Worker A

**Files:**
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

- [ ] **Step 1: Audit landed Slice 3 behavior before editing docs**

Verify in code:

- transient proposal registry exists
- learning suggestions are the only active producer
- prepared brief generation is deterministic
- proposal acceptance creates durable mission rows while preserving prepared context

- [ ] **Step 2: Update docs**

Document only landed behavior:

- proactive proposals exist
- learning suggestions are the active producer in this slice
- low-risk deterministic preparation exists
- librarian/runtime-failure proposals are not yet enabled

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

- `openspec validate proactive-agent-behavior-slice-three --strict`
- archive the change

### Task 8: Final Verification and Slice Review

**Owner:** Main agent

- [ ] **Step 1: Run focused suites**

Run:

```bash
go test ./internal/proposal ./internal/app ./internal/cli/cockpit ./internal/cli/cockpit/pages ./cmd/lango -count=1
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

- [ ] **Step 3: Run final Slice 3 review**

Before claiming Slice 3 complete:

- request final spec-compliance review
- request final code-quality review
- fix all Critical and Major findings

- [ ] **Step 4: Archive and record final state**

After review and verification succeed:

- archive the Slice 3 change
- confirm main specs were updated
- record the final commit SHA in the work log
