# Operator Summary Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first operator-facing dead-letter summary surface with a dedicated CLI overview command.

**Architecture:** Keep the current dead-letter read model unchanged and build the summary surface as a thin CLI aggregation layer over `list_dead_lettered_post_adjudication_executions`. The workstream adds a new `lango status dead-letter-summary` command that computes global overview metrics in the CLI layer, then truth-aligns public docs and OpenSpec in one closeout. No new backend summary service or direct store reads are introduced.

**Tech Stack:** Go, Cobra CLI, `internal/cli/status`, existing dead-letter list bridge, Zensical docs, OpenSpec

---

## File Map

### Worker A: CLI Code / Tests

- Modify: `internal/cli/status/status.go`
  - Add the `dead-letter-summary` subcommand and summary aggregation logic.
- Modify: `internal/cli/status/status_test.go`
  - Cover summary aggregation, output, and command wiring.
- Modify: `internal/cli/status/render.go` only if a small shared summary renderer is helpful.

### Worker B: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
  - Document the new summary command and its output semantics.
- Modify: `docs/cli/index.md`
  - Add the new summary command to the quick reference surface.
- Modify: `README.md`
  - Update the short command inventory if `status` subcommands are listed there.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the new overview surface as landed work.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow the remaining operator-summary follow-on work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirements for the new summary surface.
- Create: `openspec/changes/archive/2026-04-25-operator-summary-workstream/**`
  - Proposal, design, tasks, and delta specs.

## Task Breakdown

### Task 1: Add the Dead-Letter Summary CLI Surface

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI tests**

Add or extend coverage for:

- `lango status dead-letter-summary`
- summary aggregation over the existing dead-letter backlog
- default `table` output
- optional `json` output
- summary fields:
  - `total_dead_letters`
  - `retryable_count`
  - `by_adjudication`
  - `by_latest_family`

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the summary command**

Add:

- `lango status dead-letter-summary`

Implementation rules:

- reuse the existing dead-letter list bridge
- aggregate summary values in the CLI layer
- keep the first slice to:
  - total dead letters
  - retryable count
  - by adjudication
  - by latest family
- support `table` and `json`

Keep unchanged:

- existing list/detail/retry commands
- existing dead-letter backend surfaces

- [ ] **Step 4: Re-run the focused status CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
ok
```

### Task 2: Truth-Align Docs / README / OpenSpec

**Owner:** Worker B

**Files:**
- Modify: verified CLI/public docs
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-25-operator-summary-workstream/**`

- [ ] **Step 1: Audit the final summary command before writing docs**

Verify:

- command name
- output shape
- summary fields
- whether README command inventory needs updating

- [ ] **Step 2: Update CLI/public docs**

Document:

- `lango status dead-letter-summary`
- default `table`
- optional `json`
- global overview fields:
  - total dead letters
  - retryable count
  - by adjudication
  - by latest family

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed Operator Summary workstream slice.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-25-operator-summary-workstream/proposal.md`
- `openspec/changes/archive/2026-04-25-operator-summary-workstream/design.md`
- `openspec/changes/archive/2026-04-25-operator-summary-workstream/tasks.md`
- `openspec/changes/archive/2026-04-25-operator-summary-workstream/specs/docs-only/spec.md`

### Task 3: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review Worker A changes**

Check:

- command naming
- aggregation correctness
- output consistency
- no regression in existing `status` dead-letter subcommands

- [ ] **Step 2: Review Worker B changes**

Check:

- docs match actual command behavior
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
