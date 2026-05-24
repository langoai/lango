# Raw Background-Task Bridges Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend `get_post_adjudication_execution_status` so it keeps the existing receipts-backed canonical snapshot but also exposes a thin `latest_background_task` bridge containing the latest matching post-adjudication background task for the current transaction and adjudication outcome.

**Architecture:** Keep `internal/postadjudicationstatus` as the primary read model and add an optional background-task reader dependency. Select the latest matching task by `RetryKey` derived from `transaction_receipt_id + adjudication outcome`, project a thin snapshot (`task_id`, `status`, `attempt_count`, `next_retry_at`), and attach it only to the detail view. Do not expand the backlog list in this slice. Reuse `background.Manager.List()` rather than introducing a new store or cache.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/background`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add the optional `LatestBackgroundTask` response shape.
- Modify: `internal/postadjudicationstatus/service.go`
  - Accept an optional background-task reader.
  - Select and project the latest matching background task.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover matching, projection, and null semantics.
- Modify: `internal/app/tools_meta.go`
  - Wire the optional background-task reader into `get_post_adjudication_execution_status`.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover tool-level exposure of `latest_background_task`.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the detail-view raw background bridge.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark raw background-task bridges as landed and narrow remaining operator-surface work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the detail tool contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirement.
- Create: `openspec/changes/archive/2026-04-24-raw-background-task-bridges/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Status Read Model with a Thin Background Bridge

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing status-service tests**

Add tests covering:

- matching the latest background task by current transaction + current adjudication outcome
- projecting:
  - `task_id`
  - `status`
  - `attempt_count`
  - `next_retry_at`
- `latest_background_task = nil` when no matching task exists
- ignoring unrelated tasks for other transactions or outcomes

- [ ] **Step 2: Run the focused status-service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the thin background bridge**

Extend the read model so that:

- add a `LatestBackgroundTask` field to `TransactionStatus`
- add a small projected type with:
  - `TaskID`
  - `Status`
  - `AttemptCount`
  - `NextRetryAt`
- accept an optional background-task reader dependency
- read `background.Manager.List()`-style snapshots
- derive the match key from:
  - current `transaction_receipt_id`
  - current adjudication outcome
- select the latest matching task using the existing `RetryKey`
- return `nil` when no match exists

- [ ] **Step 4: Re-run the focused status-service tests and verify they pass**

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
git -c commit.gpgsign=false commit -m "feat: add raw background task bridge"
```

### Task 2: Wire the Detail Meta Tool to the Background Reader

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing tool-level tests**

Add tests covering:

- `get_post_adjudication_execution_status` includes `latest_background_task` when a matching task exists
- the exposed object contains:
  - `task_id`
  - `status`
  - `attempt_count`
  - `next_retry_at`
- the tool still returns the canonical detail payload when `latest_background_task` is `null`

- [ ] **Step 2: Run the focused tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(GetPostAdjudicationExecutionStatus_|BuildMetaTools_IncludesPostAdjudicationStatus)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Wire the optional background-task reader**

Update the tool wiring so that:

- the detail tool can receive an optional background-task reader
- existing call sites without a background manager still work
- runtime wiring can pass the real `background.Manager`
- no list-tool expansion happens in this slice

- [ ] **Step 4: Re-run the focused tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'Test(GetPostAdjudicationExecutionStatus_|BuildMetaTools_IncludesPostAdjudicationStatus)' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the tool-wiring slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_meta_postadjudicationstatus_test.go
git -c commit.gpgsign=false commit -m "feat: wire background task status bridge"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-raw-background-task-bridges/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- detail-view-only `latest_background_task`
- the projected fields:
  - `task_id`
  - `status`
  - `attempt_count`
  - `next_retry_at`
- `null` semantics when no matching task exists

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks raw background-task bridges as landed work and narrows the remaining operator-surface work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed detail-view bridge.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-raw-background-task-bridges/proposal.md`
- `openspec/changes/archive/2026-04-24-raw-background-task-bridges/design.md`
- `openspec/changes/archive/2026-04-24-raw-background-task-bridges/tasks.md`
- `openspec/changes/archive/2026-04-24-raw-background-task-bridges/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-raw-background-task-bridges/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-raw-background-task-bridges
git -c commit.gpgsign=false commit -m "specs: archive raw background task bridges"
```

## Self-Review

- Spec coverage:
  - bridge model: Task 1
  - selection rule: Task 1
  - response shape: Task 1 + Task 2
  - missing/null semantics: Task 1 + Task 3
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Boundary check:
  - detail-view-only bridge is explicit
  - backlog list expansion is explicitly out of scope for this slice
