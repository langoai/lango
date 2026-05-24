# Dispute Hold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first `dispute hold` slice for `knowledge exchange v1`, recording that a funded escrow has been held after canonical dispute handoff without yet implementing release vs refund adjudication.

**Architecture:** Add a small `internal/disputehold` service that gates execution on canonical transaction state (`funded` escrow + `dispute-ready` settlement progression), resolves the current submission and escrow reference, and records hold success/failure evidence in audit and submission receipt trail. The first slice does not add a new escrow lifecycle state.

**Tech Stack:** Go, `internal/receipts`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/disputehold/types.go`
  - Request/result types and deny reason model.
- Create: `internal/disputehold/service.go`
  - Canonical hold gate and evidence orchestration.
- Create: `internal/disputehold/service_test.go`
  - Focused tests for gating, success, and failure.
- Modify: `internal/receipts/types.go`
  - Add hold-evidence request type if needed.
- Modify: `internal/receipts/store.go`
  - Add helpers to append dispute hold success/failure evidence.
- Modify: `internal/receipts/store_test.go`
  - Cover hold evidence and state non-mutation.
- Modify: `internal/app/tools_meta.go`
  - Add `hold_escrow_for_dispute` meta tool and response payload.
- Modify: `internal/app/tools_parity_test.go`
  - Extend runtime-aware parity expectations.
- Create: `internal/app/tools_meta_disputehold_test.go`
  - Meta-tool coverage for `hold_escrow_for_dispute`.
- Create: `docs/architecture/dispute-hold.md`
  - Public architecture/operator doc for the first dispute hold slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark dispute hold as landed and move remaining adjudication gaps down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/dispute-hold/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `hold_escrow_for_dispute` tool contract.

### Task 1: Add Dispute Hold Service

**Files:**
- Create: `internal/disputehold/types.go`
- Create: `internal/disputehold/service.go`
- Create: `internal/disputehold/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/disputehold/service_test.go` with tests covering:

- execution denied when transaction receipt is missing
- execution denied when there is no current submission
- execution denied when escrow status is not `funded`
- execution denied when settlement progression is not `dispute-ready`
- execution denied when escrow reference is missing
- execution success returns a hold-applied shape while keeping state unchanged
- hold failure returns a failure shape while keeping state unchanged

- [ ] **Step 2: Run the dispute hold tests and verify they fail**

Run:

```bash
go test ./internal/disputehold/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/disputehold: no Go files
```

- [ ] **Step 3: Implement the service**

Create:

- request type using `transaction_receipt_id`
- deny reasons:
  - `missing_receipt`
  - `no_current_submission`
  - `escrow_not_funded`
  - `not_dispute_ready`
  - `escrow_reference_missing`

The service must:

- load transaction and current submission
- require `escrow_execution_status = funded`
- require `settlement_progression_status = dispute-ready`
- require `escrow_reference`
- record success shape with:
  - transaction receipt ID
  - submission receipt ID
  - unchanged settlement progression status
  - escrow reference
- record failure shape while keeping state unchanged

- [ ] **Step 4: Re-run the dispute hold tests and verify they pass**

Run:

```bash
go test ./internal/disputehold/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the service slice**

Run:

```bash
git add internal/disputehold/types.go internal/disputehold/service.go internal/disputehold/service_test.go
git -c commit.gpgsign=false commit -m "feat: add dispute hold service"
```

### Task 2: Extend Receipts for Dispute Hold Evidence

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add tests covering:

- hold success does not mutate escrow or settlement progression state
- hold success appends hold evidence
- hold failure does not mutate state
- hold failure appends hold-failure evidence

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

In `internal/receipts/store.go`, add helpers so dispute hold can:

- append submission-trail evidence for hold success/failure
- leave escrow execution and settlement progression unchanged

Canonical ownership must remain:

- transaction receipt owns canonical state
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
git add internal/receipts/types.go internal/receipts/store.go internal/receipts/store_test.go
git -c commit.gpgsign=false commit -m "feat: add dispute hold receipt evidence"
```

### Task 3: Add `hold_escrow_for_dispute` Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_disputehold_test.go`
- Modify: `internal/app/modules.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_disputehold_test.go` with tests for:

- runtime-aware meta tool registration includes `hold_escrow_for_dispute`
- the tool is absent when the hold runtime is unavailable
- dispute-ready funded path returns canonical result
- wrong escrow or settlement state returns an error

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesDisputeHold|TestBuildMetaTools_OmitsDisputeHoldWithoutRuntime|TestHoldEscrowForDispute_' -count=1
```

Expected:

```text
FAIL
tool not found
```

- [ ] **Step 3: Implement the meta tool**

In `internal/app/tools_meta.go`, add:

- `hold_escrow_for_dispute`
- required input: `transaction_receipt_id`
- handler using the dispute hold service
- thin receipt payload including:
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `settlement_progression_status`
  - `escrow_reference`

Update runtime-aware meta tool assembly and parity expectations.

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesDisputeHold|TestBuildMetaTools_OmitsDisputeHoldWithoutRuntime|TestHoldEscrowForDispute_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_disputehold_test.go internal/app/modules.go
git -c commit.gpgsign=false commit -m "app: add dispute hold meta tool"
```

### Task 4: Document and Close Out the Slice

**Files:**
- Create: `docs/architecture/dispute-hold.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Create: `openspec/changes/dispute-hold/**`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`

- [ ] **Step 1: Write the public architecture page**

Create `docs/architecture/dispute-hold.md` describing:

- what ships in the first dispute hold slice
- gate conditions
- canonical source resolution
- success/failure semantics
- current limits

- [ ] **Step 2: Wire the page into architecture docs and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed dispute hold slice truthfully.

- [ ] **Step 3: Write and archive the OpenSpec change**

Create:

- `openspec/changes/dispute-hold/proposal.md`
- `openspec/changes/dispute-hold/design.md`
- `openspec/changes/dispute-hold/tasks.md`
- delta specs for:
  - `project-docs`
  - `docs-only`
  - `meta-tools`

Then sync main specs and archive under:

```text
openspec/changes/archive/2026-04-23-dispute-hold
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
git add docs/architecture/dispute-hold.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-dispute-hold
git -c commit.gpgsign=false commit -m "specs: archive dispute hold"
```

## Self-Review

- Spec coverage:
  - dispute hold service: Task 1
  - dispute hold evidence helpers: Task 2
  - `hold_escrow_for_dispute` meta tool: Task 3
  - public docs and OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or adjudication overclaims remain
- Type/path consistency:
  - dispute hold leaves canonical transaction state unchanged
  - current submission remains the evidence anchor
  - escrow reference remains transaction-derived, not tool-supplied
