# Cross-Submission Lifecycle Grouping Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the dead-letter backlog list with transaction-global retry count and transaction-global any-match family aggregation while keeping the surface read-only and transaction-history based.

**Architecture:** Extend the existing `internal/postadjudicationstatus` read model rather than introducing a new store. For each backlog row, aggregate over all submission receipts that belong to the same transaction, count relevant retry lifecycle events, and collect a deduplicated family set. Expose those values on the row and add matching filters to the existing backlog list. Update the existing list meta tool to pass the new inputs through, then truth-align docs and OpenSpec.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add transaction-global row fields and query inputs.
- Modify: `internal/postadjudicationstatus/service.go`
  - Scan all submissions that belong to a transaction and aggregate retry lifecycle evidence.
  - Add transaction-global count/family filtering.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover transaction-global aggregation and filtering.
- Modify: `internal/app/tools_meta.go`
  - Add the new list-tool parameters and pass them through.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover transaction-global filters and row fields through the meta-tool surface.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe transaction-global count/family aggregation.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark cross-submission lifecycle grouping as landed and narrow the remaining grouping work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the list-tool contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Create: `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Dead-Letter Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- `transaction_global_total_retry_count`
- `transaction_global_any_match_families`
- `transaction_global_total_retry_count_min`
- `transaction_global_total_retry_count_max`
- `transaction_global_any_match_family`

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement transaction-global aggregation**

Extend the read model so that:

- it scans all submission receipts belonging to the current transaction
- it counts relevant retry lifecycle events across all those submission trails
- it derives a deduplicated transaction-global family set
- it exposes:
  - `TransactionGlobalTotalRetryCount`
  - `TransactionGlobalAnyMatchFamilies`
- it filters on:
  - `TransactionGlobalTotalRetryCountMin`
  - `TransactionGlobalTotalRetryCountMax`
  - `TransactionGlobalAnyMatchFamily`

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
git -c commit.gpgsign=false commit -m "feat: add transaction global retry grouping"
```

### Task 2: Upgrade the Read-Only Meta Tool Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `transaction_global_total_retry_count_min`
- `transaction_global_total_retry_count_max`
- `transaction_global_any_match_family`
- row fields:
  - `transaction_global_total_retry_count`
  - `transaction_global_any_match_families`

- [ ] **Step 2: Run the focused meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(ListDeadLetteredPostAdjudicationExecutions_|BuildMetaTools_IncludesPostAdjudicationStatus)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the list-tool parameter upgrade**

Update `list_dead_lettered_post_adjudication_executions` so that:

- it accepts the new transaction-global filters
- it passes them into `DeadLetterListOptions`
- it returns the new transaction-global row fields while keeping the page shape unchanged

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
git add internal/app/tools_meta.go internal/app/tools_meta_postadjudicationstatus_test.go
git -c commit.gpgsign=false commit -m "app: add transaction global backlog filters"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `transaction_global_total_retry_count`
- `transaction_global_any_match_families`
- `transaction_global_total_retry_count_min`
- `transaction_global_total_retry_count_max`
- `transaction_global_any_match_family`

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks cross-submission lifecycle grouping as landed work and narrows the remaining grouping work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed transaction-global aggregation slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping/proposal.md`
- `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping/design.md`
- `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping/tasks.md`
- `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-cross-submission-lifecycle-grouping
git -c commit.gpgsign=false commit -m "specs: archive cross submission grouping"
```

## Self-Review

- Spec coverage:
  - aggregation model: Task 1
  - filter model: Task 1 + Task 2
  - response shape: Task 1 + Task 2
  - computation model: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Type consistency:
  - `TransactionGlobalTotalRetryCount`, `TransactionGlobalAnyMatchFamilies`, and their filter names are used consistently across the plan
