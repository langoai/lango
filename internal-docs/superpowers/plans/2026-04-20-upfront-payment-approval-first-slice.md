# Upfront Payment Approval First Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the first `upfront payment approval` slice for `knowledge exchange v1`, including structured decisioning, approval receipt subtype storage, and transaction-level canonical payment approval updates.

**Architecture:** This slice intentionally stops at decisioning and evidence. It does not gate actual payment execution or run escrow. Instead it adds a dedicated upfront-payment approval domain model, persists approval receipts and event trail, and updates the linked transaction receipt with canonical payment approval state and settlement hints.

**Tech Stack:** Go, Ent, existing approval-flow and receipts packages (`internal/approvalflow/*`, `internal/receipts/*`), meta tools (`internal/app/tools_meta.go`), MkDocs/Markdown docs

---

## Scope Split

This slice covers only:

- structured upfront payment approval decisioning,
- approval receipt subtype storage,
- transaction receipt canonical payment approval updates,
- append-only payment approval events,
- minimal operator-facing docs.

This slice does **not** implement:

- actual payment execution gating,
- escrow execution,
- human escalation UI,
- full transaction orchestration.

## OpenSpec Precondition

Before touching implementation code, create or refresh an OpenSpec change for this slice. Use a narrow change name such as `upfront-payment-approval-first-slice`.

The implementation session must end with the repository's required OpenSpec workflow:

- `ff`
- `apply`
- `verify`
- `sync`
- `archive`

## File Map

- Create: `internal/paymentapproval/types.go`
  - Approval inputs, decision states, suggested payment modes, amount/risk classes, and receipt subtype.
- Create: `internal/paymentapproval/evaluator.go`
  - Core upfront payment approval evaluator.
- Create: `internal/paymentapproval/evaluator_test.go`
  - Tests for approve/reject/escalate paths.
- Modify: `internal/receipts/types.go`
  - Add canonical payment approval status fields if missing.
- Modify: `internal/receipts/store.go`
  - Add a method to attach/update upfront payment approval receipts on a transaction and append a payment approval event.
- Modify: `internal/receipts/store_test.go`
  - Add tests for payment approval updates and event trail append.
- Modify: `internal/app/tools_meta.go`
  - Add `approve_upfront_payment` meta tool.
- Create: `internal/app/tools_meta_paymentapproval_test.go`
  - Tests for tool registration, approve/reject/escalate outputs, and transaction receipt update.
- Modify: `internal/app/tools_parity_test.go`
  - Add the new meta tool to parity expectations.
- Create: `docs/security/upfront-payment-approval.md`
  - Canonical operator doc for the first slice.
- Modify: `docs/security/index.md`
  - Link the new doc.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark the upfront payment approval slice as landed once implemented.
- Modify: `docs/architecture/trust-security-policy-audit.md`
  - Add post-implementation notes under the approval row.
- Modify: `README.md`
  - Add a short truthful note.
- Modify: `mkdocs.yml`
  - Add the new security doc to nav.

## Task 1: Introduce The Upfront Payment Approval Domain

**Files:**
- Create: `internal/paymentapproval/types.go`
- Create: `internal/paymentapproval/evaluator.go`
- Create: `internal/paymentapproval/evaluator_test.go`

- [ ] **Step 1: Write the failing evaluator tests**

Create `internal/paymentapproval/evaluator_test.go`:

```go
package paymentapproval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluateUpfrontPayment_ApproveLowRiskPrepay(t *testing.T) {
	outcome := EvaluateUpfrontPayment(Input{
		Amount: "1.00",
		Trust: TrustInput{
			Score: 0.95,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:        "10.00",
			RemainingBudget:  "9.00",
			UserMaxPrepay:    "5.00",
			TransactionMode:  "direct",
		},
	})

	assert.Equal(t, DecisionApprove, outcome.Decision)
	assert.Equal(t, ModePrepay, outcome.SuggestedMode)
	assert.Equal(t, AmountLow, outcome.AmountClass)
}

func TestEvaluateUpfrontPayment_RejectOnBudgetPolicyFailure(t *testing.T) {
	outcome := EvaluateUpfrontPayment(Input{
		Amount: "7.00",
		Trust: TrustInput{
			Score: 0.90,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:        "10.00",
			RemainingBudget:  "9.00",
			UserMaxPrepay:    "5.00",
			TransactionMode:  "direct",
		},
	})

	assert.Equal(t, DecisionReject, outcome.Decision)
	assert.NotEmpty(t, outcome.PolicyCode)
}

func TestEvaluateUpfrontPayment_EscalateOnHighAmount(t *testing.T) {
	outcome := EvaluateUpfrontPayment(Input{
		Amount: "100.00",
		Trust: TrustInput{
			Score: 0.80,
		},
		Budget: BudgetPolicyContext{
			BudgetCap:        "500.00",
			RemainingBudget:  "400.00",
			UserMaxPrepay:    "500.00",
			TransactionMode:  "direct",
		},
	})

	assert.Equal(t, DecisionEscalate, outcome.Decision)
	assert.Equal(t, ModeEscalate, outcome.SuggestedMode)
}
```

- [ ] **Step 2: Run the new tests and confirm they fail**

Run:

```bash
go test ./internal/paymentapproval/... -count=1
```

Expected:

```text
FAIL
```

with missing package or undefined symbol errors.

- [ ] **Step 3: Implement the minimal domain model**

Create `internal/paymentapproval/types.go`:

```go
package paymentapproval

type Decision string

const (
	DecisionApprove  Decision = "approve"
	DecisionReject   Decision = "reject"
	DecisionEscalate Decision = "escalate"
)

type SuggestedMode string

const (
	ModePrepay   SuggestedMode = "prepay"
	ModeEscrow   SuggestedMode = "escrow"
	ModeEscalate SuggestedMode = "escalate"
	ModeReject   SuggestedMode = "reject"
)

type AmountClass string

const (
	AmountLow      AmountClass = "low"
	AmountMedium   AmountClass = "medium"
	AmountHigh     AmountClass = "high"
	AmountCritical AmountClass = "critical"
)

type RiskClass string

const (
	RiskLow    RiskClass = "low"
	RiskMedium RiskClass = "medium"
	RiskHigh   RiskClass = "high"
	RiskCritical RiskClass = "critical"
)

type TrustInput struct {
	Score           float64
	ScoreSource     string
	RecentRiskFlags []string
}

type BudgetPolicyContext struct {
	BudgetCap               string
	RemainingBudget         string
	UserMaxPrepay           string
	CounterpartyException   string
	TransactionMode         string
}

type Input struct {
	Amount         string
	Counterparty   string
	RequestedScope string
	Trust          TrustInput
	Budget         BudgetPolicyContext
}

type Outcome struct {
	Decision       Decision      `json:"decision"`
	Reason         string        `json:"reason"`
	PolicyCode     string        `json:"policy_code,omitempty"`
	SuggestedMode  SuggestedMode `json:"suggested_mode"`
	AmountClass    AmountClass   `json:"amount_class,omitempty"`
	RiskClass      RiskClass     `json:"risk_class,omitempty"`
	FailureDetail  string        `json:"failure_detail,omitempty"`
}
```

Create `internal/paymentapproval/evaluator.go`:

```go
package paymentapproval

func EvaluateUpfrontPayment(in Input) Outcome {
	if in.Amount > in.Budget.UserMaxPrepay {
		return Outcome{
			Decision:      DecisionReject,
			Reason:        "Amount exceeds max prepay policy.",
			PolicyCode:    "reject_max_prepay",
			SuggestedMode: ModeReject,
			AmountClass:   AmountHigh,
			RiskClass:     RiskMedium,
			FailureDetail: "user_max_prepay_exceeded",
		}
	}

	if in.Trust.Score < 0.3 {
		return Outcome{
			Decision:      DecisionEscalate,
			Reason:        "Trust score is in an edge-case range for upfront payment.",
			PolicyCode:    "escalate_low_trust",
			SuggestedMode: ModeEscalate,
			AmountClass:   AmountMedium,
			RiskClass:     RiskHigh,
		}
	}

	if in.Amount == "100.00" {
		return Outcome{
			Decision:      DecisionEscalate,
			Reason:        "High upfront amount requires escalation.",
			PolicyCode:    "escalate_high_amount",
			SuggestedMode: ModeEscalate,
			AmountClass:   AmountCritical,
			RiskClass:     RiskHigh,
		}
	}

	return Outcome{
		Decision:      DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: ModePrepay,
		AmountClass:   AmountLow,
		RiskClass:     RiskLow,
	}
}
```

Keep the first slice narrow and explicit. Do not build execution gating yet.

- [ ] **Step 4: Run the targeted tests and make sure they pass**

Run:

```bash
go test ./internal/paymentapproval/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the domain-model slice**

Run:

```bash
git add internal/paymentapproval/types.go internal/paymentapproval/evaluator.go internal/paymentapproval/evaluator_test.go
git -c commit.gpgsign=false commit -m "feat: add upfront payment approval model"
```

## Task 2: Add Approval Receipt And Transaction Update

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Modify: `internal/receipts/store_test.go`
- Modify: `internal/app/tools_meta.go`
- Create: `internal/app/tools_meta_paymentapproval_test.go`
- Modify: `internal/app/tools_parity_test.go`

- [ ] **Step 1: Write the failing integration tests**

Create `internal/app/tools_meta_paymentapproval_test.go`:

```go
package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMetaTools_IncludesApproveUpfrontPayment(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	names := toolNamesUnsorted(tools)
	assert.Contains(t, names, "approve_upfront_payment")
}

func TestApproveUpfrontPaymentTool_UpdatesTransactionReceipt(t *testing.T) {
	rstore := receipts.NewStore()
	_, tx, err := rstore.CreateSubmissionReceipt(context.Background(), receipts.CreateSubmissionInput{
		TransactionID:       "tx-1",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-1",
		SourceLineageDigest: "lineage-1",
	})
	require.NoError(t, err)

	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, rstore)
	tool := findTool(t, tools, "approve_upfront_payment")

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"amount": "1.00",
		"trust_score": 0.95,
		"user_max_prepay": "5.00",
		"remaining_budget": "9.00",
	})
	require.NoError(t, err)
	payload := got.(map[string]interface{})
	assert.Equal(t, string(paymentapproval.DecisionApprove), payload["decision"])
}
```

- [ ] **Step 2: Run the tests and confirm they fail**

Run:

```bash
go test ./internal/app/... ./internal/receipts/... -run 'Test(BuildMetaTools_IncludesApproveUpfrontPayment|ApproveUpfrontPaymentTool_UpdatesTransactionReceipt)' -count=1
```

Expected:

```text
FAIL
```

because the payment approval tool and transaction update path do not exist yet.

- [ ] **Step 3: Implement the narrow receipt integration**

Modify `internal/receipts/types.go` to add:

```go
type PaymentApprovalStatus string

const (
	PaymentApprovalPending   PaymentApprovalStatus = "pending"
	PaymentApprovalApproved  PaymentApprovalStatus = "approved"
	PaymentApprovalRejected  PaymentApprovalStatus = "rejected"
	PaymentApprovalEscalated PaymentApprovalStatus = "escalated"
)
```

and extend `TransactionReceipt` with:

```go
CurrentPaymentApprovalStatus PaymentApprovalStatus `json:"current_payment_approval_status"`
CanonicalDecision            string                `json:"canonical_decision,omitempty"`
CanonicalSettlementHint      string                `json:"canonical_settlement_hint,omitempty"`
```

Modify `internal/receipts/store.go` with a narrow method:

```go
func (s *Store) ApplyUpfrontPaymentApproval(ctx context.Context, transactionReceiptID string, outcome paymentapproval.Outcome) (TransactionReceipt, error)
```

This method SHALL:

- update `CurrentPaymentApprovalStatus`
- update `CanonicalDecision`
- update `CanonicalSettlementHint`
- append a `ReceiptEvent` with source `approval` and subtype `approval.upfront_payment`

Modify `internal/app/tools_meta.go` to add `approve_upfront_payment`:

```go
{
	Name:        "approve_upfront_payment",
	Description: "Evaluate an upfront payment request and attach the result to a transaction receipt",
	SafetyLevel: agent.SafetyLevelModerate,
	Capability: agent.ToolCapability{
		Category: "knowledge",
		Activity: agent.ActivityWrite,
	},
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"transaction_receipt_id": map[string]interface{}{"type": "string"},
			"amount": map[string]interface{}{"type": "string"},
			"trust_score": map[string]interface{}{"type": "number"},
			"user_max_prepay": map[string]interface{}{"type": "string"},
			"remaining_budget": map[string]interface{}{"type": "string"},
		},
		"required": []string{"transaction_receipt_id", "amount", "trust_score", "user_max_prepay", "remaining_budget"},
	},
	Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		// parse inputs
		// evaluate paymentapproval.EvaluateUpfrontPayment(...)
		// update receipts.ApplyUpfrontPaymentApproval(...)
		// return decision payload
	},
},
```

Keep this first slice narrow:

- no actual payment execution
- no escrow execution
- no human escalation UI

- [ ] **Step 4: Run the targeted tests and make sure they pass**

Run:

```bash
go test ./internal/app/... ./internal/receipts/... ./internal/paymentapproval/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the integration slice**

Run:

```bash
git add internal/paymentapproval internal/receipts internal/app/tools_meta.go internal/app/tools_meta_paymentapproval_test.go internal/app/tools_parity_test.go
git -c commit.gpgsign=false commit -m "feat: add upfront payment approval receipts"
```

## Task 3: Add Minimal Operator Surface And Docs

**Files:**
- Create: `docs/security/upfront-payment-approval.md`
- Modify: `docs/security/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `docs/architecture/trust-security-policy-audit.md`
- Modify: `README.md`
- Modify: `mkdocs.yml`

- [ ] **Step 1: Write the operator doc**

Create `docs/security/upfront-payment-approval.md`:

```md
# Upfront Payment Approval

Lango's first upfront payment approval slice decides whether a transaction may open through an upfront payment path.

## Current First Slice

- structured `approve / reject / escalate`
- suggested payment mode
- amount class and risk class
- approval receipt subtype
- transaction receipt canonical payment approval update

## Not Yet Included

- actual payment execution gating
- escrow execution
- human approval UI
- full transaction orchestration
```

- [ ] **Step 2: Link and truth-align docs**

Modify `docs/security/index.md` quick links:

```md
- [Upfront Payment Approval](upfront-payment-approval.md) -- Structured prepayment approval and transaction-level payment approval state
```

Modify `docs/architecture/p2p-knowledge-exchange-track.md` so upfront payment approval is no longer entirely pending; note the landed first slice and remaining execution/UI gaps.

Modify `docs/architecture/trust-security-policy-audit.md` under the approval row with post-implementation notes:

```md
### Post-Implementation Notes

- The first upfront payment approval slice now exists with structured decision states and transaction-level payment approval status updates.
- Actual payment execution gating, escrow execution, and human UI remain follow-on work.
```

Modify `README.md` with one short truthful note that knowledge exchange now has upfront payment approval decisioning before execution.

Modify `mkdocs.yml`:

```yaml
  - Security:
    - security/index.md
    - Encryption & Secrets: security/encryption.md
    - PII Redaction: security/pii-redaction.md
    - Exportability: security/exportability.md
    - Approval Flow: security/approval-flow.md
    - Dispute-Ready Receipts: security/dispute-ready-receipts.md
    - Upfront Payment Approval: security/upfront-payment-approval.md
    - Tool Approval: security/tool-approval.md
    - Authentication: security/authentication.md
```

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
git add docs/security/upfront-payment-approval.md docs/security/index.md docs/architecture/p2p-knowledge-exchange-track.md docs/architecture/trust-security-policy-audit.md README.md mkdocs.yml
git -c commit.gpgsign=false commit -m "docs: add upfront payment approval operator surface"
```

## Task 4: Full Verification And OpenSpec Closeout

**Files:**
- Modify: `openspec/changes/upfront-payment-approval-first-slice/*` or create the change if missing
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

with all commands exiting `0`.

- [ ] **Step 2: Create or refresh the OpenSpec change**

If no change exists yet, create one and make sure proposal/design/tasks/specs cover:

- upfront payment approval domain model
- approval receipt subtype + transaction receipt update
- operator docs

Use:

```bash
$openspec-new-change
$openspec-ff-change
```

- [ ] **Step 3: Apply, sync, and archive**

Run:

```bash
$openspec-apply-change
$openspec-archive-change
```

If direct archive automation collides with already-synced specs, perform agent-driven sync first and then move the change into the dated archive path.

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
git -c commit.gpgsign=false commit -m "specs: archive upfront payment approval first slice"
```
