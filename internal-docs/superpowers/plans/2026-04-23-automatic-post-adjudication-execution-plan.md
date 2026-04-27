# Automatic Post-Adjudication Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first inline convenience slice that allows `adjudicate_escrow_dispute` to optionally execute the adjudicated release or refund branch in the same request without rolling back adjudication on nested execution failure.

**Architecture:** Keep `internal/escrowadjudication` responsible only for canonical adjudication writes. Extend the `adjudicate_escrow_dispute` meta tool handler to accept `auto_execute`, route to existing `escrowrelease` / `escrowrefund` services after successful adjudication, and return both adjudication and nested execution results. Update public docs and OpenSpec to describe this as an inline orchestration convenience layer rather than a new lifecycle state.

**Tech Stack:** Go, `internal/app`, `internal/escrowadjudication`, `internal/escrowrelease`, `internal/escrowrefund`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/app/tools_meta.go`
  - Extend `adjudicate_escrow_dispute` with `auto_execute`.
  - Add nested execution return payload shape.
- Modify: `internal/app/tools_meta_escrowadjudication_test.go`
  - Add coverage for inline release/refund execution and nested failure semantics.
- Create: `docs/architecture/automatic-post-adjudication-execution.md`
  - Public architecture/operator doc for the first auto-execution slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark automatic post-adjudication execution as landed and push async/retry work down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/automatic-post-adjudication-execution/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `adjudicate_escrow_dispute(auto_execute=true)` contract.

### Task 1: Add Inline Auto-Execution to the Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_meta_escrowadjudication_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Extend `internal/app/tools_meta_escrowadjudication_test.go` with tests covering:

- tool schema includes `auto_execute`
- `auto_execute=true` with `release` returns:
  - adjudication result
  - nested release execution result
- `auto_execute=true` with `refund` returns:
  - adjudication result
  - nested refund execution result
- nested execution failure still returns the adjudication receipt shape while surfacing an error

- [ ] **Step 2: Run the adjudication meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesAdjudicateEscrowDispute|AdjudicateEscrowDispute_)' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement inline auto execution**

Update `internal/app/tools_meta.go` so that:

- `adjudicate_escrow_dispute` accepts optional `auto_execute`
- `auto_execute=true` is mutually exclusive with any future `background_execute`
- adjudication still happens first
- after successful adjudication:
  - `release` routes to existing `escrowrelease.NewService(...).Execute(...)`
  - `refund` routes to existing `escrowrefund.NewService(...).Execute(...)`
- the tool returns:
  - adjudication result
  - nested execution result when requested
- nested execution failure does **not** roll back adjudication

- [ ] **Step 4: Re-run the adjudication meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'Test(BuildMetaTools_IncludesAdjudicateEscrowDispute|AdjudicateEscrowDispute_)' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the inline auto-execution slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_meta_escrowadjudication_test.go
git -c commit.gpgsign=false commit -m "feat: add automatic post adjudication execution"
```

### Task 2: Publish Docs and Sync OpenSpec

**Files:**
- Create: `docs/architecture/automatic-post-adjudication-execution.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Create: `openspec/changes/archive/2026-04-23-automatic-post-adjudication-execution/**`

- [ ] **Step 1: Add the public architecture page**

Write `docs/architecture/automatic-post-adjudication-execution.md` describing:

- purpose and scope
- `auto_execute=true` trigger
- inline orchestration semantics
- adjudication vs nested execution failure separation
- current limits

- [ ] **Step 2: Wire architecture landing, track, and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed auto-execution slice truthfully.

- [ ] **Step 3: Sync OpenSpec main specs**

Update:

- `openspec/specs/project-docs/spec.md`
- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`

to reflect:

- the new public page
- the landed track status
- the `adjudicate_escrow_dispute(auto_execute=true)` contract

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-automatic-post-adjudication-execution/proposal.md`
- `openspec/changes/archive/2026-04-23-automatic-post-adjudication-execution/design.md`
- `openspec/changes/archive/2026-04-23-automatic-post-adjudication-execution/tasks.md`
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
git add docs/architecture/automatic-post-adjudication-execution.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-automatic-post-adjudication-execution
git -c commit.gpgsign=false commit -m "specs: archive automatic post adjudication execution"
```

---

## Sequencing Notes

- Task 1 first:
  - inline auto execution behavior and tests
- Task 2 second:
  - public docs
  - spec sync
  - archive closeout

## Verification Checklist

- [ ] `go test ./internal/app -run 'Test(BuildMetaTools_IncludesAdjudicateEscrowDispute|AdjudicateEscrowDispute_)' -count=1`
- [ ] `.venv/bin/zensical build`
- [ ] `go build ./...`
- [ ] `go test ./...`

## Definition of Done

- `adjudicate_escrow_dispute` supports optional `auto_execute`
- inline execution reuses existing release/refund services
- adjudication survives nested execution failure
- public docs and track page reflect the landed slice
- OpenSpec main specs are synced
- archived change exists under `openspec/changes/archive/2026-04-23-automatic-post-adjudication-execution`
- verification passes cleanly
