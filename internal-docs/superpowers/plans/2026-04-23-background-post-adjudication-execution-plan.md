# Background Post-Adjudication Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first background post-adjudication execution slice that allows `adjudicate_escrow_dispute` to optionally enqueue a single async release/refund execution job after successful adjudication, without retries or dead-letter handling.

**Architecture:** Keep `internal/escrowadjudication` responsible only for canonical adjudication writes. Extend the `adjudicate_escrow_dispute` meta tool handler with `background_execute`, validate that it is mutually exclusive with `auto_execute`, enqueue an execution request onto the existing background task substrate, and have a worker reload canonical adjudication and reuse the existing `escrowrelease` / `escrowrefund` services. Update public docs and OpenSpec to describe this as an async convenience layer with explicit current limits.

**Tech Stack:** Go, `internal/app`, existing background manager/task substrate, `internal/escrowadjudication`, `internal/escrowrelease`, `internal/escrowrefund`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/app/tools_meta.go`
  - Extend `adjudicate_escrow_dispute` with `background_execute`.
  - Add mode validation and background dispatch receipt shape.
- Modify: `internal/app/tools_meta_escrowadjudication_test.go`
  - Add coverage for background dispatch mode, mode exclusivity, and dispatch receipt.
- Modify: background task wiring under `internal/app/` or existing automation wiring
  - Register a task payload/worker that calls the existing release/refund services.
- Create: `docs/architecture/background-post-adjudication-execution.md`
  - Public architecture/operator doc for the first async dispatch slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark background post-adjudication execution as landed and push retry/dead-letter work down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/background-post-adjudication-execution/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `adjudicate_escrow_dispute(background_execute=true)` contract.

### Task 1: Add Background Dispatch Mode to the Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_escrowadjudication_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Extend `internal/app/tools_meta_escrowadjudication_test.go` with tests covering:

- tool schema includes `background_execute`
- `background_execute=true` with `release` returns:
  - adjudication result
  - background dispatch receipt
- `background_execute=true` with `refund` returns:
  - adjudication result
  - background dispatch receipt
- `auto_execute=true` and `background_execute=true` together are rejected

- [ ] **Step 2: Run the adjudication meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesAdjudicateEscrowDispute|AdjudicateEscrowDispute_)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement background dispatch mode**

Update `internal/app/tools_meta.go` so that:

- `adjudicate_escrow_dispute` accepts optional `background_execute`
- `background_execute=true` is mutually exclusive with `auto_execute`
- adjudication still happens first
- after successful adjudication:
  - a background job is enqueued onto the existing background task substrate
  - a dispatch receipt is returned alongside the adjudication result
- worker execution is not returned synchronously

- [ ] **Step 4: Re-run the adjudication meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesAdjudicateEscrowDispute|AdjudicateEscrowDispute_)' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the background dispatch slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_meta_escrowadjudication_test.go
git -c commit.gpgsign=false commit -m "feat: add background post adjudication dispatch"
```

### Task 2: Add Worker Reuse of Existing Release/Refund Services

**Files:**
- Modify: existing background task wiring under `internal/app/` or related automation package
- Add tests alongside the chosen worker registration surface

- [ ] **Step 1: Write the failing worker tests**

Add tests covering:

- background payload carries `transaction_receipt_id` and expected outcome
- worker reloads canonical adjudication
- worker calls existing `escrowrelease` or `escrowrefund` service based on outcome
- worker failure is recorded as execution evidence without rewriting dispatch success

- [ ] **Step 2: Run the worker tests and verify they fail**

Run the smallest relevant package test command for the chosen worker surface.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the worker**

Add a background task or worker that:

- reads `transaction_receipt_id`
- reloads canonical adjudication outcome
- routes to existing `escrowrelease` or `escrowrefund` service
- reuses the same executor gates
- records worker failure evidence without retry

- [ ] **Step 4: Re-run the worker tests and verify they pass**

Run the same focused test command again.

Expected:

```text
ok
```

- [ ] **Step 5: Commit the worker slice**

Run:

```bash
git add <worker files and tests>
git -c commit.gpgsign=false commit -m "feat: add background post adjudication worker"
```

### Task 3: Publish Docs and Sync OpenSpec

**Files:**
- Create: `docs/architecture/background-post-adjudication-execution.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Create: `openspec/changes/archive/2026-04-23-background-post-adjudication-execution/**`

- [ ] **Step 1: Add the public architecture page**

Write `docs/architecture/background-post-adjudication-execution.md` describing:

- purpose and scope
- `background_execute=true` trigger
- dispatch receipt semantics
- worker execution semantics
- current limits

- [ ] **Step 2: Wire architecture landing, track, and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed background-execution slice truthfully.

- [ ] **Step 3: Sync OpenSpec main specs**

Update:

- `openspec/specs/project-docs/spec.md`
- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`

to reflect:

- the new public page
- the landed track status
- the `adjudicate_escrow_dispute(background_execute=true)` contract

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-background-post-adjudication-execution/proposal.md`
- `openspec/changes/archive/2026-04-23-background-post-adjudication-execution/design.md`
- `openspec/changes/archive/2026-04-23-background-post-adjudication-execution/tasks.md`
- delta spec stubs under `specs/`

and mark the change as complete.

- [ ] **Step 5: Run verification and commit docs/OpenSpec closeout**

Run:

```bash
.venv/bin/zensical build
go build ./...
go test ./...
```

Expected:

```text
all pass
```

Then commit:

```bash
git add docs/architecture/background-post-adjudication-execution.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-background-post-adjudication-execution
git -c commit.gpgsign=false commit -m "specs: archive background post adjudication execution"
```

---

## Sequencing Notes

- Task 1 first:
  - meta tool mode and dispatch receipt
- Task 2 second:
  - worker reuse of release/refund services
- Task 3 third:
  - public docs
  - spec sync
  - archive closeout

## Verification Checklist

- [ ] `go test ./internal/app -run 'Test(BuildMetaTools_IncludesAdjudicateEscrowDispute|AdjudicateEscrowDispute_)' -count=1`
- [ ] focused worker test command for the chosen background surface
- [ ] `.venv/bin/zensical build`
- [ ] `go build ./...`
- [ ] `go test ./...`

## Definition of Done

- `adjudicate_escrow_dispute` supports optional `background_execute`
- background dispatch is mutually exclusive with `auto_execute`
- background worker reuses existing release/refund services
- dispatch and worker failure semantics are clearly separated
- public docs and track page reflect the landed slice
- OpenSpec main specs are synced
- archived change exists under `openspec/changes/archive/2026-04-23-background-post-adjudication-execution`
- verification passes cleanly
