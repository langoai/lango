# Escrow Refund Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first `escrow refund` slice for `knowledge exchange v1`, refunding a funded but unreleased escrow from the settlement review path without yet defining refund terminal states or dispute-linked branching.

**Architecture:** Add a small `internal/escrowrefund` service that gates execution on canonical transaction state, resolves the current submission and amount context, then reuses the existing escrow runtime. On success it keeps settlement progression at `review-needed` and records refund success evidence. On failure it also keeps progression unchanged and records refund-failure evidence in audit and submission receipt trail.

**Tech Stack:** Go, `internal/receipts`, `internal/escrowrelease`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/escrowrefund/types.go`
  - Request/result types and deny reason model.
- Create: `internal/escrowrefund/service.go`
  - Canonical refund gate and funded-escrow refund orchestration.
- Create: `internal/escrowrefund/service_test.go`
  - Focused tests for gating, success, failure, and review-path checks.
- Modify: `internal/receipts/store.go`
  - Add helpers to append escrow refund success/failure evidence without mutating progression.
- Modify: `internal/receipts/store_test.go`
  - Cover refund evidence and progression non-mutation.
- Modify: `internal/app/tools_meta.go`
  - Add `refund_escrow_settlement` meta tool and response payload.
- Modify: `internal/app/tools_parity_test.go`
  - Extend runtime-aware parity expectations.
- Create: `internal/app/tools_meta_escrowrefund_test.go`
  - Meta-tool coverage for `refund_escrow_settlement`.
- Create: `docs/architecture/escrow-refund.md`
  - Public architecture/operator doc for the first escrow refund slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark escrow refund as landed and move remaining refund/dispute gaps down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/escrow-refund/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `refund_escrow_settlement` tool contract.

### Task 1: Add Escrow Refund Service

**Files:**
- Create: `internal/escrowrefund/types.go`
- Create: `internal/escrowrefund/service.go`
- Create: `internal/escrowrefund/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/escrowrefund/service_test.go` with tests covering:

- execution denied when transaction receipt is missing
- execution denied when there is no current submission
- execution denied when escrow status is not `funded`
- execution denied when settlement progression is not `review-needed`
- execution denied when amount cannot be resolved
- execution success returns a refund-executed shape while keeping review-needed status
- runtime failure keeps `review-needed` status in the result

- [ ] **Step 2: Run the escrow refund tests and verify they fail**

Run:

```bash
go test ./internal/escrowrefund/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/escrowrefund: no Go files
```

- [ ] **Step 3: Implement the service**

Create:

- request type using `transaction_receipt_id`
- deny reasons:
  - `missing_receipt`
  - `no_current_submission`
  - `escrow_not_funded`
  - `not_review_needed`
  - `amount_unresolved`

The service must:

- load transaction and current submission
- require `escrow_execution_status = funded`
- require `settlement_progression_status = review-needed`
- resolve amount from canonical transaction context
- reuse an escrow runtime abstraction for refund
- return success shape with:
  - transaction receipt ID
  - submission receipt ID
  - `review-needed`
  - resolved amount
  - runtime reference
- return failure shape while keeping `review-needed`

- [ ] **Step 4: Re-run the escrow refund tests and verify they pass**

Run:

```bash
go test ./internal/escrowrefund/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the service slice**

Run:

```bash
git add internal/escrowrefund/types.go internal/escrowrefund/service.go internal/escrowrefund/service_test.go
git -c commit.gpgsign=false commit -m "feat: add escrow refund service"
```

### Task 2: Extend Receipts for Escrow Refund Evidence

**Files:**
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add tests covering:

- refund success does not mutate settlement progression
- refund success appends refund success evidence
- refund failure does not mutate settlement progression
- refund failure appends refund failure evidence

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

In `internal/receipts/store.go`, add helpers so escrow refund execution can:

- append submission-trail evidence for escrow refund success/failure
- leave settlement progression unchanged

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
git -c commit.gpgsign=false commit -m "feat: add escrow refund receipt updates"
```

### Task 3: Add `refund_escrow_settlement` Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_escrowrefund_test.go`
- Modify: `internal/app/modules.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_escrowrefund_test.go` with tests for:

- runtime-aware meta tool registration includes `refund_escrow_settlement`
- the tool is absent when the escrow refund runtime is unavailable
- funded + review-needed path executes and returns canonical result
- wrong escrow or settlement state returns an error

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesRefundEscrowSettlement|TestBuildMetaTools_OmitsRefundEscrowSettlementWithoutRuntime|TestRefundEscrowSettlement_' -count=1
```

Expected:

```text
FAIL
tool not found
```

- [ ] **Step 3: Implement the meta tool**

In `internal/app/tools_meta.go`, add:

- `refund_escrow_settlement`
- required input: `transaction_receipt_id`
- handler using the escrow refund service
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
go test ./internal/app -run 'TestBuildMetaTools_IncludesRefundEscrowSettlement|TestBuildMetaTools_OmitsRefundEscrowSettlementWithoutRuntime|TestRefundEscrowSettlement_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_escrowrefund_test.go internal/app/modules.go
git -c commit.gpgsign=false commit -m "app: add refund escrow settlement meta tool"
```

### Task 4: Document and Close Out the Slice

**Files:**
- Create: `docs/architecture/escrow-refund.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Create: `openspec/changes/escrow-refund/**`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`

- [ ] **Step 1: Write the public architecture page**

Create `docs/architecture/escrow-refund.md` describing:

- what ships in the first escrow refund slice
- gate conditions
- canonical source resolution
- success/failure semantics
- current limits

- [ ] **Step 2: Wire the page into architecture docs and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed escrow refund slice truthfully.

- [ ] **Step 3: Write and archive the OpenSpec change**

Create:

- `openspec/changes/escrow-refund/proposal.md`
- `openspec/changes/escrow-refund/design.md`
- `openspec/changes/escrow-refund/tasks.md`
- delta specs for:
  - `project-docs`
  - `docs-only`
  - `meta-tools`

Then sync main specs and archive under:

```text
openspec/changes/archive/2026-04-22-escrow-refund
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
git add docs/architecture/escrow-refund.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-22-escrow-refund
git -c commit.gpgsign=false commit -m "specs: archive escrow refund"
```

## Self-Review

- Spec coverage:
  - escrow refund service: Task 1
  - refund evidence helpers: Task 2
  - `refund_escrow_settlement` meta tool: Task 3
  - public docs and OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or dispute-branch overclaims remain
- Type/path consistency:
  - refund keeps settlement progression unchanged in this first slice
  - current submission remains the evidence anchor
  - amount resolution continues to come from canonical transaction context
