# Retry / Dead-Letter Handling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first retry / dead-letter slice for background post-adjudication execution, with a bounded retry budget, exponential backoff, and terminal dead-letter evidence when retries are exhausted.

**Architecture:** Keep retry semantics scoped to the background post-adjudication execution path only. Add a retrying worker wrapper around the existing release/refund executor reuse path, track attempt count and next retry time in background-task metadata, append retry/dead-letter evidence to the current submission receipt trail, and leave canonical adjudication untouched. Do not generalize retry into the entire background manager yet.

**Tech Stack:** Go, `internal/app`, existing background task substrate, `internal/escrowrelease`, `internal/escrowrefund`, `internal/receipts`, Zensical docs, OpenSpec

---

## File Map

- Modify: background post-adjudication worker wiring under `internal/app/`
  - Add retry wrapper, retry metadata, and dead-letter terminalization.
- Modify: `internal/app/tools_meta_escrowadjudication_test.go`
  - Add worker-path coverage where appropriate if the worker remains observable through app tests.
- Modify: `internal/receipts/types.go`
  - Add retry/dead-letter evidence request types if needed.
- Modify: `internal/receipts/store.go`
  - Add helpers for retry scheduled / retry exhausted / dead-lettered evidence.
- Modify: `internal/receipts/store_test.go`
  - Cover append-only retry/dead-letter evidence and state non-mutation.
- Create: `docs/architecture/retry-dead-letter-handling.md`
  - Public architecture/operator doc for the first retry/dead-letter slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark retry/dead-letter as landed for the background post-adjudication path and push generic recovery work down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/archive/2026-04-23-retry-dead-letter-handling/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync any externally visible retry/dead-letter contract implications.

### Task 1: Add Retrying Background Worker Wrapper

**Files:**
- Modify: background post-adjudication worker wiring under `internal/app/`
- Add tests alongside the chosen worker surface

- [ ] **Step 1: Write the failing worker tests**

Add tests covering:

- retry identity uses `transaction_receipt_id + adjudication outcome`
- worker failure schedules retry with exponential backoff metadata
- retry stops after the third retry and marks dead-letter terminal failure
- successful retry preserves prior failure evidence and adds success evidence

- [ ] **Step 2: Run the worker tests and verify they fail**

Run the smallest relevant test command for the chosen worker surface.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement retry wrapper**

Add a retrying wrapper that:

- wraps only the post-adjudication background worker path
- applies maximum `3` retries
- uses exponential backoff
- tracks attempt count and next retry time in background-task metadata
- marks terminal dead-letter when retries are exhausted

- [ ] **Step 4: Re-run the worker tests and verify they pass**

Run the same focused test command again.

Expected:

```text
ok
```

- [ ] **Step 5: Commit the retry wrapper slice**

Run:

```bash
git add <worker files and tests>
git -c commit.gpgsign=false commit -m "feat: add retrying post adjudication worker"
```

### Task 2: Add Retry / Dead-Letter Evidence Helpers

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add tests covering:

- retry scheduled appends evidence without mutating canonical transaction state
- dead-lettered failure appends terminal evidence without mutating canonical adjudication
- eventual retry success preserves prior failure evidence

- [ ] **Step 2: Run the receipt tests and verify they fail**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the receipt helpers**

Add helpers for:

- retry scheduled
- retry failed
- retry exhausted / dead-lettered

All helpers must keep:

- canonical adjudication intact
- settlement progression intact
- append-only evidence in the submission receipt trail

- [ ] **Step 4: Re-run the receipt tests and verify they pass**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the receipt evidence slice**

Run:

```bash
git add internal/receipts/types.go internal/receipts/store.go internal/receipts/store_test.go
git -c commit.gpgsign=false commit -m "feat: add retry dead letter evidence"
```

### Task 3: Publish Docs and Sync OpenSpec

**Files:**
- Create: `docs/architecture/retry-dead-letter-handling.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Create: `openspec/changes/archive/2026-04-23-retry-dead-letter-handling/**`

- [ ] **Step 1: Add the public architecture page**

Write `docs/architecture/retry-dead-letter-handling.md` describing:

- purpose and scope
- retry identity
- retry policy
- dead-letter semantics
- current limits

- [ ] **Step 2: Wire architecture landing, track, and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed retry/dead-letter slice truthfully.

- [ ] **Step 3: Sync OpenSpec main specs**

Update:

- `openspec/specs/project-docs/spec.md`
- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`

to reflect:

- the new public page
- the landed track status
- any externally visible retry/dead-letter contract implications

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-retry-dead-letter-handling/proposal.md`
- `openspec/changes/archive/2026-04-23-retry-dead-letter-handling/design.md`
- `openspec/changes/archive/2026-04-23-retry-dead-letter-handling/tasks.md`
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
git add docs/architecture/retry-dead-letter-handling.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-retry-dead-letter-handling
git -c commit.gpgsign=false commit -m "specs: archive retry dead letter handling"
```

---

## Sequencing Notes

- Task 1 first:
  - retrying worker wrapper
- Task 2 second:
  - receipt evidence helpers
- Task 3 third:
  - public docs
  - spec sync
  - archive closeout

## Verification Checklist

- [ ] focused worker test command for the chosen background worker surface
- [ ] `go test ./internal/receipts/... -count=1`
- [ ] `.venv/bin/zensical build`
- [ ] `go build ./...`
- [ ] `go test ./...`

## Definition of Done

- post-adjudication background worker retries up to `3` times with exponential backoff
- exhausted retries become terminal dead-letter failure
- retry bookkeeping lives in background metadata, not canonical transaction state
- retry/dead-letter evidence is append-only in the submission receipt trail
- public docs and track page reflect the landed slice
- OpenSpec main specs are synced
- archived change exists under `openspec/changes/archive/2026-04-23-retry-dead-letter-handling`
- verification passes cleanly
