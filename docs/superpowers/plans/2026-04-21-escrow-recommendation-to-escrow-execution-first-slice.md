# Escrow Recommendation To Escrow Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn an approved `escrow` recommendation on a transaction receipt into a real `create + fund` escrow execution path, with canonical receipt evidence preserved.

**Architecture:** Bind escrow execution inputs onto the transaction receipt during upfront payment approval, execute the recommendation through a dedicated `internal/escrowexecution` service, and let the receipt store remain the canonical evidence surface for transaction/submission updates. Keep the first slice narrow: one new meta tool, one execution service, receipt state coupling, truthful docs, and OpenSpec closeout.

**Tech Stack:** Go, `internal/receipts`, `internal/paymentapproval`, `internal/economy/escrow`, `internal/app/tools_meta.go`, MkDocs, OpenSpec

---

## File Map

- `internal/receipts/types.go`
  - Add escrow execution status, escrow execution input, escrow reference, and escrow execution event types.
- `internal/receipts/store.go`
  - Add methods to bind escrow execution input and apply escrow execution progress to transaction and submission receipts.
- `internal/receipts/store_test.go`
  - Cover escrow input binding, canonical status updates, event append behavior, and failure persistence.
- `internal/paymentapproval/types.go`
  - Reuse `ModeEscrow`; no new mode type required.
- `internal/escrowexecution/types.go`
  - New request/result types for the runtime service.
- `internal/escrowexecution/service.go`
  - New service that validates receipt state, calls escrow `Create`/`Fund`, and delegates canonical updates back to the receipt store.
- `internal/escrowexecution/service_test.go`
  - Focused tests for allow, deny, create failure, and fund failure paths.
- `internal/app/tools_meta.go`
  - Extend `approve_upfront_payment` to bind escrow execution input when the suggested mode is `escrow`.
  - Add `buildMetaToolsWithEscrow(...)` helper so existing tests can keep calling `buildMetaTools(...)`.
  - Add the new `execute_escrow_recommendation` meta tool.
- `internal/app/modules.go`
  - Register the new meta tool with the real escrow engine.
- `internal/app/tools_meta_paymentapproval_test.go`
  - Verify `approve_upfront_payment` stores escrow execution input when escrow is recommended.
- `internal/app/tools_meta_escrowexecution_test.go`
  - New tests for tool wiring, validation, success payloads, and failure propagation.
- `internal/app/tools_parity_test.go`
  - Update the expected meta tool set.
- `docs/security/escrow-execution.md`
  - New operator-facing escrow execution document.
- `docs/security/index.md`
  - Add quick link and summary row for escrow execution.
- `docs/security/upfront-payment-approval.md`
  - Truth-align the follow-on work section now that escrow execution exists.
- `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark escrow recommendation execution as landed for the first slice and list remaining gaps.
- `README.md`
  - Add one short truthful note that escrow-backed execution now exists for escrow-recommended knowledge-exchange transactions.
- `mkdocs.yml`
  - Add the new security doc to nav.
- `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/...`
  - New change artifacts.
- `openspec/specs/escrow-execution/spec.md`
  - New main spec after sync.
- `openspec/specs/dispute-ready-receipts/spec.md`
  - Add escrow execution trail requirements.
- `openspec/specs/upfront-payment-approval/spec.md`
  - Add requirement for binding escrow execution input onto receipt-backed transactions.
- `openspec/specs/security-docs-sync/spec.md`
  - Add escrow execution operator-doc requirements.

### Task 1: Extend Receipts To Store Escrow Execution Input And State

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Test: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add these tests to `internal/receipts/store_test.go`:

```go
func TestBindEscrowExecutionInput_PersistsCanonicalInputOnTransaction(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-bind",
		ArtifactLabel:       "artifact/escrow-bind",
		PayloadHash:         "hash-escrow-bind",
		SourceLineageDigest: "lineage-escrow-bind",
	})
	require.NoError(t, err)

	updated, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "3.50",
		Reason:    "knowledge exchange",
		TaskID:    "task-escrow-bind",
		Milestones: []EscrowMilestoneInput{
			{Description: "draft", Amount: "1.50"},
			{Description: "final", Amount: "2.00"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updated.EscrowExecutionInput)
	assert.Equal(t, EscrowExecutionStatusPending, updated.EscrowExecutionStatus)
	assert.Equal(t, "did:lango:buyer", updated.EscrowExecutionInput.BuyerDID)
	assert.Equal(t, "3.50", updated.EscrowExecutionInput.Amount)
}

func TestApplyEscrowExecutionProgress_RecordsCreatedFundedAndFailed(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-progress",
		ArtifactLabel:       "artifact/escrow-progress",
		PayloadHash:         "hash-escrow-progress",
		SourceLineageDigest: "lineage-escrow-progress",
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "knowledge exchange",
		TaskID:    "task-escrow-progress",
		Milestones: []EscrowMilestoneInput{
			{Description: "delivery", Amount: "4.00"},
		},
	})
	require.NoError(t, err)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "escrow-1", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	updated, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-1", EventEscrowExecutionFunded, "")
	require.NoError(t, err)

	assert.Equal(t, EscrowExecutionStatusFunded, updated.EscrowExecutionStatus)
	assert.Equal(t, "escrow-1", updated.EscrowReference)

	_, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, EventEscrowExecutionCreated, events[0].Type)
	assert.Equal(t, EventEscrowExecutionFunded, events[1].Type)
}
```

- [ ] **Step 2: Run the receipt tests and verify they fail**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
FAIL
undefined: EscrowExecutionInput
```

- [ ] **Step 3: Add the receipt model and store methods**

In `internal/receipts/types.go`, add:

```go
type EscrowExecutionStatus string

const (
	EscrowExecutionStatusPending EscrowExecutionStatus = "pending"
	EscrowExecutionStatusCreated EscrowExecutionStatus = "created"
	EscrowExecutionStatusFunded  EscrowExecutionStatus = "funded"
	EscrowExecutionStatusFailed  EscrowExecutionStatus = "failed"
)

type EscrowMilestoneInput struct {
	Description string `json:"description"`
	Amount      string `json:"amount"`
}

type EscrowExecutionInput struct {
	BuyerDID   string                `json:"buyer_did"`
	SellerDID  string                `json:"seller_did"`
	Amount     string                `json:"amount"`
	Reason     string                `json:"reason"`
	TaskID     string                `json:"task_id,omitempty"`
	Milestones []EscrowMilestoneInput `json:"milestones"`
}
```

Extend `TransactionReceipt` and add new event types:

```go
const (
	EventEscrowExecutionStarted EventType = "escrow_execution_started"
	EventEscrowExecutionCreated EventType = "escrow_execution_created"
	EventEscrowExecutionFunded  EventType = "escrow_execution_funded"
	EventEscrowExecutionFailed  EventType = "escrow_execution_failed"
)

type TransactionReceipt struct {
	TransactionReceiptID         string                `json:"transaction_receipt_id"`
	TransactionID                string                `json:"transaction_id"`
	CurrentSubmissionReceiptID   string                `json:"current_submission_receipt_id,omitempty"`
	CanonicalApprovalStatus      ApprovalStatus        `json:"canonical_approval_status"`
	CanonicalSettlementStatus    SettlementStatus      `json:"canonical_settlement_status"`
	CurrentPaymentApprovalStatus PaymentApprovalStatus `json:"current_payment_approval_status"`
	CanonicalDecision            string                `json:"canonical_decision,omitempty"`
	CanonicalSettlementHint      string                `json:"canonical_settlement_hint,omitempty"`
	EscrowExecutionStatus        EscrowExecutionStatus `json:"escrow_execution_status,omitempty"`
	EscrowReference              string                `json:"escrow_reference,omitempty"`
	EscrowExecutionInput         *EscrowExecutionInput `json:"escrow_execution_input,omitempty"`
}
```

Update `validateEventType(...)` in `internal/receipts/store.go` so these new escrow execution events are accepted:

```go
case EventDraftExportability,
	EventFinalExportability,
	EventApprovalRequested,
	EventApprovalResolved,
	EventPaymentApproval,
	EventPaymentExecutionAuthorized,
	EventPaymentExecutionDenied,
	EventEscrowExecutionStarted,
	EventEscrowExecutionCreated,
	EventEscrowExecutionFunded,
	EventEscrowExecutionFailed,
	EventSettlementUpdated,
	EventEscalated,
	EventDisputed:
	return nil
```

In `internal/receipts/store.go`, add:

```go
func (s *Store) BindEscrowExecutionInput(_ context.Context, transactionReceiptID, submissionReceiptID string, input EscrowExecutionInput) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[submissionReceiptID]
	if !ok || submission.TransactionReceiptID != transactionReceiptID {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}

	transaction.EscrowExecutionInput = &input
	transaction.EscrowExecutionStatus = EscrowExecutionStatusPending
	s.transactions[transactionReceiptID] = transaction
	return transaction, nil
}

func (s *Store) ApplyEscrowExecutionProgress(_ context.Context, transactionReceiptID, submissionReceiptID string, status EscrowExecutionStatus, escrowReference string, eventType EventType, reason string) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[submissionReceiptID]
	if !ok || submission.TransactionReceiptID != transactionReceiptID {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}

	transaction.EscrowExecutionStatus = status
	if escrowReference != "" {
		transaction.EscrowReference = escrowReference
	}
	s.transactions[transactionReceiptID] = transaction
	s.events[submissionReceiptID] = append(s.events[submissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: submissionReceiptID,
		Source:              "escrow_execution",
		Subtype:             string(status),
		Reason:              reason,
		Type:                eventType,
	})
	return transaction, nil
}
```

- [ ] **Step 4: Re-run the receipt tests and verify they pass**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the receipt model slice**

Run:

```bash
git add internal/receipts/types.go internal/receipts/store.go internal/receipts/store_test.go
git -c commit.gpgsign=false commit -m "feat: store escrow execution receipt state"
```

### Task 2: Bind Escrow Execution Input During Upfront Payment Approval

**Files:**
- Modify: `internal/app/tools_meta.go`
- Test: `internal/app/tools_meta_paymentapproval_test.go`

- [ ] **Step 1: Write the failing approval-tool tests**

Add to `internal/app/tools_meta_paymentapproval_test.go`:

```go
func TestApproveUpfrontPayment_BindsEscrowExecutionInputWhenModeEscrow(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-upfront-escrow",
		ArtifactLabel:       "artifact/upfront-escrow",
		PayloadHash:         "hash-upfront-escrow",
		SourceLineageDigest: "lineage-upfront-escrow",
	})
	require.NoError(t, err)

	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store)
	tool := findTool(tools, "approve_upfront_payment")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"submission_receipt_id":  submission.SubmissionReceiptID,
		"amount":                 "25.00",
		"trust_score":            0.30,
		"user_max_prepay":        "5.00",
		"remaining_budget":       "50.00",
		"escrow_buyer_did":       "did:lango:buyer",
		"escrow_seller_did":      "did:lango:seller",
		"escrow_reason":          "knowledge exchange",
		"escrow_task_id":         "task-upfront-escrow",
		"escrow_milestones": []interface{}{
			map[string]interface{}{"description": "draft", "amount": "10.00"},
			map[string]interface{}{"description": "final", "amount": "15.00"},
		},
	})
	require.NoError(t, err)

	payload := got.(upfrontPaymentApprovalReceipt)
	assert.Equal(t, "escrow", payload.SuggestedMode)
	assert.Equal(t, string(receipts.EscrowExecutionStatusPending), payload.EscrowExecutionStatus)

	updatedTx, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.NotNil(t, updatedTx.EscrowExecutionInput)
	assert.Equal(t, "did:lango:buyer", updatedTx.EscrowExecutionInput.BuyerDID)
}
```

- [ ] **Step 2: Run the approval-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestApproveUpfrontPayment_' -count=1
```

Expected:

```text
FAIL
unknown field EscrowExecutionStatus
```

- [ ] **Step 3: Extend the approval tool to bind escrow execution input**

In `internal/app/tools_meta.go`, extend the tool schema and payload:

```go
"escrow_buyer_did":  map[string]interface{}{"type": "string", "description": "Buyer DID for escrow-backed execution"},
"escrow_seller_did": map[string]interface{}{"type": "string", "description": "Seller DID for escrow-backed execution"},
"escrow_reason":     map[string]interface{}{"type": "string", "description": "Reason to store on the escrow entry"},
"escrow_task_id":    map[string]interface{}{"type": "string", "description": "Optional task identifier for the escrow"},
"escrow_milestones": map[string]interface{}{
	"type": "array",
	"items": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"description": map[string]interface{}{"type": "string"},
			"amount":      map[string]interface{}{"type": "string"},
		},
	},
},
```

After calling `paymentapproval.EvaluateUpfrontPayment(...)`, add:

```go
if outcome.SuggestedMode == paymentapproval.ModeEscrow {
	input, err := parseEscrowExecutionInput(params, amount)
	if err != nil {
		return nil, err
	}
	updatedTx, err = receiptStore.BindEscrowExecutionInput(ctx, transactionReceiptID, submissionReceiptID, input)
	if err != nil {
		return nil, fmt.Errorf("bind escrow execution input: %w", err)
	}
}
```

Extend the response payload:

```go
type upfrontPaymentApprovalReceipt struct {
	TransactionReceiptID         string  `json:"transaction_receipt_id"`
	SubmissionReceiptID          string  `json:"submission_receipt_id"`
	Amount                       string  `json:"amount"`
	TrustScore                   float64 `json:"trust_score"`
	UserMaxPrepay                string  `json:"user_max_prepay"`
	RemainingBudget              string  `json:"remaining_budget"`
	Decision                     string  `json:"decision"`
	Reason                       string  `json:"reason"`
	PolicyCode                   string  `json:"policy_code,omitempty"`
	SuggestedMode                string  `json:"suggested_mode"`
	AmountClass                  string  `json:"amount_class,omitempty"`
	RiskClass                    string  `json:"risk_class,omitempty"`
	FailureDetail                string  `json:"failure_detail,omitempty"`
	CurrentPaymentApprovalStatus string  `json:"current_payment_approval_status"`
	CanonicalDecision            string  `json:"canonical_decision,omitempty"`
	CanonicalSettlementHint      string  `json:"canonical_settlement_hint,omitempty"`
	EscrowExecutionStatus        string  `json:"escrow_execution_status,omitempty"`
}
```

- [ ] **Step 4: Re-run the approval-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestApproveUpfrontPayment_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the approval-binding slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_meta_paymentapproval_test.go
git -c commit.gpgsign=false commit -m "feat: bind escrow execution inputs on approval"
```

### Task 3: Add The Escrow Execution Service

**Files:**
- Create: `internal/escrowexecution/types.go`
- Create: `internal/escrowexecution/service.go`
- Test: `internal/escrowexecution/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/escrowexecution/service_test.go` with:

```go
func TestService_ExecuteRecommendation_CreatesAndFundsEscrow(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-execute-escrow",
		ArtifactLabel:       "artifact/execute-escrow",
		PayloadHash:         "hash-execute-escrow",
		SourceLineageDigest: "lineage-execute-escrow",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Escrow path approved.",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.NoError(t, err)
	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "9.00",
		Reason:    "knowledge exchange",
		TaskID:    "task-execute-escrow",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "delivery", Amount: "9.00"},
		},
	})
	require.NoError(t, err)

	runtime := &fakeRuntime{
		createdID: "escrow-123",
	}
	svc := NewService(store, runtime)

	result, err := svc.ExecuteRecommendation(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.NoError(t, err)
	assert.Equal(t, "escrow-123", result.EscrowReference)
	assert.Equal(t, receipts.EscrowExecutionStatusFunded, result.EscrowExecutionStatus)
}
```

- [ ] **Step 2: Run the service tests and verify they fail**

Run:

```bash
go test ./internal/escrowexecution/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/escrowexecution: no Go files
```

- [ ] **Step 3: Implement the execution service**

Create `internal/escrowexecution/types.go`:

```go
package escrowexecution

import "github.com/langoai/lango/internal/receipts"

type Request struct {
	TransactionReceiptID string
}

type Result struct {
	TransactionReceiptID string
	SubmissionReceiptID  string
	EscrowReference      string
	EscrowExecutionStatus receipts.EscrowExecutionStatus
}
```

Create `internal/escrowexecution/service.go`:

```go
package escrowexecution

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/wallet"
)

type runtime interface {
	Create(context.Context, escrow.CreateRequest) (*escrow.EscrowEntry, error)
	Fund(context.Context, string) (*escrow.EscrowEntry, error)
}

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	ApplyEscrowExecutionProgress(context.Context, string, string, receipts.EscrowExecutionStatus, string, receipts.EventType, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	receipts receiptStore
	runtime  runtime
}

func NewService(receipts receiptStore, runtime runtime) *Service {
	return &Service{receipts: receipts, runtime: runtime}
}
```

Continue the `ExecuteRecommendation(...)` body with:

```go
func (s *Service) ExecuteRecommendation(ctx context.Context, req Request) (Result, error) {
	tx, err := s.receipts.GetTransactionReceipt(ctx, req.TransactionReceiptID)
	if err != nil {
		return Result{}, err
	}
	if tx.CurrentSubmissionReceiptID == "" {
		return Result{}, fmt.Errorf("current submission receipt is required")
	}
	if tx.CurrentPaymentApprovalStatus != receipts.PaymentApprovalApproved {
		return Result{}, fmt.Errorf("canonical payment approval must be approved")
	}
	if tx.CanonicalSettlementHint != string(paymentapproval.ModeEscrow) {
		return Result{}, fmt.Errorf("canonical settlement hint must be escrow")
	}
	if tx.EscrowExecutionInput == nil {
		return Result{}, fmt.Errorf("escrow execution input is not bound")
	}

	input := tx.EscrowExecutionInput
	total, err := wallet.ParseUSDC(input.Amount)
	if err != nil {
		return Result{}, fmt.Errorf("parse escrow amount: %w", err)
	}

	milestones := make([]escrow.MilestoneRequest, 0, len(input.Milestones))
	for _, m := range input.Milestones {
		amt, err := wallet.ParseUSDC(m.Amount)
		if err != nil {
			return Result{}, fmt.Errorf("parse milestone amount: %w", err)
		}
		milestones = append(milestones, escrow.MilestoneRequest{
			Description: m.Description,
			Amount:      amt,
		})
	}

	_, err = s.receipts.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusPending, tx.EscrowReference, receipts.EventEscrowExecutionStarted, "")
	if err != nil {
		return Result{}, err
	}

	entry, err := s.runtime.Create(ctx, escrow.CreateRequest{
		BuyerDID:   input.BuyerDID,
		SellerDID:  input.SellerDID,
		Amount:     total,
		Reason:     input.Reason,
		TaskID:     input.TaskID,
		Milestones: milestones,
	})
	if err != nil {
		_, applyErr := s.receipts.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFailed, "", receipts.EventEscrowExecutionFailed, err.Error())
		if applyErr != nil {
			return Result{}, applyErr
		}
		return Result{}, fmt.Errorf("create escrow: %w", err)
	}

	_, err = s.receipts.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusCreated, entry.ID, receipts.EventEscrowExecutionCreated, "")
	if err != nil {
		return Result{}, err
	}

	_, err = s.runtime.Fund(ctx, entry.ID)
	if err != nil {
		_, applyErr := s.receipts.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFailed, entry.ID, receipts.EventEscrowExecutionFailed, err.Error())
		if applyErr != nil {
			return Result{}, applyErr
		}
		return Result{}, fmt.Errorf("fund escrow: %w", err)
	}

	updated, err := s.receipts.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, tx.CurrentSubmissionReceiptID, receipts.EscrowExecutionStatusFunded, entry.ID, receipts.EventEscrowExecutionFunded, "")
	if err != nil {
		return Result{}, err
	}

	return Result{
		TransactionReceiptID: updated.TransactionReceiptID,
		SubmissionReceiptID:  tx.CurrentSubmissionReceiptID,
		EscrowReference:      updated.EscrowReference,
		EscrowExecutionStatus: updated.EscrowExecutionStatus,
	}, nil
}
```

- [ ] **Step 4: Re-run the service tests and verify they pass**

Run:

```bash
go test ./internal/escrowexecution/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the execution service slice**

Run:

```bash
git add internal/escrowexecution internal/receipts
git -c commit.gpgsign=false commit -m "feat: add escrow recommendation execution service"
```

### Task 4: Add The `execute_escrow_recommendation` Meta Tool And Wire It

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/modules.go`
- Create: `internal/app/tools_meta_escrowexecution_test.go`
- Modify: `internal/app/tools_parity_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_escrowexecution_test.go`:

```go
func TestBuildMetaToolsWithEscrow_IncludesExecuteEscrowRecommendation(t *testing.T) {
	tools := buildMetaToolsWithEscrow(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore(), newEscrowEngineForTest())
	tool := findTool(tools, "execute_escrow_recommendation")
	require.NotNil(t, tool)

	required, _ := tool.Parameters["required"].([]string)
	assert.Equal(t, []string{"transaction_receipt_id"}, required)
}

func TestExecuteEscrowRecommendation_ReturnsFundedPayload(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-meta-escrow",
		ArtifactLabel:       "artifact/meta-escrow",
		PayloadHash:         "hash-meta-escrow",
		SourceLineageDigest: "lineage-meta-escrow",
	})
	require.NoError(t, err)
	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Escrow path approved.",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.NoError(t, err)
	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, receipts.EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "7.50",
		Reason:    "knowledge exchange",
		TaskID:    "task-meta-escrow",
		Milestones: []receipts.EscrowMilestoneInput{
			{Description: "delivery", Amount: "7.50"},
		},
	})
	require.NoError(t, err)

	tools := buildMetaToolsWithEscrow(nil, nil, nil, config.SkillConfig{}, nil, store, newEscrowEngineForTest())
	tool := findTool(tools, "execute_escrow_recommendation")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
	})
	require.NoError(t, err)

	payload := got.(executeEscrowRecommendationReceipt)
	assert.Equal(t, tx.TransactionReceiptID, payload.TransactionReceiptID)
	assert.Equal(t, sub.SubmissionReceiptID, payload.SubmissionReceiptID)
	assert.Equal(t, "funded", payload.EscrowExecutionStatus)
	assert.NotEmpty(t, payload.EscrowReference)
}
```

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaToolsWithEscrow_|TestExecuteEscrowRecommendation_' -count=1
```

Expected:

```text
FAIL
undefined: buildMetaToolsWithEscrow
```

- [ ] **Step 3: Implement the tool and registration path**

In `internal/app/tools_meta.go`, keep compatibility and add a new helper:

```go
func buildMetaTools(store *knowledge.Store, engine *learning.Engine, registry *skill.Registry, skillCfg config.SkillConfig, cfg *config.Config, receiptStore *receipts.Store) []*agent.Tool {
	return buildMetaToolsBase(store, engine, registry, skillCfg, cfg, receiptStore, nil)
}
```

Add:

```go
func buildMetaToolsBase(store *knowledge.Store, engine *learning.Engine, registry *skill.Registry, skillCfg config.SkillConfig, cfg *config.Config, receiptStore *receipts.Store, escrowEngine *escrow.Engine) []*agent.Tool {
	tools := []*agent.Tool{
		// existing meta tools
	}
	if escrowEngine != nil {
		tools = append(tools, newExecuteEscrowRecommendationTool(receiptStore, escrowEngine))
	}
	return tools
}

func buildMetaToolsWithEscrow(store *knowledge.Store, engine *learning.Engine, registry *skill.Registry, skillCfg config.SkillConfig, cfg *config.Config, receiptStore *receipts.Store, escrowEngine *escrow.Engine) []*agent.Tool {
	return buildMetaToolsBase(store, engine, registry, skillCfg, cfg, receiptStore, escrowEngine)
}
```

Add the payload and tool:

```go
type executeEscrowRecommendationReceipt struct {
	TransactionReceiptID  string `json:"transaction_receipt_id"`
	SubmissionReceiptID   string `json:"submission_receipt_id"`
	EscrowReference       string `json:"escrow_reference,omitempty"`
	EscrowExecutionStatus string `json:"escrow_execution_status"`
}

func newExecuteEscrowRecommendationTool(receiptStore *receipts.Store, escrowEngine *escrow.Engine) *agent.Tool {
	service := escrowexecution.NewService(receiptStore, escrowEngine)
	return &agent.Tool{
		Name:        "execute_escrow_recommendation",
		Description: "Execute an escrow recommendation by creating and funding the linked escrow path",
		SafetyLevel: agent.SafetyLevelDangerous,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string", "description": "Linked transaction receipt identifier"},
			},
			"required": []string{"transaction_receipt_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			result, err := service.ExecuteRecommendation(ctx, escrowexecution.Request{TransactionReceiptID: transactionReceiptID})
			if err != nil {
				return nil, err
			}
			return executeEscrowRecommendationReceipt{
				TransactionReceiptID:  result.TransactionReceiptID,
				SubmissionReceiptID:   result.SubmissionReceiptID,
				EscrowReference:       result.EscrowReference,
				EscrowExecutionStatus: string(result.EscrowExecutionStatus),
			}, nil
		},
	}
}
```

In `internal/app/modules.go`, register:

```go
metaTools := buildMetaToolsWithEscrow(kc.store, kc.engine, skillReg, cfg.Skill, cfg, fv.ReceiptStore, econc.escrowEngine)
```

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaToolsWithEscrow_|TestExecuteEscrowRecommendation_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/modules.go internal/app/tools_meta_escrowexecution_test.go internal/app/tools_parity_test.go
git -c commit.gpgsign=false commit -m "feat: add escrow recommendation execution tool"
```

### Task 5: Truthful Docs, OpenSpec, And Final Verification

**Files:**
- Create: `docs/security/escrow-execution.md`
- Modify: `docs/security/index.md`
- Modify: `docs/security/upfront-payment-approval.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `README.md`
- Modify: `mkdocs.yml`
- Create: `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/proposal.md`
- Create: `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/design.md`
- Create: `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/tasks.md`
- Create: `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/escrow-execution/spec.md`
- Create: `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/dispute-ready-receipts/spec.md`
- Create: `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/upfront-payment-approval/spec.md`
- Create: `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/security-docs-sync/spec.md`
- Modify: `openspec/specs/dispute-ready-receipts/spec.md`
- Modify: `openspec/specs/upfront-payment-approval/spec.md`
- Modify: `openspec/specs/security-docs-sync/spec.md`
- Create: `openspec/specs/escrow-execution/spec.md`

- [ ] **Step 1: Write the docs and OpenSpec delta files**

Put these exact requirements into `openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/escrow-execution/spec.md`:

```md
## ADDED Requirements

### Requirement: Execute escrow recommendation from canonical transaction state
The system SHALL provide a knowledge-exchange operator path that creates and funds escrow when the linked transaction receipt is approved and canonically recommends `escrow`.

#### Scenario: Approved escrow recommendation executes
- **WHEN** `current_payment_approval_status = approved` and `canonical_settlement_hint = escrow`
- **THEN** the system SHALL create and fund the linked escrow path

#### Scenario: Non-escrow transaction denied
- **WHEN** a transaction receipt does not canonically recommend `escrow`
- **THEN** escrow recommendation execution SHALL fail closed
```

Write `docs/security/escrow-execution.md` with these sections:

```md
# Escrow Execution

Lango's first escrow execution slice turns an approved `escrow` recommendation into a real `create + fund` escrow runtime path for knowledge exchange.

## What Ships
- `execute_escrow_recommendation`
- receipt-backed create + fund execution
- transaction receipt escrow execution status
- submission receipt escrow execution events

## Current Limits
- no activate
- no release or refund
- no dispute adjudication
- no human execution UI
```

- [ ] **Step 2: Run documentation and targeted integration checks**

Run:

```bash
go test ./internal/receipts/... ./internal/escrowexecution/... ./internal/app/... -count=1
python3 -m mkdocs build --strict
```

Expected:

```text
ok
INFO    -  Documentation built
```

- [ ] **Step 3: Run full repository verification**

Run:

```bash
go build ./...
go test ./...
python3 -m mkdocs build --strict
```

Expected:

```text
all commands exit 0
```

- [ ] **Step 4: Sync main specs and archive the OpenSpec change**

Run:

```bash
mkdir -p openspec/specs/escrow-execution
cp openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/escrow-execution/spec.md openspec/specs/escrow-execution/spec.md
python3 - <<'PY'
from pathlib import Path

copies = {
    "openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/dispute-ready-receipts/spec.md":
    "openspec/specs/dispute-ready-receipts/spec.md",
    "openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/upfront-payment-approval/spec.md":
    "openspec/specs/upfront-payment-approval/spec.md",
    "openspec/changes/escrow-recommendation-to-escrow-execution-first-slice/specs/security-docs-sync/spec.md":
    "openspec/specs/security-docs-sync/spec.md",
}

for src, dst in copies.items():
    Path(dst).write_text(Path(src).read_text())
PY
mv openspec/changes/escrow-recommendation-to-escrow-execution-first-slice openspec/changes/archive/2026-04-21-escrow-recommendation-to-escrow-execution-first-slice
```

Expected:

```text
change directory moved under openspec/changes/archive/
```

- [ ] **Step 5: Commit docs and OpenSpec closeout**

Run:

```bash
git add README.md docs/security/escrow-execution.md docs/security/index.md docs/security/upfront-payment-approval.md docs/architecture/p2p-knowledge-exchange-track.md mkdocs.yml openspec/specs/escrow-execution/spec.md openspec/specs/dispute-ready-receipts/spec.md openspec/specs/upfront-payment-approval/spec.md openspec/specs/security-docs-sync/spec.md openspec/changes/archive/2026-04-21-escrow-recommendation-to-escrow-execution-first-slice
git -c commit.gpgsign=false commit -m "specs: archive escrow execution first slice"
```

## Self-Review

- Spec coverage:
  - transaction receipt stores escrow execution input: Task 1 and Task 2
  - recommendation executes through service: Task 3
  - operator-facing meta tool: Task 4
  - canonical evidence and docs/OpenSpec: Task 5
- Placeholder scan:
  - no `TODO`, `TBD`, or deferred implementation markers remain
- Type consistency:
  - `EscrowExecutionInput`, `EscrowExecutionStatus`, `BindEscrowExecutionInput`, `ApplyEscrowExecutionProgress`, `execute_escrow_recommendation`, and `buildMetaToolsWithEscrow` are used consistently across tasks
