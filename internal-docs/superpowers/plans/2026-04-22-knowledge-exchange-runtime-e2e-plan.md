# Knowledge Exchange Runtime E2E Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first transaction-oriented runtime control plane for `knowledge exchange v1`, connecting transaction open, payment-path selection, work start gating, submission creation, release approval, and post-approval progression into one coherent runtime surface.

**Architecture:** Add a lightweight `internal/knowledgeruntime` orchestration layer centered on `transaction receipt` as canonical control-plane state and `submission receipt` as canonical deliverable state. Reuse the landed first slices (`evaluate_exportability`, `approve_upfront_payment`, direct prepay execution gating, `create_dispute_ready_receipt`, `approve_artifact_release`, `execute_escrow_recommendation`) instead of duplicating their logic, and add only the minimal transaction-open/runtime-branch orchestration needed to tie them together.

**Tech Stack:** Go, `internal/receipts`, `internal/paymentapproval`, `internal/paymentgate`, `internal/escrowexecution`, `internal/app/tools_meta.go`, Zensical docs, OpenSpec

---

## File Map

- Create: `internal/knowledgeruntime/types.go`
  - Runtime request/result/state types for transaction-oriented orchestration.
- Create: `internal/knowledgeruntime/service.go`
  - Orchestration service for transaction open, path selection, work-start gating, submission progression, and release aftermath.
- Create: `internal/knowledgeruntime/service_test.go`
  - Focused runtime tests for transaction progression and branch selection.
- Modify: `internal/receipts/types.go`
  - Add the minimum transaction-open and runtime-progression fields required by the orchestration layer.
- Modify: `internal/receipts/store.go`
  - Add transaction-open binding and transaction-level runtime progression helpers.
- Modify: `internal/receipts/store_test.go`
  - Cover new runtime state ownership and progression constraints.
- Modify: `internal/app/tools_meta.go`
  - Add the new runtime meta tools and wire them to the orchestration service.
- Modify: `internal/app/tools_parity_test.go`
  - Extend expected meta-tool parity.
- Create: `internal/app/tools_meta_knowledgeruntime_test.go`
  - Meta-tool coverage for transaction open and runtime progression.
- Modify: `internal/app/modules.go`
  - Register the runtime meta tools through the existing meta-tool build path.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Reflect that the runtime e2e first slice is landed once completed.
- Create: `docs/architecture/knowledge-exchange-runtime.md`
  - Public operator/architecture doc for the first runtime slice.
- Modify: `docs/architecture/index.md`
  - Add the new runtime page if it belongs in public architecture docs.
- Modify: `zensical.toml`
  - Add the runtime page to public navigation if exposed.
- Create: `openspec/changes/knowledge-exchange-runtime-e2e/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync the new runtime architecture page requirement.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track/landing/runtime page references.
- Create or modify: `openspec/specs/meta-tools/spec.md`
  - Sync new runtime meta-tool requirements.

### Task 1: Extend Receipts With Transaction-Open And Runtime State

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Test: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add to `internal/receipts/store_test.go`:

```go
func TestOpenKnowledgeExchangeTransaction_BindsCanonicalInputs(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-1",
		Counterparty:   "did:lango:peer-1",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)
	assert.Equal(t, "did:lango:peer-1", tx.Counterparty)
	assert.Equal(t, "artifact/research-note", tx.RequestedScope)
	assert.Equal(t, RuntimeStatusOpened, tx.RuntimeStatus)
}

func TestApplyKnowledgeExchangeRuntimeProgression_RejectsIllegalBranchRewinds(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-open-2",
		Counterparty:   "did:lango:peer-2",
		RequestedScope: "artifact/code-review",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.83",
	})
	require.NoError(t, err)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusPaymentApproved, "")
	require.NoError(t, err)
	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, tx.TransactionReceiptID, RuntimeStatusOpened, "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidKnowledgeExchangeRuntimeState)
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
undefined: OpenTransactionInput
```

- [ ] **Step 3: Add transaction-open and runtime state types**

In `internal/receipts/types.go`, add:

```go
var (
	ErrInvalidKnowledgeExchangeRuntimeState = errors.New("invalid knowledge exchange runtime state")
)

type KnowledgeExchangeRuntimeStatus string

const (
	RuntimeStatusOpened              KnowledgeExchangeRuntimeStatus = "opened"
	RuntimeStatusExportabilityAdvisory KnowledgeExchangeRuntimeStatus = "exportability-advisory"
	RuntimeStatusPaymentApproved     KnowledgeExchangeRuntimeStatus = "payment-approved"
	RuntimeStatusPaymentAuthorized   KnowledgeExchangeRuntimeStatus = "payment-authorized"
	RuntimeStatusEscrowFunded        KnowledgeExchangeRuntimeStatus = "escrow-funded"
	RuntimeStatusWorkStarted         KnowledgeExchangeRuntimeStatus = "work-started"
	RuntimeStatusSubmissionReceived  KnowledgeExchangeRuntimeStatus = "submission-received"
	RuntimeStatusReleaseApproved     KnowledgeExchangeRuntimeStatus = "release-approved"
	RuntimeStatusRevisionRequested   KnowledgeExchangeRuntimeStatus = "revision-requested"
	RuntimeStatusEscalated           KnowledgeExchangeRuntimeStatus = "escalated"
	RuntimeStatusDisputeReady        KnowledgeExchangeRuntimeStatus = "dispute-ready"
)

type OpenTransactionInput struct {
	TransactionID  string `json:"transaction_id"`
	Counterparty   string `json:"counterparty"`
	RequestedScope string `json:"requested_scope"`
	PriceContext   string `json:"price_context"`
	TrustContext   string `json:"trust_context"`
}
```

Extend `TransactionReceipt`:

```go
type TransactionReceipt struct {
	TransactionReceiptID           string                        `json:"transaction_receipt_id"`
	TransactionID                  string                        `json:"transaction_id"`
	Counterparty                   string                        `json:"counterparty,omitempty"`
	RequestedScope                 string                        `json:"requested_scope,omitempty"`
	PriceContext                   string                        `json:"price_context,omitempty"`
	TrustContext                   string                        `json:"trust_context,omitempty"`
	KnowledgeExchangeRuntimeStatus KnowledgeExchangeRuntimeStatus `json:"knowledge_exchange_runtime_status,omitempty"`
	CurrentSubmissionReceiptID     string                        `json:"current_submission_receipt_id,omitempty"`
	CanonicalApprovalStatus        ApprovalStatus                `json:"canonical_approval_status"`
	CanonicalSettlementStatus      SettlementStatus              `json:"canonical_settlement_status"`
	CurrentPaymentApprovalStatus   PaymentApprovalStatus         `json:"current_payment_approval_status"`
	CanonicalDecision              string                        `json:"canonical_decision,omitempty"`
	CanonicalSettlementHint        string                        `json:"canonical_settlement_hint,omitempty"`
	EscrowExecutionStatus          EscrowExecutionStatus         `json:"escrow_execution_status,omitempty"`
	EscrowReference                string                        `json:"escrow_reference,omitempty"`
	EscrowExecutionInput           *EscrowExecutionInput         `json:"escrow_execution_input,omitempty"`
}
```

In `internal/receipts/store.go`, add:

```go
func (s *Store) OpenKnowledgeExchangeTransaction(_ context.Context, in OpenTransactionInput) (TransactionReceipt, error) {
	if strings.TrimSpace(in.TransactionID) == "" || strings.TrimSpace(in.Counterparty) == "" || strings.TrimSpace(in.RequestedScope) == "" {
		return TransactionReceipt{}, fmt.Errorf("%w: transaction_id, counterparty, and requested_scope are required", ErrInvalidSubmissionInput)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	txReceiptID, ok := s.txByExternalID[in.TransactionID]
	if !ok {
		txReceiptID = uuid.NewString()
		s.txByExternalID[in.TransactionID] = txReceiptID
	}

	tx := TransactionReceipt{
		TransactionReceiptID:           txReceiptID,
		TransactionID:                  in.TransactionID,
		Counterparty:                   in.Counterparty,
		RequestedScope:                 in.RequestedScope,
		PriceContext:                   in.PriceContext,
		TrustContext:                   in.TrustContext,
		KnowledgeExchangeRuntimeStatus: RuntimeStatusOpened,
		CanonicalApprovalStatus:        ApprovalPending,
		CanonicalSettlementStatus:      SettlementPending,
		CurrentPaymentApprovalStatus:   PaymentApprovalPending,
	}

	if existing, exists := s.transactions[txReceiptID]; exists {
		tx.CurrentSubmissionReceiptID = existing.CurrentSubmissionReceiptID
		tx.CanonicalApprovalStatus = existing.CanonicalApprovalStatus
		tx.CanonicalSettlementStatus = existing.CanonicalSettlementStatus
		tx.CurrentPaymentApprovalStatus = existing.CurrentPaymentApprovalStatus
		tx.CanonicalDecision = existing.CanonicalDecision
		tx.CanonicalSettlementHint = existing.CanonicalSettlementHint
		tx.EscrowExecutionStatus = existing.EscrowExecutionStatus
		tx.EscrowReference = existing.EscrowReference
		tx.EscrowExecutionInput = existing.EscrowExecutionInput
	}

	s.transactions[txReceiptID] = tx
	return tx, nil
}

func (s *Store) ApplyKnowledgeExchangeRuntimeProgression(_ context.Context, transactionReceiptID string, next KnowledgeExchangeRuntimeStatus, submissionReceiptID string) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	if err := validateKnowledgeExchangeRuntimeTransition(tx.KnowledgeExchangeRuntimeStatus, next); err != nil {
		return TransactionReceipt{}, err
	}
	if submissionReceiptID != "" {
		tx.CurrentSubmissionReceiptID = submissionReceiptID
	}
	tx.KnowledgeExchangeRuntimeStatus = next
	s.transactions[transactionReceiptID] = tx
	return tx, nil
}
```

Add a transition validator:

```go
func validateKnowledgeExchangeRuntimeTransition(current, next KnowledgeExchangeRuntimeStatus) error {
	switch current {
	case "":
		if next == RuntimeStatusOpened {
			return nil
		}
	case RuntimeStatusOpened:
		if next == RuntimeStatusExportabilityAdvisory || next == RuntimeStatusPaymentApproved {
			return nil
		}
	case RuntimeStatusExportabilityAdvisory:
		if next == RuntimeStatusPaymentApproved {
			return nil
		}
	case RuntimeStatusPaymentApproved:
		if next == RuntimeStatusPaymentAuthorized || next == RuntimeStatusEscrowFunded {
			return nil
		}
	case RuntimeStatusPaymentAuthorized, RuntimeStatusEscrowFunded:
		if next == RuntimeStatusWorkStarted {
			return nil
		}
	case RuntimeStatusWorkStarted:
		if next == RuntimeStatusSubmissionReceived {
			return nil
		}
	case RuntimeStatusSubmissionReceived:
		if next == RuntimeStatusReleaseApproved || next == RuntimeStatusRevisionRequested || next == RuntimeStatusEscalated || next == RuntimeStatusDisputeReady {
			return nil
		}
	case RuntimeStatusRevisionRequested:
		if next == RuntimeStatusSubmissionReceived {
			return nil
		}
	}
	return fmt.Errorf("%w: %q -> %q", ErrInvalidKnowledgeExchangeRuntimeState, current, next)
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

- [ ] **Step 5: Commit the receipt/runtime-state slice**

Run:

```bash
git add internal/receipts/types.go internal/receipts/store.go internal/receipts/store_test.go
git -c commit.gpgsign=false commit -m "feat: add knowledge exchange runtime receipt state"
```

### Task 2: Add The Runtime Orchestration Service

**Files:**
- Create: `internal/knowledgeruntime/types.go`
- Create: `internal/knowledgeruntime/service.go`
- Test: `internal/knowledgeruntime/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/knowledgeruntime/service_test.go` with:

```go
func TestService_OpenTransaction_RecordsCanonicalOpenState(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	svc := NewService(store)

	tx, err := svc.OpenTransaction(ctx, OpenTransactionRequest{
		TransactionID:  "deal-rt-1",
		Counterparty:   "did:lango:peer-1",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.71",
	})
	require.NoError(t, err)
	assert.Equal(t, receipts.RuntimeStatusOpened, tx.RuntimeStatus)
}

func TestService_SelectExecutionPath_UsesPrepayBranch(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	svc := NewService(store)

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-rt-2",
		Counterparty:   "did:lango:peer-2",
		RequestedScope: "artifact/design-draft",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.90",
	})
	require.NoError(t, err)

	updated, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, "", paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	branch, err := svc.SelectExecutionPath(ctx, updated.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, BranchPrepay, branch.Branch)
}
```

- [ ] **Step 2: Run the service tests and verify they fail**

Run:

```bash
go test ./internal/knowledgeruntime/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/knowledgeruntime: no Go files
```

- [ ] **Step 3: Implement the service**

Create `internal/knowledgeruntime/types.go`:

```go
package knowledgeruntime

import "github.com/langoai/lango/internal/receipts"

type OpenTransactionRequest struct {
	TransactionID  string
	Counterparty   string
	RequestedScope string
	PriceContext   string
	TrustContext   string
}

type OpenTransactionResult struct {
	TransactionReceiptID string
	RuntimeStatus        receipts.KnowledgeExchangeRuntimeStatus
}

type Branch string

const (
	BranchPrepay Branch = "prepay"
	BranchEscrow Branch = "escrow"
)

type BranchSelection struct {
	TransactionReceiptID string
	Branch               Branch
}
```

Create `internal/knowledgeruntime/service.go`:

```go
package knowledgeruntime

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	OpenKnowledgeExchangeTransaction(context.Context, receipts.OpenTransactionInput) (receipts.TransactionReceipt, error)
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	ApplyKnowledgeExchangeRuntimeProgression(context.Context, string, receipts.KnowledgeExchangeRuntimeStatus, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) OpenTransaction(ctx context.Context, req OpenTransactionRequest) (OpenTransactionResult, error) {
	tx, err := s.store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  req.TransactionID,
		Counterparty:   req.Counterparty,
		RequestedScope: req.RequestedScope,
		PriceContext:   req.PriceContext,
		TrustContext:   req.TrustContext,
	})
	if err != nil {
		return OpenTransactionResult{}, err
	}
	return OpenTransactionResult{
		TransactionReceiptID: tx.TransactionReceiptID,
		RuntimeStatus:        tx.KnowledgeExchangeRuntimeStatus,
	}, nil
}

func (s *Service) SelectExecutionPath(ctx context.Context, transactionReceiptID string) (BranchSelection, error) {
	tx, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		return BranchSelection{}, err
	}
	switch tx.CanonicalSettlementHint {
	case string(paymentapproval.ModePrepay):
		_, err = s.store.ApplyKnowledgeExchangeRuntimeProgression(ctx, transactionReceiptID, receipts.RuntimeStatusPaymentApproved, "")
		if err != nil {
			return BranchSelection{}, err
		}
		return BranchSelection{TransactionReceiptID: transactionReceiptID, Branch: BranchPrepay}, nil
	case string(paymentapproval.ModeEscrow):
		_, err = s.store.ApplyKnowledgeExchangeRuntimeProgression(ctx, transactionReceiptID, receipts.RuntimeStatusPaymentApproved, "")
		if err != nil {
			return BranchSelection{}, err
		}
		return BranchSelection{TransactionReceiptID: transactionReceiptID, Branch: BranchEscrow}, nil
	default:
		return BranchSelection{}, fmt.Errorf("transaction %q has unsupported settlement hint %q", transactionReceiptID, tx.CanonicalSettlementHint)
	}
}
```

- [ ] **Step 4: Re-run the service tests and verify they pass**

Run:

```bash
go test ./internal/knowledgeruntime/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the orchestration service slice**

Run:

```bash
git add internal/knowledgeruntime/types.go internal/knowledgeruntime/service.go internal/knowledgeruntime/service_test.go
git -c commit.gpgsign=false commit -m "feat: add knowledge exchange runtime orchestration"
```

### Task 3: Add Runtime Meta Tools

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_knowledgeruntime_test.go`
- Modify: `internal/app/modules.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_knowledgeruntime_test.go`:

```go
func TestBuildMetaTools_IncludesKnowledgeExchangeRuntimeTools(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	require.NotNil(t, findTool(tools, "open_knowledge_exchange_transaction"))
	require.NotNil(t, findTool(tools, "select_knowledge_exchange_path"))
}

func TestOpenKnowledgeExchangeTransaction_ReturnsCanonicalPayload(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store)
	tool := findTool(tools, "open_knowledge_exchange_transaction")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_id":  "deal-meta-1",
		"counterparty":    "did:lango:peer-1",
		"requested_scope": "artifact/research-note",
		"price_context":   "quote:0.50-usdc",
		"trust_context":   "trust:0.71",
	})
	require.NoError(t, err)

	payload := got.(map[string]interface{})
	assert.Equal(t, "deal-meta-1", payload["transaction_id"])
	assert.NotEmpty(t, payload["transaction_receipt_id"])
	assert.Equal(t, string(receipts.RuntimeStatusOpened), payload["knowledge_exchange_runtime_status"])
}
```

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesKnowledgeExchangeRuntimeTools|TestOpenKnowledgeExchangeTransaction_' -count=1
```

Expected:

```text
FAIL
tool not found
```

- [ ] **Step 3: Add the new runtime tools**

In `internal/app/tools_meta.go`, add two new tools:

```go
{
	Name:        "open_knowledge_exchange_transaction",
	Description: "Open a knowledge-exchange transaction and record the canonical transaction inputs",
	SafetyLevel: agent.SafetyLevelModerate,
	Capability: agent.ToolCapability{
		Category: "knowledge",
		Activity: agent.ActivityWrite,
	},
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"transaction_id":  map[string]interface{}{"type": "string"},
			"counterparty":    map[string]interface{}{"type": "string"},
			"requested_scope": map[string]interface{}{"type": "string"},
			"price_context":   map[string]interface{}{"type": "string"},
			"trust_context":   map[string]interface{}{"type": "string"},
		},
		"required": []string{"transaction_id", "counterparty", "requested_scope", "price_context", "trust_context"},
	},
	Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if receiptStore == nil {
			return nil, fmt.Errorf("receipts store dependency is not configured")
		}
		service := knowledgeruntime.NewService(receiptStore)
		txID, err := toolparam.RequireString(params, "transaction_id")
		if err != nil {
			return nil, err
		}
		counterparty, err := toolparam.RequireString(params, "counterparty")
		if err != nil {
			return nil, err
		}
		scope, err := toolparam.RequireString(params, "requested_scope")
		if err != nil {
			return nil, err
		}
		priceContext, err := toolparam.RequireString(params, "price_context")
		if err != nil {
			return nil, err
		}
		trustContext, err := toolparam.RequireString(params, "trust_context")
		if err != nil {
			return nil, err
		}
		result, err := service.OpenTransaction(ctx, knowledgeruntime.OpenTransactionRequest{
			TransactionID:  txID,
			Counterparty:   counterparty,
			RequestedScope: scope,
			PriceContext:   priceContext,
			TrustContext:   trustContext,
		})
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"transaction_id":                   txID,
			"transaction_receipt_id":           result.TransactionReceiptID,
			"knowledge_exchange_runtime_status": string(result.RuntimeStatus),
		}, nil
	},
},
{
	Name:        "select_knowledge_exchange_path",
	Description: "Select the current knowledge-exchange runtime branch from canonical transaction state",
	SafetyLevel: agent.SafetyLevelModerate,
	Capability: agent.ToolCapability{
		Category: "knowledge",
		Activity: agent.ActivityWrite,
	},
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"transaction_receipt_id": map[string]interface{}{"type": "string"},
		},
		"required": []string{"transaction_receipt_id"},
	},
	Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if receiptStore == nil {
			return nil, fmt.Errorf("receipts store dependency is not configured")
		}
		service := knowledgeruntime.NewService(receiptStore)
		transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
		if err != nil {
			return nil, err
		}
		branch, err := service.SelectExecutionPath(ctx, transactionReceiptID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"transaction_receipt_id": transactionReceiptID,
			"branch":                 string(branch.Branch),
		}, nil
	},
},
```

Update `internal/app/tools_parity_test.go` to include:

```go
"open_knowledge_exchange_transaction",
"select_knowledge_exchange_path",
```

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesKnowledgeExchangeRuntimeTools|TestOpenKnowledgeExchangeTransaction_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the runtime meta tools**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_knowledgeruntime_test.go internal/app/modules.go
git -c commit.gpgsign=false commit -m "feat: add knowledge exchange runtime tools"
```

### Task 4: Document And Close Out The Runtime Slice

**Files:**
- Create: `docs/architecture/knowledge-exchange-runtime.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Create: `openspec/changes/knowledge-exchange-runtime-e2e/**`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`

- [ ] **Step 1: Write the public runtime doc**

Create `docs/architecture/knowledge-exchange-runtime.md`:

```md
# Knowledge Exchange Runtime

This page describes the first transaction-oriented runtime model for `knowledge exchange v1`.

## What Ships

- transaction open
- exportability advisory placement
- upfront payment approval placement
- prepay vs escrow branch selection
- work start gates
- artifact submission as canonical submission receipt creation
- release approval branch outcomes
- settlement progression vs revision vs dispute handoff

## Current Limits

- no full dispute engine
- no human approval UI
- no full escrow lifecycle beyond `create + fund`
```

- [ ] **Step 2: Wire the new runtime doc into architecture docs and nav**

Add to `docs/architecture/index.md`:

```md
-   :material-timeline-check-outline: **[Knowledge Exchange Runtime](knowledge-exchange-runtime.md)**
    Transaction-oriented runtime view that connects exportability, approval, payment, submission, settlement, and dispute handoff.
```

In `docs/architecture/p2p-knowledge-exchange-track.md`, replace the pending runtime item with:

```md
5. `knowledge exchange runtime` first end-to-end design is now landed; the follow-on work is implementation, settlement progression completion, and dispute/runtime completion
```

In `zensical.toml`, add:

```toml
{ "Knowledge Exchange Runtime" = "architecture/knowledge-exchange-runtime.md" }
```

under the `Architecture` nav.

- [ ] **Step 3: Run full verification**

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

- [ ] **Step 4: Write and archive the OpenSpec change**

Create `openspec/changes/knowledge-exchange-runtime-e2e/proposal.md`:

```md
## Why

The knowledge-exchange track now has several landed first slices, but no single transaction-oriented runtime document explaining how they compose into one coherent exchange path.

## What Changes

- add a knowledge-exchange runtime design page
- expose it in architecture docs and navigation
- record the landed runtime-design slice in OpenSpec

## Impact

- `docs/architecture/knowledge-exchange-runtime.md`
- architecture landing and track docs
- docs navigation
```

Create `openspec/changes/knowledge-exchange-runtime-e2e/specs/project-docs/spec.md`:

```md
## ADDED Requirements

### Requirement: Knowledge exchange runtime page is published
The architecture docs SHALL include a dedicated `knowledge-exchange-runtime.md` page that explains the transaction-oriented runtime model for `knowledge exchange v1`.

#### Scenario: Runtime page exists
- **WHEN** a reader opens the architecture docs
- **THEN** they SHALL find the Knowledge Exchange Runtime page
```

Create `openspec/changes/knowledge-exchange-runtime-e2e/specs/docs-only/spec.md`:

```md
## ADDED Requirements

### Requirement: Architecture landing and track docs reference the runtime page
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed runtime-design page.

#### Scenario: Landing page links runtime page
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the runtime page listed with the other architecture pages

#### Scenario: Track doc reflects landed runtime design
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the runtime-design work SHALL be described as landed design work with follow-on implementation still remaining
```

Create `openspec/changes/knowledge-exchange-runtime-e2e/specs/meta-tools/spec.md`:

```md
## ADDED Requirements

### Requirement: Knowledge exchange runtime meta tools
The system SHALL expose meta tools for opening a knowledge-exchange transaction and selecting the current runtime branch from canonical transaction state.

#### Scenario: Open transaction tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `open_knowledge_exchange_transaction` SHALL be available

#### Scenario: Branch selection tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `select_knowledge_exchange_path` SHALL be available
```

Then sync:

```bash
cp openspec/changes/knowledge-exchange-runtime-e2e/specs/project-docs/spec.md openspec/specs/project-docs/spec.md
cp openspec/changes/knowledge-exchange-runtime-e2e/specs/docs-only/spec.md openspec/specs/docs-only/spec.md
cp openspec/changes/knowledge-exchange-runtime-e2e/specs/meta-tools/spec.md openspec/specs/meta-tools/spec.md
mkdir -p openspec/changes/archive
mv openspec/changes/knowledge-exchange-runtime-e2e openspec/changes/archive/2026-04-22-knowledge-exchange-runtime-e2e
git add docs/architecture/knowledge-exchange-runtime.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-22-knowledge-exchange-runtime-e2e
git -c commit.gpgsign=false commit -m "specs: archive knowledge exchange runtime design"
```

## Self-Review

- Spec coverage:
  - receipt/runtime-state foundation: Task 1
  - orchestration layer: Task 2
  - runtime meta tools: Task 3
  - public docs and OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or deferred implementation markers remain
- Type/path consistency:
  - transaction-oriented runtime uses `transaction receipt` as canonical control-plane state
  - deliverable progression uses `submission receipt`
  - the public runtime page path is consistently `docs/architecture/knowledge-exchange-runtime.md`
