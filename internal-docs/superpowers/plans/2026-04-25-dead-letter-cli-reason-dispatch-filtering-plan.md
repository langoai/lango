# Dead-Letter CLI Reason / Dispatch Filtering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the landed dead-letter CLI list command with dead-letter reason and dispatch-reference filters while keeping the existing list/detail/retry command split intact.

**Architecture:** Keep the existing `lango status dead-letters` command and current dead-letter list bridge. Add two new list-only flags: `--dead-letter-reason-query` and `--latest-dispatch-reference`. Reuse the existing dead-letter list read path, pass both values through unchanged, and leave detail and retry commands untouched.

**Tech Stack:** Go, Cobra CLI, `internal/cli/status`, existing dead-letter read bridge, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/status/status.go`
  - Add reason/dispatch flags and forwarding.
- Modify: `internal/cli/status/status_test.go`
  - Cover valid forwarding and pass-through behavior.
- Modify: `docs/cli/status.md`
  - Document the new list-command flags.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Mark CLI reason/dispatch filtering as landed.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow remaining CLI filter work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-reason-dispatch-filtering/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend `lango status dead-letters` with Reason / Dispatch Flags

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write the failing CLI tests**

Add coverage for:

- valid `--dead-letter-reason-query`
- valid `--latest-dispatch-reference`
- forwarding into the dead-letter list bridge
- pass-through behavior without extra validation

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement reason / dispatch filtering**

Extend `lango status dead-letters` so that it:

- accepts `--dead-letter-reason-query`
- accepts `--latest-dispatch-reference`
- forwards both values into the existing dead-letter list bridge
- does not add extra validation

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
git -c commit.gpgsign=false commit -m "feat: add dead letter cli reason dispatch filters"
```

### Task 2: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/cli/status.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-reason-dispatch-filtering/**`

- [ ] **Step 1: Update CLI/public docs**

Document:

- `--dead-letter-reason-query`
- `--latest-dispatch-reference`
- pass-through semantics

- [ ] **Step 2: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed CLI reason/dispatch-filtering slice.

- [ ] **Step 3: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-dead-letter-cli-reason-dispatch-filtering/proposal.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-reason-dispatch-filtering/design.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-reason-dispatch-filtering/tasks.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-reason-dispatch-filtering/specs/docs-only/spec.md`

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
git add docs/cli/status.md docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-dead-letter-cli-reason-dispatch-filtering
git -c commit.gpgsign=false commit -m "specs: archive dead letter cli reason dispatch filtering"
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
  - no `any_match_family`
  - no extra format validation
