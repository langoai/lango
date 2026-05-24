# Cockpit Selection Preservation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the landed cockpit dead-letter page so reload-triggering actions preserve the current selection whenever the selected transaction still exists in the refreshed result set.

**Architecture:** Keep the existing cockpit dead-letter page, filter bar, retry action, and reload paths. Unify the page-local reload policy so that apply, reset, and retry-success refresh all attempt to preserve the current `selected_transaction_receipt_id`. If preservation fails, reuse deterministic fallback behavior: first-row selection when results exist, or clear selection/detail when the refreshed result set is empty.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Unify reload behavior around optional selected-ID preservation.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover selection preservation on apply, reset, and retry success.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe unified selection preservation as landed.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark selection preservation as landed and narrow the remaining cockpit UX work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-cockpit-selection-preservation/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Unify Dead-Letter Page Reload Semantics Around Selection Preservation

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- `Enter` apply preserves the current selected transaction when it remains in the filtered result set
- `Ctrl+R` reset preserves the current selected transaction when it remains in the reset result set
- retry success refresh preserves the current selected transaction when it remains in the refreshed backlog
- when the selected transaction disappears, the page falls back to the first row
- when the refreshed result set is empty, selection and detail are cleared

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement unified selection preservation**

Update the page so that:

- apply reload attempts to preserve `selectedID`
- reset reload attempts to preserve `selectedID`
- retry-success refresh attempts to preserve `selectedID`
- fallback remains deterministic:
  - first row if results exist
  - clear selection/detail if empty

- [ ] **Step 4: Re-run the focused cockpit-page tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the cockpit page slice**

Run:

```bash
git add internal/cli/cockpit/pages/deadletters.go internal/cli/cockpit/pages/deadletters_test.go
git -c commit.gpgsign=false commit -m "feat: preserve cockpit dead-letter selection"
```

### Task 2: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-cockpit-selection-preservation/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- selection preservation across apply, reset, and retry success refresh
- first-row fallback when the selected transaction disappears
- selection/detail clear when the refreshed result set is empty

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks selection preservation as landed and narrows the remaining cockpit UX work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed selection-preservation slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-cockpit-selection-preservation/proposal.md`
- `openspec/changes/archive/2026-04-25-cockpit-selection-preservation/design.md`
- `openspec/changes/archive/2026-04-25-cockpit-selection-preservation/tasks.md`
- `openspec/changes/archive/2026-04-25-cockpit-selection-preservation/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-cockpit-selection-preservation
git -c commit.gpgsign=false commit -m "specs: archive cockpit selection preservation"
```

## Self-Review

- Spec coverage:
  - preservation model: Task 1
  - fallback model: Task 1
  - apply/reset/retry refresh semantics: Task 1
  - docs/OpenSpec truth alignment: Task 2
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no empty-state transition messaging
  - no stale-detail banners
  - no per-action visual diff
  - no selection history
