# Cockpit Reset / Clear Shortcuts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a single full-reset shortcut to the landed cockpit dead-letter page so operators can return to the default unfiltered backlog state with one keybinding.

**Architecture:** Keep the existing cockpit dead-letter page, filter bar, retry action, and bridge layer. Add a page-local `ctrl+r` reset action that clears every draft and applied filter back to defaults, clears retry confirm state, and reuses the current backlog/detail reload path. While retry is in the `running` state, the reset shortcut is ignored.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add `ctrl+r` handling and a full filter reset helper.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover full reset behavior, confirm-state clearing, and running-state no-op.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the landed reset shortcut.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark reset/clear shortcuts as landed and narrow the remaining cockpit UX work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-cockpit-reset-clear-shortcuts/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Add Full Filter Reset to the Cockpit Page

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- `ctrl+r` resets all draft filter fields to defaults
- `ctrl+r` resets all applied filter state to defaults
- `ctrl+r` clears retry confirm state
- reset triggers backlog reload, first-row reset, and detail reload
- reset is ignored while retry is running

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the reset shortcut**

Extend the page so that:

- `ctrl+r` resets all draft and applied filters to their default values
- retry confirm state is cleared during reset
- reset reuses the existing backlog/detail reload path
- reset keeps first-row selection semantics
- reset does nothing while retry is running

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
git -c commit.gpgsign=false commit -m "feat: add cockpit filter reset shortcut"
```

### Task 2: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-cockpit-reset-clear-shortcuts/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- cockpit `ctrl+r` full filter reset
- confirm-state clear on reset
- running-state no-op behavior

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks reset/clear shortcuts as landed and narrows the remaining cockpit UX work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed reset/clear-shortcuts slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-cockpit-reset-clear-shortcuts/proposal.md`
- `openspec/changes/archive/2026-04-25-cockpit-reset-clear-shortcuts/design.md`
- `openspec/changes/archive/2026-04-25-cockpit-reset-clear-shortcuts/tasks.md`
- `openspec/changes/archive/2026-04-25-cockpit-reset-clear-shortcuts/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-cockpit-reset-clear-shortcuts
git -c commit.gpgsign=false commit -m "specs: archive cockpit reset clear shortcuts"
```

## Self-Review

- Spec coverage:
  - reset model: Task 1
  - running-state guard: Task 1
  - reload / selection semantics: Task 1
  - docs/OpenSpec truth alignment: Task 2
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no per-field clear
  - no selection preservation
  - no retry cancellation
  - no reset confirmation prompt
