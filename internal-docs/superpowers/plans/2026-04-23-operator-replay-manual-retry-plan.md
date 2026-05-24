# Operator Replay / Manual Retry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first operator replay / manual retry slice that allows a dead-lettered background post-adjudication execution to be re-enqueued through the existing background post-adjudication path, without clearing prior dead-letter evidence.

**Architecture:** Add a small replay service that accepts `transaction_receipt_id`, checks for dead-letter evidence plus canonical adjudication, appends `manual retry requested` evidence, and reuses the existing background post-adjudication dispatch path. Keep canonical adjudication unchanged. Update public docs and OpenSpec to describe the new operator-facing replay path.

**Tech Stack:** Go, `internal/app`, existing background post-adjudication dispatch path, `internal/receipts`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/postadjudicationreplay/types.go`
  - Request/result types and replay gate failure model.
- Create: `internal/postadjudicationreplay/service.go`
  - Replay gate, evidence append, and background dispatch reuse.
- Create: `internal/postadjudicationreplay/service_test.go`
  - Focused tests for dead-letter gating, success, and failure.
- Modify: `internal/receipts/types.go`
  - Add manual retry evidence request types if needed.
- Modify: `internal/receipts/store.go`
  - Add helper to append `manual retry requested` evidence.
- Modify: `internal/receipts/store_test.go`
  - Cover append-only replay evidence and state non-mutation.
- Modify: `internal/app/tools_meta.go`
  - Add `retry_post_adjudication_execution` meta tool.
- Modify: `internal/app/tools_parity_test.go`
  - Extend parity expectations.
- Create: `internal/app/tools_meta_postadjudicationreplay_test.go`
  - Meta-tool coverage for replay.
- Create: `docs/architecture/operator-replay-manual-retry.md`
  - Public architecture/operator doc for the first replay slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark manual retry as landed and push replay-policy follow-ons down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/archive/2026-04-23-operator-replay-manual-retry/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `retry_post_adjudication_execution` contract.

### Task 1: Add Replay Service

**Files:**
- Create: `internal/postadjudicationreplay/types.go`
- Create: `internal/postadjudicationreplay/service.go`
- Create: `internal/postadjudicationreplay/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- replay denied when transaction receipt is missing
- replay denied when dead-letter evidence is missing
- replay denied when canonical adjudication is missing
- replay success returns adjudication snapshot plus new dispatch receipt
- replay failure keeps canonical state unchanged

- [ ] **Step 2: Run the replay service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationreplay/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the replay service**

The service must:

- accept `transaction_receipt_id`
- load transaction receipt, current submission, and submission events
- require dead-letter evidence and canonical adjudication
- append `manual retry requested` evidence
- reuse existing background post-adjudication dispatch
- return:
  - canonical adjudication snapshot
  - new background dispatch receipt

- [ ] **Step 4: Re-run the replay service tests and verify they pass**

Run:

```bash
go test ./internal/postadjudicationreplay/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the replay service slice**

Run:

```bash
git add internal/postadjudicationreplay/types.go internal/postadjudicationreplay/service.go internal/postadjudicationreplay/service_test.go
git -c commit.gpgsign=false commit -m "feat: add post adjudication replay service"
```

### Task 2: Add Manual Retry Evidence Helpers

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add tests covering:

- `manual retry requested` appends evidence without mutating canonical adjudication
- dead-letter evidence remains after replay request

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

Add helper(s) for:

- `manual retry requested`

All helpers must keep:

- canonical adjudication intact
- append-only evidence in the current submission receipt trail

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
git -c commit.gpgsign=false commit -m "feat: add manual retry evidence"
```

### Task 3: Add `retry_post_adjudication_execution` Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_postadjudicationreplay_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `retry_post_adjudication_execution` is registered
- replay success returns canonical adjudication snapshot plus dispatch receipt
- replay fails when dead-letter evidence is missing

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesRetryPostAdjudicationExecution|RetryPostAdjudicationExecution_)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the meta tool**

Add:

- `retry_post_adjudication_execution`
- required input: `transaction_receipt_id`
- handler using the replay service
- thin receipt payload including:
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `escrow_reference`
  - `outcome`
  - `dispatch_reference`

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesRetryPostAdjudicationExecution|RetryPostAdjudicationExecution_)' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_postadjudicationreplay_test.go
git -c commit.gpgsign=false commit -m "app: add post adjudication replay meta tool"
```

### Task 4: Publish Docs and Sync OpenSpec

**Files:**
- Create: `docs/architecture/operator-replay-manual-retry.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Create: `openspec/changes/archive/2026-04-23-operator-replay-manual-retry/**`

- [ ] **Step 1: Add the public architecture page**

Write `docs/architecture/operator-replay-manual-retry.md` describing:

- purpose and scope
- replay gate
- replay model
- current limits

- [ ] **Step 2: Wire architecture landing, track, and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed replay slice truthfully.

- [ ] **Step 3: Sync OpenSpec main specs**

Update:

- `openspec/specs/project-docs/spec.md`
- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`

to reflect:

- the new public page
- the landed track status
- the `retry_post_adjudication_execution` contract

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-operator-replay-manual-retry/proposal.md`
- `openspec/changes/archive/2026-04-23-operator-replay-manual-retry/design.md`
- `openspec/changes/archive/2026-04-23-operator-replay-manual-retry/tasks.md`
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
git add docs/architecture/operator-replay-manual-retry.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-operator-replay-manual-retry
git -c commit.gpgsign=false commit -m "specs: archive operator replay manual retry"
```

---

## Sequencing Notes

- Task 1 first:
  - replay service
- Task 2 second:
  - manual retry evidence helpers
- Task 3 third:
  - meta tool
- Task 4 fourth:
  - public docs
  - spec sync
  - archive closeout

## Verification Checklist

- [ ] `go test ./internal/postadjudicationreplay/... -count=1`
- [ ] `go test ./internal/receipts/... -count=1`
- [ ] `go test ./internal/app -run 'Test(BuildMetaTools_IncludesRetryPostAdjudicationExecution|RetryPostAdjudicationExecution_)' -count=1`
- [ ] `.venv/bin/zensical build`
- [ ] `go build ./...`
- [ ] `go test ./...`

## Definition of Done

- dead-lettered post-adjudication execution can be manually replayed
- replay requires dead-letter evidence and canonical adjudication
- replay reuses the existing background dispatch path
- dead-letter evidence remains and `manual retry requested` is appended
- public docs and track page reflect the landed slice
- OpenSpec main specs are synced
- archived change exists under `openspec/changes/archive/2026-04-23-operator-replay-manual-retry`
- verification passes cleanly
