# Actual Settlement Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first direct `actual settlement execution` slice for `knowledge exchange v1`, connecting `approved-for-settlement` transaction state to real direct payment execution without yet implementing escrow release/refund or partial-settlement execution.

**Architecture:** Add a small `internal/settlementexecution` service that gates execution on canonical transaction state, resolves the current submission and amount context, then reuses the existing direct payment runtime. On success it closes settlement progression to `settled`; on failure it leaves progression unchanged and records execution evidence in audit and submission receipt trail.

**Tech Stack:** Go, `internal/receipts`, `internal/settlementprogression`, `internal/tools/payment`, `internal/app`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/settlementexecution/types.go`
  - Request/result types and deny reason model.
- Create: `internal/settlementexecution/service.go`
  - Canonical execution gate and direct settlement orchestration.
- Create: `internal/settlementexecution/service_test.go`
  - Focused tests for gating, success, failure, and receipt trail updates.
- Modify: `internal/receipts/store.go`
  - Add helpers to close settlement progression to `settled` while preserving execution-failure semantics.
- Modify: `internal/receipts/store_test.go`
  - Cover settled transition and settlement trail evidence.
- Modify: `internal/app/tools_meta.go`
  - Add `execute_settlement` meta tool and response payload.
- Modify: `internal/app/tools_parity_test.go`
  - Extend meta-tool parity expectations.
- Create: `internal/app/tools_meta_settlementexecution_test.go`
  - Meta-tool coverage for `execute_settlement`.
- Create: `docs/architecture/actual-settlement-execution.md`
  - Public architecture/operator doc for the first execution slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark actual settlement execution as landed and move remaining settlement gaps down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/actual-settlement-execution/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync `execute_settlement` tool contract.

### Task 1: Add Settlement Execution Service

**Files:**
- Create: `internal/settlementexecution/types.go`
- Create: `internal/settlementexecution/service.go`
- Create: `internal/settlementexecution/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/settlementexecution/service_test.go` with tests covering:

- execution denied when transaction receipt is missing
- execution denied when there is no current submission
- execution denied when settlement progression is not `approved-for-settlement`
- execution denied when amount cannot be resolved
- execution success closes progression to `settled`
- execution failure keeps progression at `approved-for-settlement`
- success/failure both append evidence to the current submission trail

- [ ] **Step 2: Run the settlement execution tests and verify they fail**

Run:

```bash
go test ./internal/settlementexecution/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/settlementexecution: no Go files
```

- [ ] **Step 3: Implement the service**

Create `internal/settlementexecution/types.go` with:

- request type containing `transaction_receipt_id`
- result type containing transaction receipt ID, submission receipt ID, settlement status, and resolved amount context
- deny reason constants:
  - `missing receipt`
  - `no current submission`
  - `not approved-for-settlement`
  - `amount unresolved`

Create `internal/settlementexecution/service.go` with:

- a receipt-store dependency for transaction and submission lookups and progression updates
- a direct-payment runtime dependency used for final settlement execution
- gate logic:
  - transaction must exist
  - current submission must exist
  - settlement progression must equal `approved-for-settlement`
  - amount must resolve from canonical transaction context
- success path:
  - execute direct settlement
  - move progression to `settled`
  - append success evidence to the submission trail
- failure path:
  - keep progression unchanged
  - append failure evidence to the submission trail

- [ ] **Step 4: Re-run the settlement execution tests and verify they pass**

Run:

```bash
go test ./internal/settlementexecution/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the service slice**

Run:

```bash
git add internal/settlementexecution/types.go internal/settlementexecution/service.go internal/settlementexecution/service_test.go
git -c commit.gpgsign=false commit -m "feat: add settlement execution service"
```

### Task 2: Extend Receipts for Final Settlement Closeout

**Files:**
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add tests covering:

- `approved-for-settlement -> settled` transition on successful execution
- failure does not move progression away from `approved-for-settlement`
- settlement success appends `settlement_updated`
- settlement failure appends a failure trail event without mutating progression

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

In `internal/receipts/store.go`, add or extend helpers so the settlement execution service can:

- close progression to `settled`
- keep `approved-for-settlement` on execution failure
- append submission-trail evidence for execution success/failure

Keep canonical ownership the same:

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
git -c commit.gpgsign=false commit -m "feat: add settlement execution receipt updates"
```

### Task 3: Add `execute_settlement` Meta Tool

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_settlementexecution_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_settlementexecution_test.go` with tests for:

- `buildMetaTools(...)` includes `execute_settlement`
- approve-for-settlement path executes and returns canonical result
- missing-current-submission or wrong-progression path returns an error

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesExecuteSettlement|TestExecuteSettlement_' -count=1
```

Expected:

```text
FAIL
tool not found
```

- [ ] **Step 3: Implement the meta tool**

In `internal/app/tools_meta.go`, add:

- `execute_settlement`
- required input: `transaction_receipt_id`
- handler using the settlement execution service
- thin receipt payload including:
  - `transaction_receipt_id`
  - `submission_receipt_id`
  - `settlement_progression_status`
  - `resolved_amount`

Update parity expectations in `internal/app/tools_parity_test.go`.

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesExecuteSettlement|TestExecuteSettlement_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_settlementexecution_test.go
git -c commit.gpgsign=false commit -m "app: add execute settlement meta tool"
```

### Task 4: Document and Close Out the Slice

**Files:**
- Create: `docs/architecture/actual-settlement-execution.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Create: `openspec/changes/actual-settlement-execution/**`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`

- [ ] **Step 1: Write the public architecture page**

Create `docs/architecture/actual-settlement-execution.md` describing:

- what ships in the first direct settlement execution slice
- canonical gate inputs and deny reasons
- success/failure semantics
- current limits

- [ ] **Step 2: Wire the page into architecture docs and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed actual settlement execution slice truthfully.

- [ ] **Step 3: Write and archive the OpenSpec change**

Create:

- `openspec/changes/actual-settlement-execution/proposal.md`
- `openspec/changes/actual-settlement-execution/design.md`
- `openspec/changes/actual-settlement-execution/tasks.md`
- delta specs for:
  - `project-docs`
  - `docs-only`
  - `meta-tools`

Then sync main specs and archive under:

```text
openspec/changes/archive/2026-04-22-actual-settlement-execution
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
git add docs/architecture/actual-settlement-execution.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-22-actual-settlement-execution
git -c commit.gpgsign=false commit -m "specs: archive actual settlement execution"
```

## Self-Review

- Spec coverage:
  - settlement execution service: Task 1
  - receipt closeout and trail evidence: Task 2
  - `execute_settlement` meta tool: Task 3
  - public docs and OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or fake execution claims remain
- Type/path consistency:
  - settlement execution stays transaction-owned
  - current submission remains the evidence anchor
  - public docs describe only the direct settlement first slice
