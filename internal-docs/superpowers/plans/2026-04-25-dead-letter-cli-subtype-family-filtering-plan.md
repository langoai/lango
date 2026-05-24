# Dead-Letter CLI Subtype / Family Filtering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the landed dead-letter CLI list command with latest-subtype and latest-family filters while keeping the existing list/detail command split and output model intact.

**Architecture:** Keep the existing dead-letter CLI surface and current bridge/data-source reuse. Extend only `lango status dead-letters` with two validated string flags: `--latest-status-subtype` and `--latest-status-subtype-family`. Forward them through the existing dead-letter list bridge into the current meta-tool-backed read model. Leave the detail command unchanged.

**Tech Stack:** Go, Cobra CLI, `internal/cli/status`, existing dead-letter read bridge, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/status/status.go`
  - Add subtype/family flags, validation, and forwarding to the dead-letter list bridge.
- Modify: `internal/cli/status/status_test.go`
  - Cover valid values, invalid values, and forwarding behavior.
- Modify: `docs/cli/status.md`
  - Document the new list-command filters.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Mark CLI subtype/latest-family filtering as landed.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow the remaining CLI filter work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-subtype-family-filtering/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend `lango status dead-letters` with Subtype / Family Flags

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write the failing CLI tests**

Add coverage for:

- valid `--latest-status-subtype`
- valid `--latest-status-subtype-family`
- forwarding into the dead-letter list bridge
- invalid subtype rejection
- invalid family rejection

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement subtype / family filtering**

Extend `lango status dead-letters` so that it:

- accepts `--latest-status-subtype`
- accepts `--latest-status-subtype-family`
- validates:
  - subtype:
    - `retry-scheduled`
    - `manual-retry-requested`
    - `dead-lettered`
  - family:
    - `retry`
    - `manual-retry`
    - `dead-letter`
- forwards validated values into the existing dead-letter list bridge

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
git -c commit.gpgsign=false commit -m "feat: add dead letter cli subtype filters"
```

### Task 2: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/cli/status.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-subtype-family-filtering/**`

- [ ] **Step 1: Update CLI/public docs**

Document:

- `--latest-status-subtype`
- `--latest-status-subtype-family`
- allowed values

- [ ] **Step 2: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed CLI subtype/family-filtering slice.

- [ ] **Step 3: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-dead-letter-cli-subtype-family-filtering/proposal.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-subtype-family-filtering/design.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-subtype-family-filtering/tasks.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-subtype-family-filtering/specs/docs-only/spec.md`

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
git add docs/cli/status.md docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-dead-letter-cli-subtype-family-filtering
git -c commit.gpgsign=false commit -m "specs: archive dead letter cli subtype family filtering"
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
  - no `any_match_family`
  - no actor/time filters
  - no reason/dispatch filters
  - no detail command changes
  - no CLI recovery actions
