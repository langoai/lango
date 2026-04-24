# Cockpit Dead-Letter Filtering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the landed cockpit dead-letter master-detail page with a thin filter bar that supports `query` and `adjudication`, applies on `Enter`, reloads the filtered backlog, and resets selection to the first row.

**Architecture:** Keep the existing cockpit dead-letter page and existing list/detail status surfaces. Add local filter draft state to `pages/deadletters.go`, extend the list loader bridge to accept `query` and `adjudication`, and keep the detail pane semantics simple: after apply, reload backlog, reset to the first row, and reload detail for that row. Do not add live filtering, selection preservation, or write controls.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add filter draft state, input rendering, key handling, and apply semantics.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover filter editing, apply, reload, and first-row reset semantics.
- Modify: `internal/cli/cockpit/deps.go`
  - Extend the dead-letter list bridge to accept `query` and `adjudication`.
- Modify: `internal/cli/cockpit/deps_test.go`
  - Cover list bridge forwarding the new filter params.
- Modify: `cmd/lango/main.go`
  - Wire the page with the richer list loader if needed.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the cockpit filter bar as landed.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark thin cockpit filtering as landed and narrow remaining cockpit work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-24-cockpit-dead-letter-filtering/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Cockpit Page with a Thin Filter Bar

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- `query` draft editing
- `adjudication` toggle transitions:
  - `all`
  - `release`
  - `refund`
- `Enter` apply triggers backlog reload
- apply resets selection to the first row
- detail reload follows the new first row
- empty filtered result clears the selected detail

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the thin filter bar**

Extend the dead-letter page so that:

- add filter draft state:
  - `query`
  - `adjudication`
- render a small filter bar above the backlog table
- support text editing for `query`
- support enum toggle for `adjudication`
- apply the current draft on `Enter`
- after apply:
  - reload the filtered backlog
  - reset selection to the first row
  - reload detail for that row
- if no rows remain:
  - show empty backlog state
  - clear selected detail

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
git -c commit.gpgsign=false commit -m "feat: add cockpit dead letter filters"
```

### Task 2: Extend the Cockpit Tool Bridge for Filtered List Loading

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/deps_test.go`
- Modify: `cmd/lango/main.go`

- [ ] **Step 1: Write the failing bridge tests**

Add tests covering:

- `query` forwarding to `list_dead_lettered_post_adjudication_executions`
- `adjudication` forwarding to `list_dead_lettered_post_adjudication_executions`
- `all` adjudication mode omits the filter

- [ ] **Step 2: Run the focused bridge tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/... -run 'TestDeadLetterToolBridge_|TestDeadLettersPage_' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement filtered list loading**

Update the cockpit bridge so that:

- the list loader accepts `query` and `adjudication`
- those params are forwarded to the existing list meta tool
- detail loading stays unchanged
- main wiring keeps using the existing bridge/page construction path

- [ ] **Step 4: Re-run the focused bridge tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/... -run 'TestDeadLetterToolBridge_|TestDeadLettersPage_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the bridge/wiring slice**

Run:

```bash
git add internal/cli/cockpit/deps.go internal/cli/cockpit/deps_test.go cmd/lango/main.go
git -c commit.gpgsign=false commit -m "feat: wire cockpit dead letter filter bridge"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-cockpit-dead-letter-filtering/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- cockpit dead-letter filter bar
- `query`
- `adjudication`
- `Enter` apply
- first-row reset semantics

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks thin cockpit filtering as landed and narrows the remaining cockpit work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed cockpit filtering slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-filtering/proposal.md`
- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-filtering/design.md`
- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-filtering/tasks.md`
- `openspec/changes/archive/2026-04-24-cockpit-dead-letter-filtering/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-cockpit-dead-letter-filtering
git -c commit.gpgsign=false commit -m "specs: archive cockpit dead letter filtering"
```

## Self-Review

- Spec coverage:
  - filter surface model: Task 1
  - interaction model: Task 1
  - reload / selection semantics: Task 1
  - data source reuse: Task 2
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no live filtering
  - no selection preservation
  - no replay/write controls
