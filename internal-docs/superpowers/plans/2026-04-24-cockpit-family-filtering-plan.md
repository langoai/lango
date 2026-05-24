# Cockpit Family Filtering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the landed cockpit dead-letter filter bar with `latest_status_subtype_family`, while keeping the existing draft/apply interaction and reload semantics intact.

**Architecture:** Keep the existing cockpit dead-letter page, filter bar, and dead-letter list bridge. Add one more filter control to the existing page-local draft state machine: a small family toggle with values `all / retry / manual-retry / dead-letter`. Reuse the existing backlog list surface by forwarding `latest_status_subtype_family` through the cockpit bridge. Keep the current `Enter` apply and first-row reset semantics unchanged.

**Tech Stack:** Go, Bubble Tea cockpit/TUI, `internal/cli/cockpit`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add `latest_status_subtype_family` draft state, rendering, and key handling.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover family toggle and apply semantics.
- Modify: `internal/cli/cockpit/deps.go`
  - Extend the dead-letter list bridge to forward `latest_status_subtype_family`.
- Modify: `internal/cli/cockpit/deps_test.go`
  - Cover family forwarding semantics.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe family filtering as landed in the cockpit filter bar.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark family filtering as landed and narrow the remaining cockpit filter work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-24-cockpit-family-filtering/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Cockpit Filter Bar with Family State

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write the failing cockpit-page tests**

Add tests covering:

- family draft toggle transitions:
  - `all`
  - `retry`
  - `manual-retry`
  - `dead-letter`
- family rendering in the filter bar
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

- [ ] **Step 3: Implement family filtering on the page**

Extend the page so that:

- add family draft state
- render family in the existing filter bar
- support enum toggle for family
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
git -c commit.gpgsign=false commit -m "feat: add cockpit family filter"
```

### Task 2: Extend the Dead-Letter Tool Bridge

**Files:**
- Modify: `internal/cli/cockpit/deps.go`
- Modify: `internal/cli/cockpit/deps_test.go`

- [ ] **Step 1: Write the failing bridge tests**

Add tests covering:

- `latest_status_subtype_family` forwarding to `list_dead_lettered_post_adjudication_executions`
- `all` family mode omits the filter

- [ ] **Step 2: Run the focused bridge tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/... -run 'TestDeadLetterToolBridge_|TestDeadLettersPage_' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement family forwarding**

Update the bridge so that:

- `latest_status_subtype_family` is forwarded to the existing list meta tool
- `all` omits the param
- existing query/adjudication/subtype/actor/time forwarding remains unchanged

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
git -c commit.gpgsign=false commit -m "feat: wire cockpit family filter bridge"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-cockpit-family-filtering/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- cockpit `latest_status_subtype_family` filter
- its enum values
- continued `Enter` apply semantics

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks family filtering as landed and narrows the remaining cockpit filter work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed family-filtering slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-cockpit-family-filtering/proposal.md`
- `openspec/changes/archive/2026-04-24-cockpit-family-filtering/design.md`
- `openspec/changes/archive/2026-04-24-cockpit-family-filtering/tasks.md`
- `openspec/changes/archive/2026-04-24-cockpit-family-filtering/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-cockpit-family-filtering
git -c commit.gpgsign=false commit -m "specs: archive cockpit family filtering"
```

## Self-Review

- Spec coverage:
  - filter surface extension: Task 1
  - interaction model: Task 1
  - reload / selection semantics: Task 1
  - data source reuse: Task 2
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no `any_match_family`
  - no live filtering
  - no selection preservation
  - no advanced family grouping
