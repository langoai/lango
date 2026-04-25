# Cockpit Reason / Dispatch Filtering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the landed cockpit dead-letter filter bar with `dead_letter_reason_query` and `latest_dispatch_reference`, while keeping the existing draft/apply interaction and reload semantics intact.

**Architecture:** Keep the existing cockpit dead-letter page, filter bar, and dead-letter list bridge. Add two more text-input controls to the existing page-local draft state machine: one for dead-letter reason substring filtering and one for exact dispatch-reference lookup. Reuse the existing backlog list surface by forwarding both values through the cockpit bridge. Keep the current `Enter` apply and first-row reset semantics unchanged.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add reason and dispatch draft state, rendering, and key handling.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover reason/dispatch text editing and apply semantics.
- Modify: `internal/cli/cockpit/deps.go`
  - Extend the dead-letter list bridge to forward `dead_letter_reason_query` and `latest_dispatch_reference`.
- Modify: `internal/cli/cockpit/deps_test.go`
  - Cover reason/dispatch forwarding semantics.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe reason/dispatch filtering as landed in the cockpit filter bar.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark reason/dispatch filtering as landed and narrow the remaining cockpit filter work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-cockpit-reason-dispatch-filtering/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Cockpit Filter Bar with Reason / Dispatch State

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- dead-letter reason draft editing
- dispatch-reference draft editing
- reason/dispatch rendering in the filter bar
- `Enter` apply keeping the current reload behavior
- first-row reset semantics remaining unchanged

- [ ] **Step 2: Run the focused cockpit-page tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement reason / dispatch filtering on the page**

Extend the page so that:

- add reason and dispatch draft state
- render both fields in the existing filter bar
- support text editing for both fields
- keep `Enter` as the only apply trigger
- keep current first-row reset + detail reload semantics

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
git -c commit.gpgsign=false commit -m "feat: add cockpit reason dispatch filters"
```

### Task 2: Extend the Dead-Letter Tool Bridge

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/deps_test.go`

- [ ] **Step 1: Write the failing bridge tests**

Add tests covering:

- `dead_letter_reason_query` forwarding to `list_dead_lettered_post_adjudication_executions`
- `latest_dispatch_reference` forwarding to `list_dead_lettered_post_adjudication_executions`
- empty reason/dispatch inputs omit the params

- [ ] **Step 2: Run the focused bridge tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/... -run 'TestDeadLetterToolBridge_|TestDeadLettersPage_' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement reason / dispatch forwarding**

Update the bridge so that:

- `dead_letter_reason_query` is forwarded to the existing list meta tool
- `latest_dispatch_reference` is forwarded to the existing list meta tool
- empty values omit the params
- existing query/adjudication/subtype/family/actor/time forwarding remains unchanged

- [ ] **Step 4: Re-run the focused bridge tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/... -run 'TestDeadLetterToolBridge_|TestDeadLettersPage_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the bridge slice**

Run:

```bash
git add internal/cli/cockpit/deps.go internal/cli/cockpit/deps_test.go
git -c commit.gpgsign=false commit -m "feat: wire cockpit reason dispatch filters"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-cockpit-reason-dispatch-filtering/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- cockpit `dead_letter_reason_query` filtering
- cockpit `latest_dispatch_reference` filtering
- continued `Enter` apply semantics

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks reason/dispatch filtering as landed and narrows the remaining cockpit filter work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed reason/dispatch-filtering slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-cockpit-reason-dispatch-filtering/proposal.md`
- `openspec/changes/archive/2026-04-25-cockpit-reason-dispatch-filtering/design.md`
- `openspec/changes/archive/2026-04-25-cockpit-reason-dispatch-filtering/tasks.md`
- `openspec/changes/archive/2026-04-25-cockpit-reason-dispatch-filtering/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-cockpit-reason-dispatch-filtering
git -c commit.gpgsign=false commit -m "specs: archive cockpit reason dispatch filtering"
```

## Self-Review

- Spec coverage:
  - filter surface extension: Task 1
  - input model: Task 1
  - interaction model: Task 1
  - data source reuse: Task 2
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no reset / clear shortcuts
  - no selection preservation
  - no advanced filter modal
  - no result highlighting
