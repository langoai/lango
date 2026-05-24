# Dead-Letter CLI Actor / Time Filtering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the landed dead-letter CLI list command with manual-replay-actor and dead-letter-time filters while keeping the existing list/detail/retry command split intact.

**Architecture:** Keep the existing `lango status dead-letters` command and current dead-letter list bridge. Add three new list-only flags: `--manual-replay-actor`, `--dead-lettered-after`, and `--dead-lettered-before`. Reuse the existing dead-letter list read path, validate RFC3339 time inputs in the CLI, and leave the detail and retry commands unchanged.

**Tech Stack:** Go, Cobra CLI, `internal/cli/status`, existing dead-letter read bridge, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/status/status.go`
  - Add actor/time flags, validation, and forwarding.
- Modify: `internal/cli/status/status_test.go`
  - Cover valid forwarding and invalid RFC3339 rejection.
- Modify: `docs/cli/status.md`
  - Document the new list-command flags.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Mark CLI actor/time filtering as landed.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow remaining CLI filter work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-actor-time-filtering/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend `lango status dead-letters` with Actor / Time Flags

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write the failing CLI tests**

Add coverage for:

- valid `--manual-replay-actor`
- valid `--dead-lettered-after`
- valid `--dead-lettered-before`
- forwarding into the dead-letter list bridge
- invalid RFC3339 rejection for both time flags

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement actor / time filtering**

Extend `lango status dead-letters` so that it:

- accepts `--manual-replay-actor`
- accepts `--dead-lettered-after`
- accepts `--dead-lettered-before`
- validates both time flags as RFC3339
- forwards all three values into the existing dead-letter list bridge

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the CLI slice**

Run:

```bash
git add internal/cli/status/status.go internal/cli/status/status_test.go
git -c commit.gpgsign=false commit -m "feat: add dead letter cli actor time filters"
```

### Task 2: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/cli/status.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-actor-time-filtering/**`

- [ ] **Step 1: Update CLI/public docs**

Document:

- `--manual-replay-actor`
- `--dead-lettered-after`
- `--dead-lettered-before`
- RFC3339 requirement for time flags

- [ ] **Step 2: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed CLI actor/time-filtering slice.

- [ ] **Step 3: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-dead-letter-cli-actor-time-filtering/proposal.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-actor-time-filtering/design.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-actor-time-filtering/tasks.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-actor-time-filtering/specs/docs-only/spec.md`

- [ ] **Step 4: Run full verification**

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

- [ ] **Step 5: Commit the docs/OpenSpec slice**

Run:

```bash
git add docs/cli/status.md docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-dead-letter-cli-actor-time-filtering
git -c commit.gpgsign=false commit -m "specs: archive dead letter cli actor time filtering"
```

## Self-Review

- Spec coverage:
  - command surface extension: Task 1
  - flag model: Task 1
  - validation model: Task 1
  - data source reuse: Task 1
  - docs/OpenSpec truth alignment: Task 2
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no detail-command changes
  - no retry-command changes
  - no reason/dispatch filters
  - no `any_match_family`
  - no after/before ordering validation
