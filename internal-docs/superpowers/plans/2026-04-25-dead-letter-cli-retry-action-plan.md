# Dead-Letter CLI Retry Action Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first CLI recovery action to the landed dead-letter CLI surface so operators can request a retry from the status CLI.

**Architecture:** Keep the existing dead-letter CLI read surface and control-plane reuse. Add a new explicit command, `lango status dead-letter retry <transaction-receipt-id>`, that first reads the current detail status, checks `can_retry`, prompts for confirmation unless `--yes` is present, and then invokes the existing `retry_post_adjudication_execution` replay path. Keep result handling simple: no polling or action history in this slice.

**Tech Stack:** Go, Cobra CLI, `internal/cli/status`, existing dead-letter read/retry bridge, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/cli/status/status.go`
  - Add retry subcommand, precheck, confirm flow, and `--yes`.
- Modify: `internal/cli/status/status_test.go`
  - Cover retryable precheck, non-retryable rejection, confirm/yes behavior, and retry invocation.
- Modify: `docs/cli/status.md`
  - Document the retry command.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Mark the CLI retry action as landed.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow the remaining CLI operator workflow work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-retry-action/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Add `lango status dead-letter retry <transaction-receipt-id>`

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write the failing CLI tests**

Add coverage for:

- successful retry path when `can_retry=true`
- rejection before mutation when `can_retry=false`
- `--yes` bypasses the confirm prompt
- default interactive path requires confirmation
- retry bridge invocation happens only after precheck passes

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the retry subcommand**

Add:

- `lango status dead-letter retry <transaction-receipt-id>`
- existing detail read precheck
- `can_retry` guard
- confirm prompt
- `--yes`
- existing replay-path invocation

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the CLI action slice**

Run:

```bash
git add internal/cli/status/status.go internal/cli/status/status_test.go
git -c commit.gpgsign=false commit -m "feat: add dead letter cli retry action"
```

### Task 2: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/cli/status.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-dead-letter-cli-retry-action/**`

- [ ] **Step 1: Update CLI/public docs**

Document:

- `lango status dead-letter retry <transaction-receipt-id>`
- precheck semantics
- confirm prompt
- `--yes`

- [ ] **Step 2: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed CLI retry-action slice.

- [ ] **Step 3: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-25-dead-letter-cli-retry-action/proposal.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-retry-action/design.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-retry-action/tasks.md`
- `openspec/changes/archive/2026-04-25-dead-letter-cli-retry-action/specs/docs-only/spec.md`

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
git add docs/cli/status.md docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-dead-letter-cli-retry-action
git -c commit.gpgsign=false commit -m "specs: archive dead letter cli retry action"
```

## Self-Review

- Spec coverage:
  - command surface: Task 1
  - precheck model: Task 1
  - confirmation model: Task 1
  - control reuse: Task 1
  - docs/OpenSpec truth alignment: Task 2
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Scope check:
  - no polling
  - no action history
  - no bulk recovery
  - no other recovery actions
