# Cockpit Confirm / Refresh Recovery UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the landed cockpit `Retry` action with inline confirm semantics and success-path backlog/detail refresh, while keeping the existing replay backend path unchanged.

**Architecture:** Keep the existing cockpit dead-letter page, filter bar, and replay bridge. Add page-local confirm state around the `Retry` action. First `r` enters confirm state, second `r` performs the replay, and successful replay triggers backlog reload plus selected-detail reload. Clear confirm state on `Esc`, row selection change, and filter apply/change. Do not introduce modal confirms, auto-timeouts, or richer action history.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, `cmd/lango`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add inline confirm state and success-path refresh semantics around the retry action.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover first/second `r`, confirm reset, and success refresh behavior.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the inline confirm and success refresh UX.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark confirm/refresh recovery UX as landed and narrow remaining cockpit recovery work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-24-cockpit-confirm-refresh-recovery-ux/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Add Inline Confirm State to the Cockpit Retry Action

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- first `r` enters confirm state
- second `r` invokes the retry action
- `Esc` clears confirm state
- selection change clears confirm state
- filter apply/change clears confirm state

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement inline confirm semantics**

Extend the dead-letter page so that:

- first `r` enters confirm state
- second `r` invokes the existing retry action
- detail pane renders confirm hint text
- `Esc` resets confirm
- selection change resets confirm
- filter change/apply resets confirm

- [ ] **Step 4: Re-run the focused cockpit-page tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the confirm-state slice**

Run:

```bash
git add internal/cli/cockpit/pages/deadletters.go internal/cli/cockpit/pages/deadletters_test.go
git -c commit.gpgsign=false commit -m "feat: add cockpit retry confirm state"
```

### Task 2: Refresh Backlog and Detail After Successful Retry

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing success-refresh tests**

Add tests covering:

- successful retry triggers backlog reload
- successful retry triggers selected detail reload
- failure does not auto-refresh data

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement success-path refresh**

Update the page so that after a successful replay:

- reload the backlog
- reload the selected transaction detail
- keep status feedback simple

Do not add:

- confirm timeout
- background polling
- richer loading indicators

- [ ] **Step 4: Re-run the focused cockpit-page tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the success-refresh slice**

Run:

```bash
git add internal/cli/cockpit/pages/deadletters.go internal/cli/cockpit/pages/deadletters_test.go
git -c commit.gpgsign=false commit -m "feat: refresh cockpit retry success state"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-cockpit-confirm-refresh-recovery-ux/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- inline retry confirm
- confirm reset on escape/selection/filter change
- success-path backlog/detail refresh

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks confirm/refresh recovery UX as landed and narrows the remaining cockpit recovery work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed confirm/refresh slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-cockpit-confirm-refresh-recovery-ux/proposal.md`
- `openspec/changes/archive/2026-04-24-cockpit-confirm-refresh-recovery-ux/design.md`
- `openspec/changes/archive/2026-04-24-cockpit-confirm-refresh-recovery-ux/tasks.md`
- `openspec/changes/archive/2026-04-24-cockpit-confirm-refresh-recovery-ux/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-cockpit-confirm-refresh-recovery-ux
git -c commit.gpgsign=false commit -m "specs: archive cockpit confirm refresh recovery ux"
```

## Self-Review

- Spec coverage:
  - confirm model: Task 1
  - reset semantics: Task 1
  - success refresh model: Task 2
  - interaction / feedback model: Task 1 + Task 2
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no modal confirm
  - no auto-timeout reset
  - no action history
