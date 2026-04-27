# Partial Settlement Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first direct `partial settlement execution` slice for `knowledge exchange v1`, executing a single canonical partial amount from transaction state without yet supporting multi-round partial execution, escrow partial release, or percentage-based hints.

**Architecture:** Add a small `internal/partialsettlementexecution` service that gates execution on `approved-for-settlement`, resolves a single absolute partial amount from `transaction receipt.partial_settlement_hint`, reuses the existing direct payment runtime, moves progression to `partially-settled` on success, and writes a new canonical remaining-amount hint. On failure it keeps progression unchanged and records evidence in audit and submission receipt trail.

**Tech Stack:** Go, `internal/receipts`, `internal/settlementexecution`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/partialsettlementexecution/types.go`
  - Request/result types and deny reason model.
- Create: `internal/partialsettlementexecution/service.go`
  - Canonical execution gate, hint parsing, and one-shot partial execution orchestration.
- Create: `internal/partialsettlementexecution/service_test.go`
  - Focused tests for hint parsing, gating, success, and failure.
- Modify: `internal/receipts/types.go`
  - Add remaining-amount helper field if needed for canonicalization wording.
- Modify: `internal/receipts/store.go`
  - Add helpers to mark `partially-settled`, write remaining hint, and append partial execution evidence.
- Modify: `internal/receipts/store_test.go`
  - Cover partial success/failure transitions and trail evidence.
- Modify: `internal/app/tools_meta.go`
  - Add `execute_partial_settlement` meta tool and response payload.
- Modify: `internal/app/tools_parity_test.go`
  - Extend parity expectations for runtime-aware meta tools.
- Create: `internal/app/tools_meta_partialsettlementexecution_test.go`
  - Meta-tool coverage for `execute_partial_settlement`.
- Create: `docs/architecture/partial-settlement-execution.md`
  - Public architecture/operator doc for the first partial execution slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark partial settlement execution as landed and move remaining gaps down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/partial-settlement-execution/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `execute_partial_settlement` tool contract.

### Task 1: Add Partial Settlement Execution Service

**Files:**
- Create: `internal/partialsettlementexecution/types.go`
- Create: `internal/partialsettlementexecution/service.go`
- Create: `internal/partialsettlementexecution/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/partialsettlementexecution/service_test.go` with tests covering:

- execution denied when transaction receipt is missing
- execution denied when there is no current submission
- execution denied when progression is not `approved-for-settlement`
- execution denied when partial hint is missing
- execution denied when partial hint is invalid
- execution denied when already `partially-settled`
- execution success returns `partially-settled` shape with remaining hint
- runtime failure keeps `approved-for-settlement` shape

- [ ] **Step 2: Run the partial settlement tests and verify they fail**

Run:

```bash
go test ./internal/partialsettlementexecution/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/partialsettlementexecution: no Go files
```

- [ ] **Step 3: Implement the service**

Create:

- request type using `transaction_receipt_id`
- deny reasons:
  - `missing_receipt`
  - `no_current_submission`
  - `not_approved_for_settlement`
  - `partial_hint_missing`
  - `partial_hint_invalid`
  - `already_partially_settled`

The service must:

- load transaction and current submission
- require `approved-for-settlement`
- parse `partial_settlement_hint` only in form `settle:<amount>-usdc`
- resolve total amount from `price_context`
- calculate remaining amount after partial execution
- reuse a direct payment runtime abstraction
- return success shape with:
  - transaction receipt ID
  - submission receipt ID
  - `partially-settled`
  - executed partial amount
  - remaining amount
- return failure shape while keeping canonical progression unchanged

- [ ] **Step 4: Re-run the partial settlement tests and verify they pass**

Run:

```bash
go test ./internal/partialsettlementexecution/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the service slice**

Run:

```bash
git add internal/partialsettlementexecution/types.go internal/partialsettlementexecution/service.go internal/partialsettlementexecution/service_test.go
git -c commit.gpgsign=false commit -m "feat: add partial settlement execution service"
```

### Task 2: Extend Receipts for Partial Settlement Closeout

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add tests covering:

- `approved-for-settlement -> partially-settled` transition on successful partial execution
- remaining amount hint is canonicalized as a new absolute hint
- failure does not move progression away from `approved-for-settlement`
- partial success appends settlement success evidence
- partial failure appends failure evidence

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

In `internal/receipts/store.go`, add helpers so partial settlement execution can:

- mark settlement progression as `partially-settled`
- write the new remaining absolute hint
- append submission-trail evidence for partial success/failure

Canonical ownership must remain:

- transaction receipt owns progression and remaining hint
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
git -c commit.gpgsign=false commit -m "feat: add partial settlement receipt updates"
```

### Task 3: Add `execute_partial_settlement` Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_partialsettlementexecution_test.go`
- Modify: `internal/app/modules.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_partialsettlementexecution_test.go` with tests for:

- runtime-aware meta tool registration includes `execute_partial_settlement`
- the tool is absent when the partial-settlement runtime is unavailable
- approved path executes and returns canonical partial result
- missing or invalid partial hint path returns an error

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesExecutePartialSettlement|TestBuildMetaTools_OmitsExecutePartialSettlementWithoutRuntime|TestExecutePartialSettlement_' -count=1
```

Expected:

```text
FAIL
tool not found
```

- [ ] **Step 3: Implement the meta tool**

In `internal/app/tools_meta.go`, add:

- `execute_partial_settlement`
- required input: `transaction_receipt_id`
- handler using the partial settlement execution service
- thin receipt payload including:
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `settlement_progression_status`
  - `executed_amount`
  - `remaining_amount`

Update runtime-aware meta tool assembly and parity expectations.

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesExecutePartialSettlement|TestBuildMetaTools_OmitsExecutePartialSettlementWithoutRuntime|TestExecutePartialSettlement_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_partialsettlementexecution_test.go internal/app/modules.go
git -c commit.gpgsign=false commit -m "app: add execute partial settlement meta tool"
```

### Task 4: Document and Close Out the Slice

**Files:**
- Create: `docs/architecture/partial-settlement-execution.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Create: `openspec/changes/partial-settlement-execution/**`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`

- [ ] **Step 1: Write the public architecture page**

Create `docs/architecture/partial-settlement-execution.md` describing:

- what ships in the first direct partial settlement slice
- the canonical hint model
- success/failure semantics
- remaining amount canonicalization
- current limits

- [ ] **Step 2: Wire the page into architecture docs and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed partial settlement slice truthfully.

- [ ] **Step 3: Write and archive the OpenSpec change**

Create:

- `openspec/changes/partial-settlement-execution/proposal.md`
- `openspec/changes/partial-settlement-execution/design.md`
- `openspec/changes/partial-settlement-execution/tasks.md`
- delta specs for:
  - `project-docs`
  - `docs-only`
  - `meta-tools`

Then sync main specs and archive under:

```text
openspec/changes/archive/2026-04-22-partial-settlement-execution
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
git add docs/architecture/partial-settlement-execution.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-22-partial-settlement-execution
git -c commit.gpgsign=false commit -m "specs: archive partial settlement execution"
```

## Self-Review

- Spec coverage:
  - partial settlement execution service: Task 1
  - receipt closeout and remaining-hint canonicalization: Task 2
  - `execute_partial_settlement` meta tool: Task 3
  - public docs and OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or fake multi-round claims remain
- Type/path consistency:
  - partial execution stays transaction-owned
  - current submission remains the evidence anchor
  - remaining amount is always canonicalized as an absolute amount hint
