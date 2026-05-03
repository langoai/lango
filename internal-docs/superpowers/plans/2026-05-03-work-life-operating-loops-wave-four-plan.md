# Work And Life Operating Loops Wave 4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use `superpowers:subagent-driven-development` for each task. Every task must go through implementation, spec-compliance review, and code-quality review before the next task starts.

**Goal:** Land the first practical slice of Wave 4 by adding an operator-facing loop and agenda projection on top of existing durable missions, pending inquiries, dead letters, and deterministic follow-up signals, without pretending unsupported calendar or inbox integrations already exist.

**Architecture:** Wave 4 remains projection-first. It does not add a new durable loop database table in the first slice. Instead it adds a deterministic loop model and loop projector over existing sources, then surfaces those loops in Mission Control as a compact agenda/open-loops layer.

First-slice source contract:

- durable missions: enabled
- pending inquiries: enabled
- dead-letter backlog: enabled
- deterministic follow-up signals: enabled
- scheduled automation: `cron jobs` only
- workflow-run loops: deferred until a dedicated adapter exists

Source scoping contract:

- `missions`, `inquiries`, and deterministic follow-up signals are session-scoped
- `cron jobs` and `dead-letter backlog` are operator-global operational sources in the first slice
- operator-global loops may appear in every cockpit session until a later identity model exists

**Tech Stack:** Go, Bubble Tea, `internal/loopview`, `internal/app`, `internal/cli/cockpit`, OpenSpec, Zensical docs

## Scope Guardrails

- loops are a derived operator surface, not a replacement for `Mission`
- the first slice uses only real existing sources
- calendar, inbox, and external task-system integrations remain out of scope
- scheduled automation is `cron job` based only in the first slice; do not fabricate workflow or generic schedule state
- multi-mission clustering stays intentionally shallow in the first slice
- no new durable loop table in this slice

## File Map

### Worker A: OpenSpec / Docs / Public Truth

- Create: `openspec/changes/work-life-operating-loops-wave-four/proposal.md`
- Create: `openspec/changes/work-life-operating-loops-wave-four/design.md`
- Create: `openspec/changes/work-life-operating-loops-wave-four/tasks.md`
- Create: `openspec/changes/work-life-operating-loops-wave-four/specs/mission-control-tui/spec.md`
- Modify: `README.md`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`

### Worker B: Loop Projection Domain / App Wiring

- Create: `internal/loopview/types.go`
- Create: `internal/loopview/projector.go`
- Create: `internal/loopview/projector_test.go`
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/modules_test.go`
- Modify: `internal/app/app_test.go`

### Worker C: Mission Control Loop Surface

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

### Task 1: Create the Wave 4 OpenSpec Change

**Owner:** Worker A

**Files:**
- Create: `openspec/changes/work-life-operating-loops-wave-four/proposal.md`
- Create: `openspec/changes/work-life-operating-loops-wave-four/design.md`
- Create: `openspec/changes/work-life-operating-loops-wave-four/tasks.md`
- Create: `openspec/changes/work-life-operating-loops-wave-four/specs/mission-control-tui/spec.md`

- [ ] **Step 1: Create the change skeleton**

The change must capture:

- loop projection over existing sources
- deterministic agenda ordering
- initial real sources only
- no fake calendar/inbox/task-system integrations

- [ ] **Step 2: Add a Mission Control delta**

The `mission-control-tui` delta must state:

- Mission Control can surface loop/agenda rows in addition to missions/proposals/decisions
- loops are projected from real existing sources
- the first slice does not replace durable missions

- [ ] **Step 3: Validate the change**

Run:

```bash
openspec validate work-life-operating-loops-wave-four --strict
```

Expected:

```text
Change 'work-life-operating-loops-wave-four' is valid
```

### Task 2: Add the Loop Projection Domain

**Owner:** Worker B

**Files:**
- Create: `internal/loopview/types.go`
- Create: `internal/loopview/projector.go`
- Create: `internal/loopview/projector_test.go`

- [ ] **Step 1: Define loop and agenda types**

Add explicit types such as:

- `LoopKind`
- `LoopStatus`
- `LoopView`
- `AgendaView`

Required first-slice loop kinds:

- `mission_cluster`
- `inquiry`
- `dead_letter`
- `follow_up`
- `scheduled_automation` from cron jobs only in this slice

- [ ] **Step 2: Implement deterministic prioritization**

Initial priority order:

1. waiting-user
2. blocked
3. active
4. scheduled
5. needs-review
6. resolved

Ordering must remain deterministic and testable.

- [ ] **Step 3: Add projector helpers for real sources**

The first slice should support:

- durable mission rows
- pending inquiries
- dead-letter backlog
- deterministic follow-up signals from recent mission/proposal/session state
- scheduled automation from cron jobs only

Workflow runs are explicitly deferred in this slice.

Required source adapter shapes:

- `LoopMissionReader`
- `LoopInquiryReader`
- `LoopDeadLetterReader`
- `LoopCronReader`

- [ ] **Step 4: Add projection tests**

Cover:

- loop status ordering
- session scoping
- mission loops
- inquiry loops
- dead-letter loops
- follow-up loops
- no fabricated scheduled loops when the source is unavailable

Run:

```bash
go test ./internal/loopview -count=1
```

Expected:

```text
ok
```

### Task 3: Wire Loop Sources Into the App Surface

**Owner:** Worker B

**Files:**
- Modify: `internal/app/types.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/modules_test.go`
- Modify: `internal/app/app_test.go`

- [ ] **Step 1: Expose loop-relevant readers on App**

Expose only what Mission Control needs for the first slice:

- durable mission reader
- proposal registry access
- librarian inquiry store or a narrow inquiry reader
- a narrow dead-letter reader instead of raw cockpit-specific bridge behavior
- cron job reader as the scheduled automation source

- [ ] **Step 2: Keep unsupported sources absent**

If calendar/inbox/external task sources do not exist, the app surface must not imply they do.

- [ ] **Step 3: Add wiring tests**

Cover:

- loop-relevant app readers are populated when available
- absent sources remain nil rather than fabricated

Run:

```bash
go test ./internal/app -run 'Test.*Loop' -count=1
```

Expected:

```text
ok
```

### Task 4: Add Loop And Agenda Projection To Mission Control

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

- [ ] **Step 1: Add loop source dependencies to cockpit**

Cockpit deps should accept only the narrow read surfaces needed by the loop projector.

- [ ] **Step 2: Project loops and agenda**

Mission Control should derive:

- loop rows
- agenda ordering
- unresolved/open-loop counts

Durable missions remain visible first; loops are an additional coordination surface.

- [ ] **Step 3: Integrate loops into the page without another major layout rewrite**

The first slice should prefer a compact loop lane or agenda band rather than a whole new screen architecture.

Deterministic follow-up predicates for the first slice:

- accepted proposal with no active linked execution after `10m`
- mission in `done` state updated within `24h` and still needing review
- pending inquiry older than `24h`
- cron job whose most recent execution failed within `24h`
- dead-letter entry that remains retryable

No other heuristic follow-up generation is allowed in this slice.

Threshold handling rule:

- these thresholds are fixed constants inside `internal/loopview` in the first slice
- projector code must use an injectable/test-controlled clock so boundary tests stay deterministic

- [ ] **Step 4: Add Mission Control loop tests**

Cover:

- loop rows render from real sources
- agenda ordering is deterministic
- absent scheduled source does not fabricate loops
- degraded behavior stays truthful when optional sources are unavailable

Run:

```bash
go test ./internal/cli/cockpit ./internal/cli/cockpit/pages ./cmd/lango -run 'TestMissionControl(Loop|Agenda)|TestRunCockpit' -count=1
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

- [ ] **Step 1: Audit landed Wave 4 behavior before editing docs**

Verify in code:

- loop projection exists
- agenda ordering exists
- real first-slice sources only
- no fake integrations are implied

- [ ] **Step 2: Update docs**

Document only landed behavior:

- loops/agenda exist
- source coverage is limited to real existing sources
- external calendar/inbox/task integrations are not yet present

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

- `openspec validate work-life-operating-loops-wave-four --strict`
- archive the change

### Task 6: Final Verification And Wave Review

**Owner:** Main agent

- [ ] **Step 1: Run focused suites**

Run:

```bash
go test ./internal/loopview ./internal/app ./internal/cli/cockpit ./internal/cli/cockpit/pages ./cmd/lango -count=1
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

- [ ] **Step 3: Run final Wave 4 review**

Before claiming Wave 4 complete:

- request final spec-compliance review
- request final code-quality review
- fix all Critical and Major findings

- [ ] **Step 4: Archive and record final state**

After review and verification succeed:

- archive the Wave 4 change
- confirm main specs were updated
- record the final commit SHA in the work log
