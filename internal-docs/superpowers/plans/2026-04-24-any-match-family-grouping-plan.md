# Any-Match Family Grouping Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the dead-letter backlog list with `any_match_families` and `any_match_family` while keeping the surface read-only and current-submission centered.

**Architecture:** Extend the existing `internal/postadjudicationstatus` read model rather than introducing a new store. Derive a deduplicated family set from relevant retry lifecycle events on the current submission trail, expose it on each backlog row, and add a single-family membership filter. Update the existing list meta tool to pass the new filter through, then truth-align docs and OpenSpec.

**Tech Stack:** Go, `internal/postadjudicationstatus`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationstatus/types.go`
  - Add `AnyMatchFamilies` to the row/summary shape and `AnyMatchFamily` to list options.
- Modify: `internal/postadjudicationstatus/service.go`
  - Derive the deduplicated family set from relevant current-submission retry events.
  - Add membership filtering for `AnyMatchFamily`.
- Modify: `internal/postadjudicationstatus/service_test.go`
  - Cover family derivation and single-family membership filtering.
- Modify: `internal/app/tools_meta.go`
  - Add the new list-tool parameter and pass it through.
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`
  - Cover `any_match_family` and `any_match_families` through the meta-tool surface.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe `any_match_families` and `any_match_family`.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark any-match family grouping as landed and narrow the remaining grouping work.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync the list-tool contract.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the dead-letter browsing page and track requirements.
- Create: `openspec/changes/archive/2026-04-24-any-match-family-grouping/**`
  - Proposal, design, tasks, and delta specs.

### Task 1: Extend the Dead-Letter Read Model

**Files:**
- Modify: `internal/postadjudicationstatus/types.go`
- Modify: `internal/postadjudicationstatus/service.go`
- Modify: `internal/postadjudicationstatus/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Add tests covering:

- derivation of `any_match_families`
- `any_match_family` membership filtering
- deduplication of repeated family hits

- [ ] **Step 2: Run the status service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement any-match family derivation and filtering**

Extend the read model so that:

- `RetryDeadLetterSummary` includes `AnyMatchFamilies`
- `DeadLetterBacklogEntry` includes `AnyMatchFamilies`
- `DeadLetterListOptions` accepts `AnyMatchFamily`
- relevant events:
  - `retry-scheduled`
  - `manual-retry-requested`
  - `dead-lettered`
- subtype-to-family mapping:
  - `retry`
  - `manual-retry`
  - `dead-letter`
- the derived family set is deduplicated
- the list matcher applies membership matching for `AnyMatchFamily`

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
git -c commit.gpgsign=false commit -m "feat: add any match family grouping"
```

### Task 2: Upgrade the Read-Only Meta Tool Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_postadjudicationstatus_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Add tests covering:

- `any_match_family`
- list entries carrying `any_match_families`

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

- it accepts `any_match_family`
- it passes the value into `DeadLetterListOptions`
- it returns the new `any_match_families` row field while keeping the page shape unchanged

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
git -c commit.gpgsign=false commit -m "app: add any match family filter"
```

### Task 3: Truth-Align Docs and OpenSpec

**Files:**
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-24-any-match-family-grouping/**`

- [ ] **Step 1: Update the public architecture page**

Update `docs/architecture/dead-letter-browsing-status-observation.md` to describe:

- `any_match_family`
- `any_match_families`

- [ ] **Step 2: Update the track doc**

Update `docs/architecture/p2p-knowledge-exchange-track.md` so it marks any-match family grouping as landed work and narrows the remaining grouping work.

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/meta-tools/spec.md`
- `openspec/specs/docs-only/spec.md`

to reflect the landed any-match family grouping slice.

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-24-any-match-family-grouping/proposal.md`
- `openspec/changes/archive/2026-04-24-any-match-family-grouping/design.md`
- `openspec/changes/archive/2026-04-24-any-match-family-grouping/tasks.md`
- `openspec/changes/archive/2026-04-24-any-match-family-grouping/specs/meta-tools/spec.md`
- `openspec/changes/archive/2026-04-24-any-match-family-grouping/specs/docs-only/spec.md`

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
git add docs/architecture/dead-letter-browsing-status-observation.md docs/architecture/p2p-knowledge-exchange-track.md openspec/specs/meta-tools/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-24-any-match-family-grouping
git -c commit.gpgsign=false commit -m "specs: archive any match family grouping"
```

## Self-Review

- Spec coverage:
  - grouping model: Task 1
  - filter model: Task 1 + Task 2
  - response shape: Task 1 + Task 2
  - evidence source: Task 1
  - docs/OpenSpec truth alignment: Task 3
- Placeholder scan:
  - no placeholders or deferred implementation notes remain in task steps
- Type consistency:
  - `AnyMatchFamilies` and `AnyMatchFamily` are used consistently across the plan
