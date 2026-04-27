# Escrow Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first `escrow release` slice for `knowledge exchange v1`, connecting a funded escrow and an `approved-for-settlement` transaction to real release execution without yet implementing refund or dispute-linked escrow branching.

**Architecture:** Add a small `internal/escrowrelease` service that gates execution on canonical transaction state, resolves the current submission and amount context, then reuses the existing escrow runtime. On success it closes settlement progression to `settled`; on failure it leaves progression unchanged and records release-failure evidence in audit and submission receipt trail.

**Tech Stack:** Go, `internal/receipts`, `internal/escrowexecution`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/escrowrelease/types.go`
  - Request/result types and deny reason model.
- Create: `internal/escrowrelease/service.go`
  - Canonical release gate and funded-escrow release orchestration.
- Create: `internal/escrowrelease/service_test.go`
  - Focused tests for gating, success, failure, and escrow state checks.
- Modify: `internal/receipts/store.go`
  - Add helpers to mark escrow release success and append escrow release failure evidence.
- Modify: `internal/receipts/store_test.go`
  - Cover settled transition and escrow release trail evidence.
- Modify: `internal/app/tools_meta.go`
  - Add `release_escrow_settlement` meta tool and response payload.
- Modify: `internal/app/tools_parity_test.go`
  - Extend runtime-aware parity expectations.
- Create: `internal/app/tools_meta_escrowrelease_test.go`
  - Meta-tool coverage for `release_escrow_settlement`.
- Create: `docs/architecture/escrow-release.md`
  - Public architecture/operator doc for the first escrow release slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark escrow release as landed and move remaining escrow gaps down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/escrow-release/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `release_escrow_settlement` tool contract.

### Task 1: Add Escrow Release Service

**Files:**
- Create: `internal/escrowrelease/types.go`
- Create: `internal/escrowrelease/service.go`
- Create: `internal/escrowrelease/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/escrowrelease/service_test.go` with tests covering:

- execution denied when transaction receipt is missing
- execution denied when there is no current submission
- execution denied when escrow status is not `funded`
- execution denied when settlement progression is not `approved-for-settlement`
- execution denied when amount cannot be resolved
- execution success returns `settled` target shape
- runtime failure keeps `approved-for-settlement` shape

- [ ] **Step 2: Run the escrow release tests and verify they fail**

Run:

```bash
go test ./internal/escrowrelease/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/escrowrelease: no Go files
```

- [ ] **Step 3: Implement the service**

Create:

- request type using `transaction_receipt_id`
- deny reasons:
  - `missing_receipt`
  - `no_current_submission`
  - `escrow_not_funded`
  - `not_approved_for_settlement`
  - `amount_unresolved`

The service must:

- load transaction and current submission
- require `escrow_execution_status = funded`
- require `settlement_progression_status = approved-for-settlement`
- resolve amount from canonical transaction context
- reuse an escrow runtime abstraction for release
- return success shape with:
  - transaction receipt ID
  - submission receipt ID
  - `settled`
  - resolved amount
  - runtime reference
- return failure shape while keeping `approved-for-settlement`

- [ ] **Step 4: Re-run the escrow release tests and verify they pass**

Run:

```bash
go test ./internal/escrowrelease/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the service slice**

Run:

```bash
git add internal/escrowrelease/types.go internal/escrowrelease/service.go internal/escrowrelease/service_test.go
git -c commit.gpgsign=false commit -m "feat: add escrow release service"
```

### Task 2: Extend Receipts for Escrow Release Closeout

**Files:**
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add tests covering:

- `approved-for-settlement -> settled` transition on successful escrow release
- failure does not move progression away from `approved-for-settlement`
- escrow release success appends settlement success evidence
- escrow release failure appends release failure evidence

- [ ] **Step 2: Run the receipt tests and verify they fail**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the receipt helpers**

In `internal/receipts/store.go`, add helpers so escrow release execution can:

- mark settlement progression as `settled`
- append submission-trail evidence for escrow release success/failure

Canonical ownership must remain:

- transaction receipt owns settlement progression state
- submission receipt trail owns append-only execution evidence

- [ ] **Step 4: Re-run the receipt tests and verify they pass**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the receipt slice**

Run:

```bash
git add internal/receipts/store.go internal/receipts/store_test.go
git -c commit.gpgsign=false commit -m "feat: add escrow release receipt updates"
```

### Task 3: Add `release_escrow_settlement` Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_escrowrelease_test.go`
- Modify: `internal/app/modules.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_escrowrelease_test.go` with tests for:

- runtime-aware meta tool registration includes `release_escrow_settlement`
- the tool is absent when the escrow release runtime is unavailable
- funded + approved path executes and returns canonical result
- wrong escrow or settlement state returns an error

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesReleaseEscrowSettlement|TestBuildMetaTools_OmitsReleaseEscrowSettlementWithoutRuntime|TestReleaseEscrowSettlement_' -count=1
```

Expected:

```text
FAIL
tool not found
```

- [ ] **Step 3: Implement the meta tool**

In `internal/app/tools_meta.go`, add:

- `release_escrow_settlement`
- required input: `transaction_receipt_id`
- handler using the escrow release service
- thin receipt payload including:
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `settlement_progression_status`
  - `resolved_amount`
  - `runtime_reference`

Update runtime-aware meta tool assembly and parity expectations.

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesReleaseEscrowSettlement|TestBuildMetaTools_OmitsReleaseEscrowSettlementWithoutRuntime|TestReleaseEscrowSettlement_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_escrowrelease_test.go internal/app/modules.go
git -c commit.gpgsign=false commit -m "app: add release escrow settlement meta tool"
```

### Task 4: Document and Close Out the Slice

**Files:**
- Create: `docs/architecture/escrow-release.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Create: `openspec/changes/escrow-release/**`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`

- [ ] **Step 1: Write the public architecture page**

Create `docs/architecture/escrow-release.md` describing:

- what ships in the first escrow release slice
- gate conditions
- canonical source resolution
- success/failure semantics
- current limits

- [ ] **Step 2: Wire the page into architecture docs and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed escrow release slice truthfully.

- [ ] **Step 3: Write and archive the OpenSpec change**

Create:

- `openspec/changes/escrow-release/proposal.md`
- `openspec/changes/escrow-release/design.md`
- `openspec/changes/escrow-release/tasks.md`
- delta specs for:
  - `project-docs`
  - `docs-only`
  - `meta-tools`

Then sync main specs and archive under:

```text
openspec/changes/archive/2026-04-22-escrow-release
```

- [ ] **Step 4: Run final verification**

Run:

```bash
.venv/bin/zensical build
go build ./...
go test ./...
```

Expected:

```text
All commands exit 0.
```

- [ ] **Step 5: Commit the docs/OpenSpec slice**

Run:

```bash
git add docs/architecture/escrow-release.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-22-escrow-release
git -c commit.gpgsign=false commit -m "specs: archive escrow release"
```

## Self-Review

- Spec coverage:
  - escrow release service: Task 1
  - receipt closeout and trail evidence: Task 2
  - `release_escrow_settlement` meta tool: Task 3
  - public docs and OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or refund/dispute overclaims remain
- Type/path consistency:
  - escrow release stays transaction-owned
  - current submission remains the evidence anchor
  - amount resolution continues to come from canonical transaction context
