# Reason / Dispatch Reference Filters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the dead-letter backlog list with dead-letter reason substring filtering and dispatch-reference exact-match filtering without changing the existing response shape.

**Architecture:** Keep the current transaction-centered read model and extend only the existing backlog list path. Add `dead_letter_reason_query` and `latest_dispatch_reference` to `internal/postadjudicationstatus.DeadLetterListOptions`, implement the new matching rules in the existing filter matcher, and pass the new query params through the existing read-only meta tool. Update docs and OpenSpec to describe the richer backlog query surface.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add the new reason/dispatch list options.
- Modify: `internal/postadjudicationstatus/service.go`
  - Extend backlog filter matching with reason substring and dispatch exact matching.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover reason and dispatch filtering in the read model.
- Modify: `internal/app/tools_meta.go`
  - Add the new read-only query parameters to the backlog list tool.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover reason/dispatch filtering through the meta-tool surface.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the reason and dispatch filters.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark reason/dispatch filtering as landed and narrow the remaining list work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the enriched list-tool contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Create: `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters/**`
  - Proposal, design, tasks, and delta specs for the completed slice.

### Task 1: Extend the Dead-Letter Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- `dead_letter_reason_query` substring matching
- `latest_dispatch_reference` exact matching
- both new filters working together with existing filters using `AND`

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the new list filters**

Extend the read model so that:

- `DeadLetterListOptions` accepts:
  - `DeadLetterReasonQuery`
  - `LatestDispatchReference`
- the list filter matcher adds:
  - case-insensitive substring matching on `latest_dead_letter_reason`
  - exact matching on `latest_dispatch_reference`
- the list response shape remains unchanged

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
git -c commit.gpgsign=false commit -m "feat: add reason and dispatch status filters"
```

### Task 2: Upgrade the Read-Only Meta Tool Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `dead_letter_reason_query`
- `latest_dispatch_reference`
- both parameters flowing through the list tool correctly

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

- it accepts `dead_letter_reason_query`
- it accepts `latest_dispatch_reference`
- it passes both into `DeadLetterListOptions`

Keep the tool read-only and keep the response shape unchanged.

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
git -c commit.gpgsign=false commit -m "app: add reason and dispatch backlog filters"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `dead_letter_reason_query`
- `latest_dispatch_reference`

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks reason/dispatch filters as landed work and narrows the remaining backlog-list work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed reason/dispatch list filters.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters/proposal.md`
- `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters/design.md`
- `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters/tasks.md`
- `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-reason-dispatch-reference-filters
git -c commit.gpgsign=false commit -m "specs: archive reason dispatch filters"
```

## Self-Review

- Spec coverage:
  - filter model: Task 1 + Task 2
  - matching semantics: Task 1
  - unchanged response shape: Task 1 + Task 2
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders, `TBD`, or implicit follow-up work inside the tasks
- Type consistency:
  - `dead_letter_reason_query` and `latest_dispatch_reference` are used consistently across the plan
