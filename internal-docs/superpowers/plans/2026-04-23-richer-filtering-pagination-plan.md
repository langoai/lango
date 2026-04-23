# Richer Filtering / Pagination Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the current dead-letter browsing / post-adjudication status read-only surface with practical filtering, offset/limit pagination, total counts, and lightweight navigation hints.

**Architecture:** Extend the existing `internal/postadjudicationstatus` service instead of introducing a new state store. Keep both read-only meta tool names the same while enriching their input and output shapes. The list surface adds outcome filtering, retry-attempt range filtering, text query, and offset/limit pagination. The detail surface adds `is_dead_lettered`, `can_retry`, and `adjudication` hints. Update docs and OpenSpec to describe the richer operator read model.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add query and response metadata types.
- Modify: `internal/postadjudicationstatus/service.go`
  - Add filters, pagination, total count, and detail navigation hints.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover filtering, pagination, totals, and detail hints.
- Modify: `internal/app/tools_meta.go`
  - Upgrade list/detail read-only tool inputs and outputs.
- Modify: `internal/app/tools_parity_test.go`
  - Keep parity expectations aligned if needed.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover new filters/pagination/detail hint behavior through meta tools.
- Create: `docs/architecture/richer-filtering-pagination.md`
  - Public architecture/operator doc for the first usability upgrade slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Truth-align the base read-model page with the richer operator surface.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark richer filtering/pagination as landed and push raw background bridges and higher-level UI work down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/archive/2026-04-23-richer-filtering-pagination/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync enriched read-only tool contracts.

### Task 1: Extend the Post-Adjudication Status Service

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- list filter by `adjudication outcome`
- list filter by retry attempt min/max
- list filter by text query against transaction and submission receipt IDs
- list pagination with `offset` and `limit`
- list response includes `total`, `count`, `offset`, and `limit`
- detail response includes:
  - `is_dead_lettered`
  - `can_retry`
  - `adjudication`

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement filtering, pagination, and hints**

Extend the service so that:

- list query supports:
  - `adjudication`
  - `retry_attempt_min`
  - `retry_attempt_max`
  - `query`
  - `offset`
  - `limit`
- list response includes:
  - `entries`
  - `count`
  - `total`
  - `offset`
  - `limit`
- detail adds:
  - `is_dead_lettered`
  - `can_retry`
  - `adjudication`

- [ ] **Step 4: Re-run the status service tests and verify they pass**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the service slice**

Run:

```bash
git add internal/postadjudicationstatus/types.go internal/postadjudicationstatus/service.go internal/postadjudicationstatus/service_test.go
git -c commit.gpgsign=false commit -m "feat: add post adjudication status filtering"
```

### Task 2: Upgrade the Read-Only Meta Tools

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
- Modify: `internal/app/tools_parity_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- list tool accepts the new filter/pagination parameters
- list tool returns response metadata (`total`, `count`, `offset`, `limit`)
- detail tool returns the new navigation hints

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesPostAdjudicationStatus|ListDeadLetteredPostAdjudicationExecutions_|GetPostAdjudicationExecutionStatus_)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the tool upgrades**

Update the existing read-only tool handlers so that:

- `list_dead_lettered_post_adjudication_executions` accepts the new query params
- the list output includes metadata
- `get_post_adjudication_execution_status` returns the new hint fields

Keep the tools read-only and reuse the extended status service.

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesPostAdjudicationStatus|ListDeadLetteredPostAdjudicationExecutions_|GetPostAdjudicationExecutionStatus_)' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_meta_postadjudicationstatus_test.go internal/app/tools_parity_test.go
git -c commit.gpgsign=false commit -m "app: add status filtering and pagination"
```

### Task 3: Publish Docs and Sync OpenSpec

**Files:**
- Create: `docs/architecture/richer-filtering-pagination.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Create: `openspec/changes/archive/2026-04-23-richer-filtering-pagination/**`

- [ ] **Step 1: Add the public architecture page**

Write `docs/architecture/richer-filtering-pagination.md` describing:

- purpose and scope
- list filter model
- pagination model
- detail navigation hints
- current limits

- [ ] **Step 2: Wire architecture landing, track, and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/dead-letter-browsing-status-observation.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed usability-upgrade slice truthfully.

- [ ] **Step 3: Sync OpenSpec main specs**

Update:

- `openspec/specs/project-docs/spec.md`
- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`

to reflect:

- the new public page
- the landed track status
- the enriched read-only tool contracts

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-richer-filtering-pagination/proposal.md`
- `openspec/changes/archive/2026-04-23-richer-filtering-pagination/design.md`
- `openspec/changes/archive/2026-04-23-richer-filtering-pagination/tasks.md`
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
git add docs/architecture/richer-filtering-pagination.md docs/architecture/index.md docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-richer-filtering-pagination
git -c commit.gpgsign=false commit -m "specs: archive richer filtering pagination"
```

---

## Sequencing Notes

- Task 1 first:
  - status service extension
- Task 2 second:
  - read-only meta tool upgrade
- Task 3 third:
  - public docs
  - spec sync
  - archive closeout

## Verification Checklist

- [ ] `go test ./internal/postadjudicationstatus/... -count=1`
- [ ] `go test ./internal/app -run 'Test(BuildMetaTools_IncludesPostAdjudicationStatus|ListDeadLetteredPostAdjudicationExecutions_|GetPostAdjudicationExecutionStatus_)' -count=1`
- [ ] `.venv/bin/zensical build`
- [ ] `go build ./...`
- [ ] `go test ./...`

## Definition of Done

- dead-letter backlog supports practical filtering
- dead-letter backlog supports offset/limit pagination and total count
- detail view exposes lightweight navigation hints
- public docs and track page reflect the landed slice
- OpenSpec main specs are synced
- archived change exists under `openspec/changes/archive/2026-04-23-richer-filtering-pagination`
- verification passes cleanly
