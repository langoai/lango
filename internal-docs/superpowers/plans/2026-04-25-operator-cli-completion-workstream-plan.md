# Operator CLI Completion Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the first operator-grade dead-letter CLI surface by landing the remaining core list filters and the first retry action under `lango status`.

**Architecture:** Keep the existing dead-letter CLI surface under `lango status`, reuse the current dead-letter read model and retry control plane, and batch the remaining operator-facing CLI completion into one workstream. The workstream extends the list command with actor/time and reason/dispatch filters, adds the explicit retry command with precheck and confirmation semantics, and truth-aligns CLI/public docs and OpenSpec in one closeout. Existing list/detail commands remain the source of truth for CLI read behavior; the retry action remains a thin wrapper over `retry_post_adjudication_execution`.

**Tech Stack:** Go, Cobra CLI, `internal/cli/status`, existing dead-letter read/retry bridge, Zensical docs, OpenSpec

---

## File Map

### Worker A: CLI Code / Tests

- Modify: `internal/cli/status/status.go`
  - Extend list flags and forwarding.
  - Add retry subcommand.
- Modify: `internal/cli/status/status_test.go`
  - Cover new filters, validation, precheck, confirm, and retry invocation.
- Modify: `internal/cli/status/render.go` only if new output helpers are needed.

### Worker B: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
  - Document the completed dead-letter CLI surface.
- Modify: `docs/cli/index.md`
  - Keep the CLI reference aligned if command examples or quick references need expansion.
- Modify: `README.md`
  - Update the short CLI command inventory if the new retry command is user-facing there.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the completed CLI operator surface.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark the workstream outcome as landed and narrow follow-on CLI work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirements.
- Create: `openspec/changes/archive/2026-04-25-operator-cli-completion-workstream/**`
  - Proposal, design, tasks, and delta specs.

## Task Breakdown

### Task 1: Finish Dead-Letter CLI List Filters

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI tests**

Add or extend coverage for:

- `--manual-replay-actor`
- `--dead-lettered-after`
- `--dead-lettered-before`
- `--dead-letter-reason-query`
- `--latest-dispatch-reference`
- forwarding into the dead-letter list bridge
- valid combinations with the already-landed:
  - `--query`
  - `--adjudication`
  - `--latest-status-subtype`
  - `--latest-status-subtype-family`

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the remaining list filters**

Extend `lango status dead-letters` so that it accepts and forwards:

- `--manual-replay-actor`
- `--dead-lettered-after`
- `--dead-lettered-before`
- `--dead-letter-reason-query`
- `--latest-dispatch-reference`

Keep:

- current list/detail split
- current `table` / `json`
- current subtype/family validation behavior

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

### Task 2: Add the CLI Retry Action

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI tests**

Add or extend coverage for:

- `lango status dead-letter retry <transaction-receipt-id>`
- status-detail precheck before mutation
- `can_retry=false` rejection
- default confirmation prompt behavior
- `--yes` prompt bypass
- retry invocation only after precheck passes
- default `table` result and optional `json` result

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the retry action**

Add:

- `lango status dead-letter retry <transaction-receipt-id>`
- existing detail read precheck
- `can_retry` guard
- confirm prompt
- `--yes`
- existing replay-path invocation

Keep result handling intentionally small:

- simple success output
- simple failure output
- optional `json`

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

### Task 3: Truth-Align Docs / README / OpenSpec

**Owner:** Worker B

**Files:**
- Modify: verified CLI/public docs
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-operator-cli-completion-workstream/**`

- [ ] **Step 1: Audit the final CLI surface before writing docs**

Verify the final command names, flags, output shape, and retry UX in code before editing docs.

- [ ] **Step 2: Update CLI/public docs**

Document the completed workstream surface:

- `lango status dead-letters`
  - query
  - adjudication
  - latest subtype
  - latest family
  - actor/time
  - reason/dispatch
- `lango status dead-letter <transaction-receipt-id>`
- `lango status dead-letter retry <transaction-receipt-id>`
  - confirm prompt
  - `--yes`

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed Operator CLI Completion workstream.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-25-operator-cli-completion-workstream/proposal.md`
- `openspec/changes/archive/2026-04-25-operator-cli-completion-workstream/design.md`
- `openspec/changes/archive/2026-04-25-operator-cli-completion-workstream/tasks.md`
- `openspec/changes/archive/2026-04-25-operator-cli-completion-workstream/specs/docs-only/spec.md`

### Task 4: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review Worker A changes**

Check:

- command naming
- flag validation
- bridge forwarding
- retry precheck semantics
- output consistency

- [ ] **Step 2: Review Worker B changes**

Check:

- docs match actual CLI behavior
- README additions only describe wired commands
- OpenSpec language matches the landed implementation

- [ ] **Step 3: Run full verification**

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

- [ ] **Step 4: Commit the integrated workstream**

Run:

```bash
git add internal/cli/status docs/cli docs/architecture README.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-25-operator-cli-completion-workstream
git -c commit.gpgsign=false commit -m "feat: complete operator dead letter cli"
```

## Parallelization Notes

- Worker A and Worker B can run in parallel because the primary write sets do not overlap.
- Avoid parallel edits to `README.md` from multiple workers.
- If Worker A changes any operator-visible command name or flag shape, Worker B must re-sync docs after that change lands.

## Self-Review

- Scope check:
  - no `any_match_family`
  - no polling / result follow-up UX
  - no bulk recovery
  - no dedicated background-task CLI browsing
- Contract check:
  - read surfaces continue to reuse existing dead-letter list/detail paths
  - retry continues to reuse `retry_post_adjudication_execution`
- Success criteria:
  - status CLI supports the targeted richer list filters
  - status CLI supports the first retry action
  - docs/OpenSpec describe the completed CLI surface accurately
