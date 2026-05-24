# Transaction-Global Dominant Family Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the dead-letter backlog list with `transaction_global_dominant_family` as both a row field and a transaction-wide exact-match filter, while keeping the surface read-only and transaction-history based.

**Architecture:** Extend the existing `internal/postadjudicationstatus` read model rather than introducing a new store. Reuse the transaction-global submission aggregation path, derive a dominant family across all submissions in the current transaction using count-first selection with latest-event tie-break, expose it on each backlog row, and add an exact-match filter. Update the existing list meta tool to pass the filter through, then truth-align docs and OpenSpec.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add `TransactionGlobalDominantFamily` to the row shape and list options.
- Modify: `internal/postadjudicationstatus/service.go`
  - Extend the transaction-global aggregation helper to derive dominant family.
  - Add exact-match filtering for `TransactionGlobalDominantFamily`.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover transaction-global dominant-family selection, tie-break behavior, and filtering.
- Modify: `internal/app/tools_meta.go`
  - Add the new list-tool parameter and pass it through.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover `transaction_global_dominant_family` through the meta-tool surface.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe `transaction_global_dominant_family`.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark transaction-global dominant family as landed and narrow the remaining grouping work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the list-tool contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Create: `openspec/changes/archive/2026-04-24-transaction-global-dominant-family/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Dead-Letter Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- transaction-global dominant family selection by highest count
- transaction-global dominant family tie-break by latest relevant event
- `transaction_global_dominant_family` filtering

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement transaction-global dominant family**

Extend the read model so that:

- `DeadLetterBacklogEntry` includes `TransactionGlobalDominantFamily`
- `DeadLetterListOptions` accepts `TransactionGlobalDominantFamily`
- the transaction-global aggregation helper:
  - counts family hits across all submissions in the current transaction
  - uses highest count as the primary selector
  - uses latest relevant event family as the tie-break
- the list matcher applies exact matching for `TransactionGlobalDominantFamily`

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
git -c commit.gpgsign=false commit -m "feat: add transaction global dominant family"
```

### Task 2: Upgrade the Read-Only Meta Tool Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `transaction_global_dominant_family`
- list entries carrying `transaction_global_dominant_family`

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

- it accepts `transaction_global_dominant_family`
- it passes that value into `DeadLetterListOptions`
- it returns the new row field while keeping the page shape unchanged

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
git -c commit.gpgsign=false commit -m "app: add transaction global dominant filter"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-transaction-global-dominant-family/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `transaction_global_dominant_family`

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks transaction-global dominant family as landed work and narrows the remaining grouping work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed transaction-global dominant-family slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-transaction-global-dominant-family/proposal.md`
- `openspec/changes/archive/2026-04-24-transaction-global-dominant-family/design.md`
- `openspec/changes/archive/2026-04-24-transaction-global-dominant-family/tasks.md`
- `openspec/changes/archive/2026-04-24-transaction-global-dominant-family/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-transaction-global-dominant-family/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-transaction-global-dominant-family
git -c commit.gpgsign=false commit -m "specs: archive transaction global dominant family"
```

## Self-Review

- Spec coverage:
  - dominance model: Task 1
  - filter model: Task 1 + Task 2
  - response shape: Task 1 + Task 2
  - evidence source: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Type consistency:
  - `TransactionGlobalDominantFamily` is used consistently across the plan
