# Cockpit Dead-Letter Read Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a read-only cockpit master-detail surface for the post-adjudication dead-letter backlog, using the existing meta-tool-backed list and detail status contracts.

**Architecture:** Reuse the existing read model and meta tools rather than introducing a cockpit-specific backend. Add a new cockpit page that loads the dead-letter backlog table from `list_dead_lettered_post_adjudication_executions`, lets the operator select a row, and then loads the selected detail pane from `get_post_adjudication_execution_status`. Keep interaction minimal: selection only, no write actions or filter form in this slice.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/router.go`
  - Register a new cockpit page and sidebar metadata.
- Modify: `internal/cli/cockpit/cockpit_test.go`
  - Cover page registration / routing behavior if needed.
- Create: `internal/cli/cockpit/pages/deadletters.go`
  - Implement the master-detail read-only page.
- Create: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover list loading, selection, and detail-pane rendering.
- Modify: `internal/cli/cockpit/deps.go`
  - Add the minimal dependencies needed to fetch the backlog list and selected detail.
- Modify: `internal/app/modules.go`
  - Wire the new cockpit page into app initialization.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the landed cockpit read surface.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark the cockpit read surface as landed and narrow remaining operator-surface work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-24-cockpit-dead-letter-read-surface/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Add the Cockpit Dead-Letter Page

**Files:**
- Create: `internal/cli/cockpit/pages/deadletters.go`
- Create: `internal/cli/cockpit/pages/deadletters_test.go`
- Modify: `internal/cli/cockpit/router.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- initial backlog load into a table
- row selection updates the selected transaction
- selected transaction detail pane loads canonical detail status
- no-write/read-only behavior in the first slice

- [ ] **Step 2: Run the focused cockpit tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the master-detail cockpit page**

Implement a new page that:

- renders a dead-letter backlog table
- renders a selected transaction detail pane
- loads list data via the existing list surface
- loads detail data via the existing detail surface
- supports selection-only interaction

Also:

- register the page in the cockpit router/sidebar
- keep the page read-only

- [ ] **Step 4: Re-run the focused cockpit tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the cockpit page slice**

Run:

```bash
git add internal/cli/cockpit/router.go internal/cli/cockpit/cockpit_test.go internal/cli/cockpit/pages/deadletters.go internal/cli/cockpit/pages/deadletters_test.go
git -c commit.gpgsign=false commit -m "feat: add cockpit dead letter page"
```

### Task 2: Wire Data Dependencies Through App/Cockpit Initialization

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/app/modules.go`

- [ ] **Step 1: Write the failing integration wiring tests**

Add or extend tests covering:

- the cockpit can construct the dead-letter page with the required loaders
- the page receives list and detail dependencies from app wiring

- [ ] **Step 2: Run the focused wiring tests and verify they fail**

Run:

```bash
go test ./internal/app ./internal/cli/cockpit/... -run 'Test.*DeadLetter.*|Test.*Cockpit.*' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Wire the new page through deps/app initialization**

Update wiring so that:

- cockpit deps expose the minimal list/detail loader functions
- app initialization registers the page
- no new backend endpoints or write flows are introduced

- [ ] **Step 4: Re-run the focused wiring tests and verify they pass**

Run:

```bash
go test ./internal/app ./internal/cli/cockpit/... -run 'Test.*DeadLetter.*|Test.*Cockpit.*' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the wiring slice**

Run:

```bash
git add internal/cli/cockpit/deps.go internal/app/modules.go
git -c commit.gpgsign=false commit -m "feat: wire cockpit dead letter read surface"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-cockpit-dead-letter-read-surface/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- the cockpit master-detail read surface
- backlog table
- selected transaction detail pane
- selection-driven detail refresh

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks the cockpit read surface as landed work and narrows the remaining operator-surface work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed cockpit read surface.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-read-surface/proposal.md`
- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-read-surface/design.md`
- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-read-surface/tasks.md`
- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-read-surface/specs/docs-only/spec.md`

- [ ] **Step 5: Run full verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
```

Expected:

```text
ok
Build finished
```

- [ ] **Step 6: Commit the docs/OpenSpec slice**

Run:

```bash
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-cockpit-dead-letter-read-surface
git -c commit.gpgsign=false commit -m "specs: archive cockpit dead letter read surface"
```

## Self-Review

- Spec coverage:
  - surface model: Task 1
  - data sources: Task 1 + Task 2
  - interaction model: Task 1
  - state / refresh model: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - read-only only
  - no filter form
  - no replay/write actions
