# Actor / Time-Based Filters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the dead-letter backlog list with manual replay actor filtering and dead-letter time-window filtering while keeping the surface read-only and transaction-centered.

**Architecture:** Upgrade the existing `internal/postadjudicationstatus` read model instead of introducing a new store. Extract `latest_manual_replay_actor` and `latest_dead_lettered_at` from the submission receipt trail, add `manual_replay_actor`, `dead_lettered_after`, and `dead_lettered_before` filters to the list path, and expose the new fields in backlog entries. Keep `get_post_adjudication_execution_status` unchanged for this slice. Update docs and OpenSpec to reflect the richer operator backlog surface.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add actor/time filters and response fields for backlog entries.
- Modify: `internal/postadjudicationstatus/service.go`
  - Extract latest manual replay actor and latest dead-letter timestamp from submission trail.
  - Extend list filter matching with actor and time-window checks.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover actor filter, time filters, combined filtering, and response field extraction.
- Modify: `internal/app/tools_meta.go`
  - Add the new list-tool query parameters and return the richer entry fields.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover actor/time filters through the read-only meta tool surface.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe actor/time filters and the new entry fields.
- Modify: `docs/architecture/index.md`
  - Keep the architecture landing summary aligned with the richer backlog surface.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark actor/time filters as landed and push remaining list work down one level.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the list tool contract with actor/time filters and response fields.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync the architecture landing summary requirement.
- Create: `openspec/changes/archive/2026-04-23-actor-time-based-filters/**`
  - Proposal, design, tasks, and delta specs for the completed slice.

### Task 1: Extend the Post-Adjudication Status Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- `manual_replay_actor` filter
- `dead_lettered_after`
- `dead_lettered_before`
- actor/time filters combined with existing filters using `AND`
- backlog entry fields:
  - `latest_dead_lettered_at`
  - `latest_manual_replay_actor`

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement actor/time extraction and filtering**

Extend the read model so that:

- list options accept:
  - `manual_replay_actor`
  - `dead_lettered_after`
  - `dead_lettered_before`
- backlog entries expose:
  - `latest_dead_lettered_at`
  - `latest_manual_replay_actor`
- trail summarization extracts:
  - latest dead-letter timestamp
  - latest manual replay actor
- filter matching applies all filters with `AND`

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
git -c commit.gpgsign=false commit -m "feat: add actor and time dead letter filters"
```

### Task 2: Upgrade the Read-Only Meta Tool Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `list_dead_lettered_post_adjudication_executions` accepts:
  - `manual_replay_actor`
  - `dead_lettered_after`
  - `dead_lettered_before`
- the list output includes:
  - `latest_dead_lettered_at`
  - `latest_manual_replay_actor`

- [ ] **Step 2: Run the focused meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(ListDeadLetteredPostAdjudicationExecutions_|BuildMetaTools_IncludesPostAdjudicationStatus)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the list-tool upgrade**

Update `list_dead_lettered_post_adjudication_executions` so that:

- it accepts the new actor/time filters
- it passes them into the status service
- it returns the richer entry fields without changing the overall page shape

Keep the tool read-only and leave `get_post_adjudication_execution_status` unchanged.

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
git -c commit.gpgsign=false commit -m "app: add actor and time status filters"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/project-docs/spec.md`
- Create: `openspec/changes/archive/2026-04-23-actor-time-based-filters/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `manual_replay_actor`
- `dead_lettered_after`
- `dead_lettered_before`
- `latest_dead_lettered_at`
- `latest_manual_replay_actor`

- [ ] **Step 2: Update landing and track docs**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`

so they describe actor/time filtering as landed work and narrow the remaining backlog work accordingly.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/project-docs/spec.md`

to reflect the landed actor/time filter surface.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-actor-time-based-filters/proposal.md`
- `openspec/changes/archive/2026-04-23-actor-time-based-filters/design.md`
- `openspec/changes/archive/2026-04-23-actor-time-based-filters/tasks.md`
- `openspec/changes/archive/2026-04-23-actor-time-based-filters/specs/docs-only/spec.md`
- `openspec/changes/archive/2026-04-23-actor-time-based-filters/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-23-actor-time-based-filters/specs/project-docs/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/specs/project-docs/spec.md openspec/changes/archive/2026-04-23-actor-time-based-filters
git -c commit.gpgsign=false commit -m "specs: archive actor time filters"
```

## Self-Review

- Spec coverage:
  - filter model: Task 1 + Task 2
  - evidence sources: Task 1
  - response shape: Task 1 + Task 2
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no `TODO`, `TBD`, or deferred implementation placeholders in task steps
- Type consistency:
  - uses `manual_replay_actor`, `dead_lettered_after`, `dead_lettered_before`, `latest_dead_lettered_at`, and `latest_manual_replay_actor` consistently across tasks
