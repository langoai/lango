# Cockpit Actor / Time Filtering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the landed cockpit dead-letter filter bar with `manual_replay_actor`, `dead_lettered_after`, and `dead_lettered_before`, while keeping the existing draft/apply interaction and reload semantics intact.

**Architecture:** Keep the existing cockpit dead-letter page, filter bar, and dead-letter list bridge. Add actor/time text inputs to the existing page-local draft state machine. Reuse the existing backlog list surface by forwarding `manual_replay_actor`, `dead_lettered_after`, and `dead_lettered_before` through the cockpit bridge. Keep the current `Enter` apply and first-row reset semantics unchanged.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add actor/time draft state, rendering, and key handling.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover actor/time editing and apply semantics.
- Modify: `internal/cli/cockpit/deps.go`
  - Extend the dead-letter list bridge to forward actor/time filters.
- Modify: `internal/cli/cockpit/deps_test.go`
  - Cover actor/time forwarding semantics.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe actor/time filtering as landed in the cockpit filter bar.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark actor/time filtering as landed and narrow the remaining cockpit filter work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-24-cockpit-actor-time-filtering/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Cockpit Filter Bar with Actor/Time State

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- `manual_replay_actor` draft editing
- `dead_lettered_after` draft editing
- `dead_lettered_before` draft editing
- actor/time rendering in the filter bar
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

- [ ] **Step 3: Implement actor/time filtering on the page**

Extend the page so that:

- add actor/time draft state
- render actor/time in the existing filter bar
- support text editing for:
  - `manual_replay_actor`
  - `dead_lettered_after`
  - `dead_lettered_before`
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
git -c commit.gpgsign=false commit -m "feat: add cockpit actor time filters"
```

### Task 2: Extend the Dead-Letter Tool Bridge

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/deps_test.go`

- [ ] **Step 1: Write the failing bridge tests**

Add tests covering:

- `manual_replay_actor` forwarding to `list_dead_lettered_post_adjudication_executions`
- `dead_lettered_after` forwarding
- `dead_lettered_before` forwarding
- empty values omit the params

- [ ] **Step 2: Run the focused bridge tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/... -run 'TestDeadLetterToolBridge_|TestDeadLettersPage_' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement actor/time forwarding**

Update the bridge so that:

- `manual_replay_actor` is forwarded to the existing list meta tool
- `dead_lettered_after` is forwarded to the existing list meta tool
- `dead_lettered_before` is forwarded to the existing list meta tool
- existing query/adjudication/subtype forwarding remains unchanged

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
git -c commit.gpgsign=false commit -m "feat: wire cockpit actor time filter bridge"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-cockpit-actor-time-filtering/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- cockpit `manual_replay_actor` filter
- cockpit `dead_lettered_after`
- cockpit `dead_lettered_before`
- continued `Enter` apply semantics

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks actor/time filtering as landed and narrows the remaining cockpit filter work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed actor/time-filtering slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-cockpit-actor-time-filtering/proposal.md`
- `openspec/changes/archive/2026-04-24-cockpit-actor-time-filtering/design.md`
- `openspec/changes/archive/2026-04-24-cockpit-actor-time-filtering/tasks.md`
- `openspec/changes/archive/2026-04-24-cockpit-actor-time-filtering/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-cockpit-actor-time-filtering
git -c commit.gpgsign=false commit -m "specs: archive cockpit actor time filtering"
```

## Self-Review

- Spec coverage:
  - filter surface extension: Task 1
  - input model: Task 1
  - interaction/apply model: Task 1
  - reload semantics: Task 1
  - bridge reuse: Task 2
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no date picker
  - no actor picker
  - no live filtering
  - no selection preservation
