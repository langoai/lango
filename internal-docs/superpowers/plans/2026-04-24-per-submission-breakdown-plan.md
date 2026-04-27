# Per-Submission Breakdown Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the dead-letter backlog row with a compact per-submission breakdown that shows submission receipt ID, retry count, and any-match families for every submission in the transaction.

**Architecture:** Extend the existing `internal/postadjudicationstatus` read model rather than introducing a new store. Reuse the transaction-wide submission scan path, compute a compact summary for each submission, and expose the ordered `submission_breakdown` array directly on each backlog row. Do not add new filters in this slice. Update docs and OpenSpec to describe the new compact row-level breakdown.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add `SubmissionBreakdownItem` and `SubmissionBreakdown` row field.
- Modify: `internal/postadjudicationstatus/service.go`
  - Build compact per-submission summaries across all submissions in the current transaction.
  - Order them `oldest -> newest`.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover per-submission breakdown contents and ordering.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover `submission_breakdown` flowing through the list meta tool.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe `submission_breakdown` and its item shape.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark per-submission breakdown as landed and narrow the remaining breakdown work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the enriched row contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Create: `openspec/changes/archive/2026-04-24-per-submission-breakdown/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Dead-Letter Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- per-submission compact summary shape
- ordering `oldest -> newest`
- retry count per submission
- any-match families per submission

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the per-submission breakdown**

Extend the read model so that:

- add `SubmissionBreakdownItem`
  - `SubmissionReceiptID`
  - `RetryCount`
  - `AnyMatchFamilies`
- add `SubmissionBreakdown` to the backlog row
- reuse the transaction submission scan helper
- compute one compact summary per submission
- order the array `oldest -> newest`

- [ ] **Step 4: Re-run the status service tests and verify they pass**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the read-model slice**

Run:

```bash
git add internal/postadjudicationstatus/types.go internal/postadjudicationstatus/service.go internal/postadjudicationstatus/service_test.go
git -c commit.gpgsign=false commit -m "feat: add per submission breakdown"
```

### Task 2: Verify the Existing Meta Tool Surface Exposes the Breakdown

**Files:**
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool test**

Add tests covering:

- `submission_breakdown` present in list entries
- `submission_breakdown` ordered `oldest -> newest`
- per-item fields:
  - `submission_receipt_id`
  - `retry_count`
  - `any_match_families`

- [ ] **Step 2: Run the focused meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(ListDeadLetteredPostAdjudicationExecutions_|BuildMetaTools_IncludesPostAdjudicationStatus)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Keep the tool surface unchanged and verify the richer row shape flows through**

Do not add new filters or tools here. Only ensure the existing list tool returns the richer row shape from the read model.

- [ ] **Step 4: Re-run the focused meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'Test(ListDeadLetteredPostAdjudicationExecutions_|BuildMetaTools_IncludesPostAdjudicationStatus)' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta_postadjudicationstatus_test.go
git -c commit.gpgsign=false commit -m "test: cover per submission breakdown rows"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-per-submission-breakdown/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `submission_breakdown`
- per-item fields:
  - `submission_receipt_id`
  - `retry_count`
  - `any_match_families`

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks per-submission breakdown as landed work and narrows the remaining breakdown work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed per-submission breakdown slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-per-submission-breakdown/proposal.md`
- `openspec/changes/archive/2026-04-24-per-submission-breakdown/design.md`
- `openspec/changes/archive/2026-04-24-per-submission-breakdown/tasks.md`
- `openspec/changes/archive/2026-04-24-per-submission-breakdown/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-per-submission-breakdown/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-per-submission-breakdown
git -c commit.gpgsign=false commit -m "specs: archive per submission breakdown"
```

## Self-Review

- Spec coverage:
  - breakdown model: Task 1
  - ordering model: Task 1
  - response shape: Task 1 + Task 2
  - computation model: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Type consistency:
  - `SubmissionBreakdownItem` and `SubmissionBreakdown` are used consistently across the plan
