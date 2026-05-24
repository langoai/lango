# Dead-Letter Browsing / Status Observation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first read-only operator surface for post-adjudication failures, including a dead-letter backlog view and a per-transaction status view grounded in transaction receipts and submission receipt trails.

**Architecture:** Add a small read-only status service that reads `transaction receipt`, `current submission receipt`, and `submission receipt trail` to produce a dead-letter backlog list and a detail view. Expose these through two read-only meta tools: `list_dead_lettered_post_adjudication_executions` and `get_post_adjudication_execution_status`. Update public docs and OpenSpec to describe the first operator visibility layer for dead-lettered post-adjudication execution.

**Tech Stack:** Go, `internal/receipts`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/postadjudicationstatus/types.go`
  - Read-model types for list and detail responses.
- Create: `internal/postadjudicationstatus/service.go`
  - Read-only status service for dead-letter backlog and per-transaction detail.
- Create: `internal/postadjudicationstatus/service_test.go`
  - Focused tests for list and detail extraction.
- Modify: `internal/app/tools_meta.go`
  - Add `list_dead_lettered_post_adjudication_executions`.
  - Add `get_post_adjudication_execution_status`.
- Modify: `internal/app/tools_parity_test.go`
  - Extend parity expectations.
- Create: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Meta-tool coverage for list/detail read-only surfaces.
- Create: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Public architecture/operator doc for the first read-only visibility slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark dead-letter browsing/status observation as landed and push richer observability work down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/archive/2026-04-23-dead-letter-browsing-status-observation/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the two new read-only tool contracts.

### Task 1: Add Post-Adjudication Status Service

**Files:**
- Create: `internal/postadjudicationstatus/types.go`
- Create: `internal/postadjudicationstatus/service.go`
- Create: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- list returns only transactions that are currently dead-lettered
- list includes:
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `adjudication`
  - `latest_dead_letter_reason`
  - `latest_retry_attempt`
  - `latest_dispatch_reference`
- detail returns:
  - current canonical snapshot
  - latest retry/dead-letter summary
- detail fails when transaction receipt is missing

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the read-only status service**

The service must:

- inspect transaction receipts
- follow the current submission pointer
- summarize latest dead-letter and retry evidence from the submission trail
- provide:
  - current dead-letter backlog list
  - per-transaction detail

- [ ] **Step 4: Re-run the status service tests and verify they pass**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the status service slice**

Run:

```bash
git add internal/postadjudicationstatus/types.go internal/postadjudicationstatus/service.go internal/postadjudicationstatus/service_test.go
git -c commit.gpgsign=false commit -m "feat: add post adjudication status service"
```

### Task 2: Add Read-Only Meta Tools

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `list_dead_lettered_post_adjudication_executions` is registered
- `get_post_adjudication_execution_status` is registered
- list tool returns only current dead-lettered backlog
- detail tool returns current canonical snapshot plus latest retry/dead-letter summary

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesPostAdjudicationStatus|ListDeadLetteredPostAdjudicationExecutions_|GetPostAdjudicationExecutionStatus_)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the read-only meta tools**

Add:

- `list_dead_lettered_post_adjudication_executions`
- `get_post_adjudication_execution_status`

Both tools must:

- be read-only
- use the new status service
- not mutate any canonical state

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
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_postadjudicationstatus_test.go
git -c commit.gpgsign=false commit -m "app: add post adjudication status tools"
```

### Task 3: Publish Docs and Sync OpenSpec

**Files:**
- Create: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Create: `openspec/changes/archive/2026-04-23-dead-letter-browsing-status-observation/**`

- [ ] **Step 1: Add the public architecture page**

Write `docs/architecture/dead-letter-browsing-status-observation.md` describing:

- purpose and scope
- list surface
- detail surface
- current limits

- [ ] **Step 2: Wire architecture landing, track, and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed read-only visibility slice truthfully.

- [ ] **Step 3: Sync OpenSpec main specs**

Update:

- `openspec/specs/project-docs/spec.md`
- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`

to reflect:

- the new public page
- the landed track status
- the `list_dead_lettered_post_adjudication_executions` and `get_post_adjudication_execution_status` contracts

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-dead-letter-browsing-status-observation/proposal.md`
- `openspec/changes/archive/2026-04-23-dead-letter-browsing-status-observation/design.md`
- `openspec/changes/archive/2026-04-23-dead-letter-browsing-status-observation/tasks.md`
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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-dead-letter-browsing-status-observation
git -c commit.gpgsign=false commit -m "specs: archive dead letter browsing status observation"
```

---

## Sequencing Notes

- Task 1 first:
  - read-only status service
- Task 2 second:
  - list/detail meta tools
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

- current dead-letter backlog can be listed
- per-transaction post-adjudication status can be inspected
- both tools are read-only
- public docs and track page reflect the landed slice
- OpenSpec main specs are synced
- archived change exists under `openspec/changes/archive/2026-04-23-dead-letter-browsing-status-observation`
- verification passes cleanly
