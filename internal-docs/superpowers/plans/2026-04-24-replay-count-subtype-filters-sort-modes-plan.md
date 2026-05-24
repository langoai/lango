# Replay-Count / Subtype Filters + Alternate Sort Modes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the dead-letter backlog list with subtype filtering, manual-replay-count filtering, and a small set of alternate sort modes while keeping the surface read-only and transaction-centered.

**Architecture:** Extend the existing `internal/postadjudicationstatus` read model rather than introducing a new store. Add `latest_status_subtype`, `manual_retry_count_min`, `manual_retry_count_max`, and `sort_by` to the existing list query path, derive `manual_retry_count` and `latest_manual_replay_at` from the submission receipt trail, and keep sorting direction fixed per mode. Update the existing backlog meta tool to pass through the new inputs and expose the new entry fields. Then truth-align docs and OpenSpec.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add subtype/count/sort inputs and new list-entry fields.
- Modify: `internal/postadjudicationstatus/service.go`
  - Derive `manual_retry_count`, `latest_manual_replay_at`, and `latest_status_subtype`.
  - Add subtype/count filtering and sort-mode switching.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover subtype filtering, manual retry count filtering, and sort-mode behavior.
- Modify: `internal/app/tools_meta.go`
  - Add the new list-tool parameters and pass them through.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover subtype/count filtering and `sort_by` through the meta-tool surface.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the new filters, sort modes, and row fields.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark subtype/count filters plus sort modes as landed and narrow the remaining backlog work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the enriched list-tool contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Create: `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Dead-Letter Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- `latest_status_subtype` exact-match filtering
- `manual_retry_count_min`
- `manual_retry_count_max`
- `sort_by=latest_dead_lettered_at`
- `sort_by=latest_retry_attempt`
- `sort_by=latest_manual_replay_at`
- list entries exposing:
  - `manual_retry_count`
  - `latest_manual_replay_at`
  - `latest_status_subtype`

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement subtype/count filters and sort modes**

Extend the read model so that:

- `DeadLetterListOptions` accepts:
  - `LatestStatusSubtype`
  - `ManualRetryCountMin`
  - `ManualRetryCountMax`
  - `SortBy`
- summary extraction derives:
  - `ManualRetryCount`
  - `LatestManualReplayAt`
  - `LatestStatusSubtype`
- list entries expose those values
- list sorting supports:
  - `latest_dead_lettered_at desc`
  - `latest_retry_attempt desc`
  - `latest_manual_replay_at desc`

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
git -c commit.gpgsign=false commit -m "feat: add backlog subtype and sort filters"
```

### Task 2: Upgrade the Read-Only Meta Tool Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `latest_status_subtype`
- `manual_retry_count_min`
- `manual_retry_count_max`
- `sort_by`
- list entries carrying:
  - `manual_retry_count`
  - `latest_manual_replay_at`
  - `latest_status_subtype`

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

- it accepts the new subtype/count/sort inputs
- it passes them into `DeadLetterListOptions`
- it returns the richer list-entry fields while keeping the page shape unchanged

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
git -c commit.gpgsign=false commit -m "app: add backlog subtype and sort params"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `latest_status_subtype`
- `manual_retry_count_min`
- `manual_retry_count_max`
- `sort_by`
- `manual_retry_count`
- `latest_manual_replay_at`
- `latest_status_subtype`

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks subtype/count filters plus sort modes as landed work and narrows the remaining backlog-list work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed subtype/count filters and sort modes.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes/proposal.md`
- `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes/design.md`
- `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes/tasks.md`
- `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-replay-count-subtype-filters-sort-modes
git -c commit.gpgsign=false commit -m "specs: archive backlog sort filters"
```

## Self-Review

- Spec coverage:
  - filter model: Task 1 + Task 2
  - sort model: Task 1 + Task 2
  - response shape: Task 1 + Task 2
  - evidence sources: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Type consistency:
  - `latest_status_subtype`, `manual_retry_count_min`, `manual_retry_count_max`, `sort_by`, `manual_retry_count`, and `latest_manual_replay_at` are used consistently across tasks
