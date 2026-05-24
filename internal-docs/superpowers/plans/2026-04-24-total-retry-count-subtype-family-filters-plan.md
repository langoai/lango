# Total Retry-Count / Subtype-Family Filters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the dead-letter backlog list with total retry-count filtering and latest subtype-family filtering while keeping the surface read-only and transaction-centered.

**Architecture:** Extend the existing `internal/postadjudicationstatus` read model rather than introducing a new store. Add `total_retry_count_min`, `total_retry_count_max`, and `latest_status_subtype_family` to the list query path, derive `total_retry_count` from relevant `post_adjudication_retry` events on the current submission trail, map the latest subtype to an operator-facing family, and expose both values on each backlog row. Update the existing list meta tool to pass through the new inputs and then truth-align docs and OpenSpec.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add total-count and family query inputs plus new row fields.
- Modify: `internal/postadjudicationstatus/service.go`
  - Derive `total_retry_count` and `latest_status_subtype_family`.
  - Add total-count range and family filtering.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover total retry count and family filter behavior.
- Modify: `internal/app/tools_meta.go`
  - Add the new list-tool parameters and pass them through.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover total-count and family filters through the meta-tool surface.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the new filters and row fields.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark total-count / family filters as landed and narrow the remaining backlog work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the enriched list-tool contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Create: `openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Dead-Letter Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- `total_retry_count_min`
- `total_retry_count_max`
- `latest_status_subtype_family`
- list entries exposing:
  - `total_retry_count`
  - `latest_status_subtype_family`

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement total-count and family filtering**

Extend the read model so that:

- `DeadLetterListOptions` accepts:
  - `TotalRetryCountMin`
  - `TotalRetryCountMax`
  - `LatestStatusSubtypeFamily`
- summary extraction derives:
  - total relevant retry lifecycle count
  - family from latest subtype
- list entries expose both values
- list filters apply:
  - total count range
  - family exact match

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
git -c commit.gpgsign=false commit -m "feat: add retry count family filters"
```

### Task 2: Upgrade the Read-Only Meta Tool Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `total_retry_count_min`
- `total_retry_count_max`
- `latest_status_subtype_family`
- list entries carrying:
  - `total_retry_count`
  - `latest_status_subtype_family`

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

- it accepts the new total-count and family filters
- it passes them into `DeadLetterListOptions`
- it returns the richer row fields while keeping the page shape unchanged

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
git -c commit.gpgsign=false commit -m "app: add retry count family params"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `total_retry_count_min`
- `total_retry_count_max`
- `latest_status_subtype_family`
- `total_retry_count`
- `latest_status_subtype_family`

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks total-count / family filters as landed work and narrows the remaining backlog-list work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed total-count and family filters.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters/proposal.md`
- `openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters/design.md`
- `openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters/tasks.md`
- `openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-total-retry-count-subtype-family-filters
git -c commit.gpgsign=false commit -m "specs: archive retry count family filters"
```

## Self-Review

- Spec coverage:
  - filter model: Task 1 + Task 2
  - family mapping model: Task 1
  - response shape: Task 1 + Task 2
  - evidence source: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Type consistency:
  - `total_retry_count_min`, `total_retry_count_max`, `latest_status_subtype_family`, and `total_retry_count` are used consistently across tasks
