# Richer Dead-Letter Summaries Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the first dead-letter summary surface with top 5 latest dead-letter reasons.

**Architecture:** Keep the current `lango status dead-letter-summary` command and extend it additively. The workstream reuses the existing dead-letter backlog read model, performs latest-reason aggregation in the CLI layer, and extends `table`/`json` output without adding a new backend summary service or a new command surface.

**Tech Stack:** Go, Cobra CLI, `internal/cli/status`, existing dead-letter list bridge, Zensical docs, OpenSpec

---

## File Map

### Worker A: CLI Code / Tests

- Modify: `internal/cli/status/status.go`
  - Extend summary result types and aggregation logic.
- Modify: `internal/cli/status/render.go`
  - Extend summary rendering with top-reasons section.
- Modify: `internal/cli/status/status_test.go`
  - Cover top-reason aggregation and output.

### Worker B: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
  - Document the richer summary output.
- Modify: `docs/cli/index.md`
  - Update the summary command description only if it needs richer wording.
- Modify: `README.md`
  - Update only if the summary command description appears in a public inventory.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe top latest dead-letter reasons as landed work.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow the remaining summary follow-on work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirements for richer summaries.
- Create: `openspec/changes/archive/2026-04-25-richer-dead-letter-summaries-workstream/**`
  - Proposal, design, tasks, and delta specs.

## Task Breakdown

### Task 1: Extend the Dead-Letter Summary CLI Surface

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/render.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI tests**

Add or extend coverage for:

- top 5 latest dead-letter reason aggregation
- additive extension of `lango status dead-letter-summary`
- default `table` output with a top-reasons section
- optional `json` output with `top_latest_dead_letter_reasons`
- item shape:
  - `reason`
  - `count`

- [ ] **Step 2: Run the focused status CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the richer summary extension**

Extend the existing summary command so that it:

- keeps:
  - `total_dead_letters`
  - `retryable_count`
  - `by_adjudication`
  - `by_latest_family`
- adds:
  - `top_latest_dead_letter_reasons`

Implementation rules:

- aggregate from `latest_dead_letter_reason`
- use top 5 reasons only
- keep the existing command name
- keep `table` / `json`
- do not add flags or new backend summary services

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
- Create: `openspec/changes/archive/2026-04-25-richer-dead-letter-summaries-workstream/**`

- [ ] **Step 1: Audit the final richer summary command before writing docs**

Verify:

- command name stays unchanged
- top-reasons section is actually present
- `json` includes the new array field
- README needs updating only if the command inventory mentions summary detail

- [ ] **Step 2: Update CLI/public docs**

Document:

- `lango status dead-letter-summary`
- existing overview fields
- new `top 5 latest dead-letter reasons`
- `table` / `json` output extension

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed richer dead-letter summary slice.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-25-richer-dead-letter-summaries-workstream/proposal.md`
- `openspec/changes/archive/2026-04-25-richer-dead-letter-summaries-workstream/design.md`
- `openspec/changes/archive/2026-04-25-richer-dead-letter-summaries-workstream/tasks.md`
- `openspec/changes/archive/2026-04-25-richer-dead-letter-summaries-workstream/specs/docs-only/spec.md`

### Task 3: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review Worker A changes**

Check:

- aggregation correctness
- top-5 ordering behavior
- output consistency
- no regression in existing summary/list/detail/retry commands

- [ ] **Step 2: Review Worker B changes**

Check:

- docs match the landed summary output
- README additions only describe actual surfaced behavior
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
