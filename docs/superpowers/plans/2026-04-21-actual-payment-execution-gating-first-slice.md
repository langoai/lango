# Actual Payment Execution Gating First Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the first actual payment execution gate for direct payment paths, using receipt-backed canonical state to allow or deny `payment_send` and `p2p_pay`.

**Architecture:** This slice keeps the integration points in the two payment handlers but centralizes decision logic in one shared gate service. The gate consumes transaction receipt canonical payment approval state, enforces direct-prepay-only execution, and records allow/deny outcomes into both audit and receipt trails. It does not execute escrow or introduce human UI.

**Tech Stack:** Go, existing payment tools (`internal/tools/payment/*`, `internal/app/tools_p2p.go`), receipts store (`internal/receipts/*`), audit surfaces, MkDocs/Markdown docs

---

## Scope Split

This first slice covers only:

- a shared direct-payment execution gate service
- handler integration for `payment_send`
- handler integration for `p2p_pay`
- structured allow/deny reason codes
- audit and receipt-trail event recording
- minimal operator docs

This slice does **not** implement:

- escrow execution
- human payment approval UI
- broader middleware-wide enforcement
- complete transaction orchestration

## OpenSpec Precondition

Before touching implementation code, create or refresh an OpenSpec change for this slice. Use a narrow change name such as `actual-payment-execution-gating-first-slice`.

The implementation session must end with the repository's required OpenSpec workflow:

- `ff`
- `apply`
- `verify`
- `sync`
- `archive`

## File Map

- Create: `internal/paymentgate/types.go`
  - Decision results, deny reason codes, and event payload shapes.
- Create: `internal/paymentgate/service.go`
  - Shared gate service for direct payment execution.
- Create: `internal/paymentgate/service_test.go`
  - Unit tests for allow/deny decisions.
- Modify: `internal/receipts/types.go`
  - Add event constants or event payload support for execution authorization/denial if needed.
- Modify: `internal/receipts/store.go`
  - Add helper(s) to append execution events against a specific receipt context.
- Modify: `internal/receipts/store_test.go`
  - Add tests for allow/deny event append behavior.
- Modify: `internal/tools/payment/payment.go`
  - Gate `payment_send` before actual send.
- Modify: `internal/app/tools_p2p.go`
  - Gate `p2p_pay` before actual send.
- Modify: `internal/app/modules.go` or related wiring
  - Provide the shared payment gate service to the tool builders.
- Create: `docs/security/actual-payment-execution-gating.md`
  - Canonical operator doc for the first slice.
- Modify: `docs/security/index.md`
  - Link the new doc.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark direct payment execution gating as landed for the direct payment path.
- Modify: `docs/architecture/trust-security-policy-audit.md`
  - Add post-implementation notes under the approval/auditability row.
- Modify: `README.md`
  - Add a short truthful note.
- Modify: `mkdocs.yml`
  - Add the new security doc to nav.

## Task 1: Introduce The Shared Payment Gate Service

**Files:**
- Create: `internal/paymentgate/types.go`
- Create: `internal/paymentgate/service.go`
- Create: `internal/paymentgate/service_test.go`

- [ ] **Step 1: Write the failing payment gate tests**

Create `internal/paymentgate/service_test.go`:

```go
package paymentgate

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/receipts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateDirectPayment_AllowsApprovedPrepay(t *testing.T) {
	rstore := receipts.NewStore()
	ctx := context.Background()

	sub, tx, err := rstore.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-allow",
		ArtifactLabel:       "artifact",
		PayloadHash:         "hash-1",
		SourceLineageDigest: "lineage-1",
	})
	require.NoError(t, err)

	tx.CurrentPaymentApprovalStatus = receipts.PaymentApprovalApproved
	tx.CanonicalDecision = "approve"
	tx.CanonicalSettlementHint = "prepay"
	rstore.ForceSetTransactionForTest(tx)

	svc := NewService(rstore)
	got, err := svc.EvaluateDirectPayment(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  sub.SubmissionReceiptID,
		ToolName:             "payment_send",
	})
	require.NoError(t, err)
	assert.Equal(t, Allow, got.Decision)
}

func TestEvaluateDirectPayment_DeniesWithoutTransactionReceiptID(t *testing.T) {
	svc := NewService(receipts.NewStore())
	got, err := svc.EvaluateDirectPayment(context.Background(), Request{
		ToolName: "payment_send",
	})
	require.NoError(t, err)
	assert.Equal(t, Deny, got.Decision)
	assert.Equal(t, ReasonMissingReceipt, got.Reason)
}
```

- [ ] **Step 2: Run the tests and confirm they fail**

Run:

```bash
go test ./internal/paymentgate/... -count=1
```

Expected:

```text
FAIL
```

because the package and symbols do not exist yet.

- [ ] **Step 3: Implement the gate service**

Create `internal/paymentgate/types.go`:

```go
package paymentgate

type Decision string

const (
	Allow Decision = "allow"
	Deny  Decision = "deny"
)

type DenyReason string

const (
	ReasonMissingReceipt      DenyReason = "missing_receipt"
	ReasonApprovalNotApproved DenyReason = "approval_not_approved"
	ReasonStaleState          DenyReason = "stale_state"
	ReasonExecutionModeMismatch DenyReason = "execution_mode_mismatch"
)

type Request struct {
	TransactionReceiptID string
	SubmissionReceiptID  string
	ToolName             string
	Context              map[string]interface{}
}

type Result struct {
	Decision Decision
	Reason   DenyReason
}
```

Create `internal/paymentgate/service.go`:

```go
package paymentgate

import (
	"context"

	"github.com/langoai/lango/internal/receipts"
)

type Service struct {
	store *receipts.Store
}

func NewService(store *receipts.Store) *Service {
	return &Service{store: store}
}

func (s *Service) EvaluateDirectPayment(_ context.Context, req Request) (Result, error) {
	if req.TransactionReceiptID == "" {
		return Result{Decision: Deny, Reason: ReasonMissingReceipt}, nil
	}
	tx, ok := s.store.GetTransactionForTest(req.TransactionReceiptID)
	if !ok {
		return Result{Decision: Deny, Reason: ReasonMissingReceipt}, nil
	}
	if tx.CurrentPaymentApprovalStatus != receipts.PaymentApprovalApproved {
		return Result{Decision: Deny, Reason: ReasonApprovalNotApproved}, nil
	}
	if tx.CanonicalSettlementHint != "prepay" {
		return Result{Decision: Deny, Reason: ReasonExecutionModeMismatch}, nil
	}
	return Result{Decision: Allow}, nil
}
```

Add minimal test-only helpers in `internal/receipts/store.go` if needed to keep the first slice narrow:

```go
func (s *Store) ForceSetTransactionForTest(tx TransactionReceipt) { ... }
func (s *Store) GetTransactionForTest(id string) (TransactionReceipt, bool) { ... }
```

- [ ] **Step 4: Run the targeted tests and make sure they pass**

Run:

```bash
go test ./internal/paymentgate/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the gate service slice**

Run:

```bash
git add internal/paymentgate internal/receipts
git -c commit.gpgsign=false commit -m "feat: add direct payment gate service"
```

## Task 2: Gate The Direct Payment Tools

**Files:**
- Modify: `internal/tools/payment/payment.go`
- Modify: `internal/app/tools_p2p.go`
- Modify: wiring files that provide the gate service
- Add or modify tests near those tool surfaces

- [ ] **Step 1: Write the failing handler-level tests**

Add tests that prove:

- `payment_send` denies without `transaction_receipt_id`
- `payment_send` denies when canonical payment approval is not approved
- `p2p_pay` denies when canonical settlement hint is not `prepay`
- `allow` path records success events

- [ ] **Step 2: Implement handler integration**

Wire the shared service into:

- `payment_send`
- `p2p_pay`

Each handler should:

1. parse `transaction_receipt_id`
2. call the shared gate service
3. on deny:
   - return a structured error
   - write `payment execution denied` to audit + receipt trail
4. on allow:
   - write `payment execution authorized` to audit + receipt trail
   - continue into the existing send path

Keep the handlers thin by pushing common logic into the shared service or a helper.

- [ ] **Step 3: Run targeted tool tests**

Run:

```bash
go test ./internal/tools/payment/... ./internal/app/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 4: Commit the handler integration slice**

Run:

```bash
git add internal/tools/payment internal/app
git -c commit.gpgsign=false commit -m "feat: gate direct payment execution"
```

## Task 3: Add Minimal Operator Surface And Docs

**Files:**
- Create: `docs/security/actual-payment-execution-gating.md`
- Modify: `docs/security/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `docs/architecture/trust-security-policy-audit.md`
- Modify: `README.md`
- Modify: `mkdocs.yml`

- [ ] **Step 1: Write the operator doc**

Create `docs/security/actual-payment-execution-gating.md`:

```md
# Actual Payment Execution Gating

Lango's first actual payment execution gating slice enforces receipt-backed direct payment execution for `payment_send` and `p2p_pay`.

## Current First Slice

- shared direct payment gate service
- `allow / deny`
- deny reasons:
  - `missing_receipt`
  - `approval_not_approved`
  - `stale_state`
  - `execution_mode_mismatch`
- allow/deny events written to audit and receipt trail

## Not Yet Included

- escrow execution gating
- human payment approval UI
- full transaction orchestration
- middleware-wide enforcement
```

- [ ] **Step 2: Link and truth-align docs**

Modify `docs/security/index.md` quick links:

```md
- [Actual Payment Execution Gating](actual-payment-execution-gating.md) -- Receipt-backed allow/deny control for direct payment execution
```

Modify `docs/architecture/p2p-knowledge-exchange-track.md` so the direct payment execution path is no longer entirely pending; note the landed direct-payment gate and remaining escrow/UI gaps.

Modify `docs/architecture/trust-security-policy-audit.md` with post-implementation notes under the approval/auditability area:

```md
### Post-Implementation Notes

- Direct payment execution now has a receipt-backed allow/deny gate for `payment_send` and `p2p_pay`.
- Escrow execution, human UI, and full transaction orchestration remain follow-on work.
```

Modify `README.md` with one short truthful note that direct payment execution is now gated by canonical payment approval state.

Modify `mkdocs.yml` to add the new doc under Security.

- [ ] **Step 3: Run docs verification**

Run:

```bash
python3 -m mkdocs build --strict
```

Expected:

```text
Documentation built
```

with exit code `0`.

- [ ] **Step 4: Commit the docs slice**

Run:

```bash
git add docs/security/actual-payment-execution-gating.md docs/security/index.md docs/architecture/p2p-knowledge-exchange-track.md docs/architecture/trust-security-policy-audit.md README.md mkdocs.yml
git -c commit.gpgsign=false commit -m "docs: add payment execution gate surface"
```

## Task 4: Full Verification And OpenSpec Closeout

**Files:**
- Modify: `openspec/changes/actual-payment-execution-gating-first-slice/*` or create the change if missing
- Sync: `openspec/specs/*` as required by the implemented delta

- [ ] **Step 1: Verify the full repository**

Run:

```bash
go test ./...
go build ./...
python3 -m mkdocs build --strict
```

Expected:

```text
ok
```

- [ ] **Step 2: Create or refresh the OpenSpec change**

If no change exists yet, create one and make sure proposal/design/tasks/specs cover:

- direct payment execution gate service
- payment tool handler integration
- operator docs

- [ ] **Step 3: Apply, sync, and archive**

Run the repository's OpenSpec workflow and archive after sync.

- [ ] **Step 4: Confirm a clean worktree**

Run:

```bash
git status --short
```

Expected:

```text
[no output]
```

- [ ] **Step 5: Commit OpenSpec closeout if needed**

Run:

```bash
git add openspec/specs openspec/changes/archive
git -c commit.gpgsign=false commit -m "specs: archive actual payment execution gating first slice"
```
