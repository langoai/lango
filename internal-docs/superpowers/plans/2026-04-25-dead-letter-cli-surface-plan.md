# Dead-Letter CLI Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first dedicated dead-letter CLI surface so operators can inspect backlog rows and per-transaction detail from the non-interactive status CLI.

**Architecture:** Keep the existing dead-letter read model and meta-tool-backed status path. Extend the `status` CLI surface with two new commands: one for dead-letter backlog listing and one for per-transaction status detail. Reuse the existing dead-letter list/detail surfaces instead of introducing new backend paths. Start with default `table` output and optional `json`, and keep first-slice filters to `query` and `adjudication`.

**Tech Stack:** Go, Cobra-style CLI wiring in `internal/cli/status`, existing app/tool bridge layer, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/status/*`
  - Add `dead-letters` list command and `dead-letter` detail command.
- Modify: `cmd/lango/main.go` or current status command registration path if needed
  - Wire the new status subcommands into the CLI.
- Modify: downstream help/docs surfaces affected by the new commands
  - CLI docs
  - public docs
- Modify: `README.md` only if the current status command index or examples actually include the new surface after verification
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the landed CLI view.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark the higher-level CLI slice as landed.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-surface/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Add `lango status dead-letters`

**Files:**
- Modify: `internal/cli/status/*`
- Modify: relevant CLI tests

- [ ] **Step 1: Write the failing CLI tests**

Add coverage for:

- `lango status dead-letters`
- default `table` output
- optional `json` output
- `--query`
- `--adjudication`

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the dead-letter backlog list command**

Add:

- `lango status dead-letters`
- reuse `list_dead_lettered_post_adjudication_executions`
- support:
  - `--query`
  - `--adjudication`
  - `--output table|json`

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the list-command slice**

Run:

```bash
git add <status-cli-files>
git -c commit.gpgsign=false commit -m "feat: add dead letter status list command"
```

### Task 2: Add `lango status dead-letter <transaction-receipt-id>`

**Files:**
- Modify: `internal/cli/status/*`
- Modify: relevant CLI tests

- [ ] **Step 1: Write the failing CLI tests**

Add coverage for:

- `lango status dead-letter <transaction-receipt-id>`
- default `table` output
- optional `json` output
- detail includes the existing background-task bridge when present

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the dead-letter detail command**

Add:

- `lango status dead-letter <transaction-receipt-id>`
- reuse `get_post_adjudication_execution_status`
- support:
  - `--output table|json`

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the detail-command slice**

Run:

```bash
git add <status-cli-files>
git -c commit.gpgsign=false commit -m "feat: add dead letter status detail command"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: verified CLI/public docs that surface the new commands
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-surface/**`

- [ ] **Step 1: Audit the actual CLI wiring before documenting**

Verify the final command names, flags, and output behavior in code before editing docs.

- [ ] **Step 2: Update public docs**

Update docs to describe:

- `lango status dead-letters`
- `lango status dead-letter <transaction-receipt-id>`
- `table` default
- `json` support
- first-slice filters:
  - `query`
  - `adjudication`

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed CLI slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-dead-letter-cli-surface/proposal.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-surface/design.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-surface/tasks.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-surface/specs/docs-only/spec.md`

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
git add <verified-doc-files> openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-dead-letter-cli-surface
git -c commit.gpgsign=false commit -m "specs: archive dead letter cli surface"
```

## Self-Review

- Spec coverage:
  - command surface: Tasks 1-2
  - data source reuse: Tasks 1-2
  - output model: Tasks 1-2
  - filter model: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - implementation tasks intentionally use `<status-cli-files>` and `<verified-doc-files>` as commit-stage placeholders because exact touched files must be confirmed from the actual CLI wiring before final commit
- Scope check:
  - no richer CLI filters
  - no replay actions
  - no background-task browsing command
  - no plain-output polish
