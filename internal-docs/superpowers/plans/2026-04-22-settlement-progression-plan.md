# Settlement Progression Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first transaction-level settlement progression layer for `knowledge exchange v1`, mapping release outcomes into canonical settlement states without yet implementing the full money-moving executor or dispute engine.

**Architecture:** Extend `transaction receipt` with settlement progression state, progression reason, partial-settlement hint, and dispute-ready marker. Add a small `internal/settlementprogression` service that translates release approval outcomes into transaction-level progression while keeping actual fund movement delegated to separate executors. Reuse the existing receipt model and runtime foundation instead of introducing a separate settlement record.

**Tech Stack:** Go, `internal/receipts`, `internal/approvalflow`, `internal/knowledgeruntime`, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/receipts/types.go`
  - Add settlement progression state, progression reason, partial-settlement hint, and dispute-ready marker types/fields.
- Modify: `internal/receipts/store.go`
  - Add transaction-level settlement progression helpers and validation.
- Modify: `internal/receipts/store_test.go`
  - Cover state ownership and illegal transition rejection.
- Create: `internal/settlementprogression/types.go`
  - Progression request/result types and disagreement classes.
- Create: `internal/settlementprogression/service.go`
  - Transaction-level settlement progression service.
- Create: `internal/settlementprogression/service_test.go`
  - Focused tests for outcome mapping and dispute-ready opening.
- Modify: `internal/app/tools_meta.go`
  - Add the first settlement progression meta tool(s).
- Modify: `internal/app/tools_parity_test.go`
  - Extend expected meta-tool parity.
- Create: `internal/app/tools_meta_settlementprogression_test.go`
  - Meta-tool coverage for settlement progression.
- Modify: `docs/architecture/knowledge-exchange-runtime.md`
  - Reflect that settlement progression now has an implementation slice.
- Create: `docs/architecture/settlement-progression.md`
  - Public architecture/operator doc for the settlement progression slice.
- Modify: `docs/architecture/index.md`
  - Add the new settlement progression page.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark settlement progression design/implementation slice as landed and move remaining work down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/settlement-progression/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync runtime/architecture page requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track/landing references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync settlement progression meta-tool requirements.

### Task 1: Extend Receipts With Settlement Progression State

**Files:**
- Modify: `internal/receipts/types.go`
- Modify: `internal/receipts/store.go`
- Test: `internal/receipts/store_test.go`

- [ ] **Step 1: Write the failing receipt tests**

Add to `internal/receipts/store_test.go`:

```go
func TestApplySettlementProgression_MapsReleaseOutcomeToCanonicalState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-settle-1",
		Counterparty:   "did:lango:peer-1",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	updated, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, "approve", "")
	require.NoError(t, err)
	assert.Equal(t, SettlementProgressionApprovedForSettlement, updated.SettlementProgressionStatus)
	assert.Equal(t, "approve", updated.SettlementProgressionReason)
}

func TestApplySettlementProgression_RejectsIllegalRewind(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "deal-settle-2",
		Counterparty:   "did:lango:peer-2",
		RequestedScope: "artifact/code-review",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.83",
	})
	require.NoError(t, err)

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, "approve", "")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionPending, "rewind", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
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
undefined: SettlementProgressionApprovedForSettlement
```

- [ ] **Step 3: Add settlement progression types and store helpers**

In `internal/receipts/types.go`, add:

```go
var (
	ErrInvalidSettlementProgressionState = errors.New("invalid settlement progression state")
)

type SettlementProgressionStatus string

const (
	SettlementProgressionPending               SettlementProgressionStatus = "pending"
	SettlementProgressionInProgress            SettlementProgressionStatus = "in-progress"
	SettlementProgressionReviewNeeded          SettlementProgressionStatus = "review-needed"
	SettlementProgressionApprovedForSettlement SettlementProgressionStatus = "approved-for-settlement"
	SettlementProgressionPartiallySettled      SettlementProgressionStatus = "partially-settled"
	SettlementProgressionSettled               SettlementProgressionStatus = "settled"
	SettlementProgressionDisputeReady          SettlementProgressionStatus = "dispute-ready"
)
```

Extend `TransactionReceipt`:

```go
type TransactionReceipt struct {
	TransactionReceiptID           string                         `json:"transaction_receipt_id"`
	TransactionID                  string                         `json:"transaction_id"`
	Counterparty                   string                         `json:"counterparty,omitempty"`
	RequestedScope                 string                         `json:"requested_scope,omitempty"`
	PriceContext                   string                         `json:"price_context,omitempty"`
	TrustContext                   string                         `json:"trust_context,omitempty"`
	KnowledgeExchangeRuntimeStatus KnowledgeExchangeRuntimeStatus `json:"knowledge_exchange_runtime_status,omitempty"`
	SettlementProgressionStatus    SettlementProgressionStatus    `json:"settlement_progression_status,omitempty"`
	SettlementProgressionReason    string                         `json:"settlement_progression_reason,omitempty"`
	PartialSettlementHint          string                         `json:"partial_settlement_hint,omitempty"`
	DisputeReady                   bool                           `json:"dispute_ready,omitempty"`
	CurrentSubmissionReceiptID     string                         `json:"current_submission_receipt_id,omitempty"`
	CanonicalApprovalStatus        ApprovalStatus                 `json:"canonical_approval_status"`
	CanonicalSettlementStatus      SettlementStatus               `json:"canonical_settlement_status"`
	CurrentPaymentApprovalStatus   PaymentApprovalStatus          `json:"current_payment_approval_status"`
	CanonicalDecision              string                         `json:"canonical_decision,omitempty"`
	CanonicalSettlementHint        string                         `json:"canonical_settlement_hint,omitempty"`
	EscrowExecutionStatus          EscrowExecutionStatus          `json:"escrow_execution_status,omitempty"`
	EscrowReference                string                         `json:"escrow_reference,omitempty"`
	EscrowExecutionInput           *EscrowExecutionInput          `json:"escrow_execution_input,omitempty"`
}
```

In `internal/receipts/store.go`, add:

```go
func (s *Store) ApplySettlementProgression(_ context.Context, transactionReceiptID string, next SettlementProgressionStatus, reason string, partialHint string) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	if err := validateSettlementProgressionTransition(tx.SettlementProgressionStatus, next); err != nil {
		return TransactionReceipt{}, err
	}

	tx.SettlementProgressionStatus = next
	tx.SettlementProgressionReason = reason
	tx.PartialSettlementHint = partialHint
	tx.DisputeReady = next == SettlementProgressionDisputeReady
	s.transactions[transactionReceiptID] = tx
	return tx, nil
}

func validateSettlementProgressionTransition(current, next SettlementProgressionStatus) error {
	switch current {
	case "":
		if next == SettlementProgressionPending || next == SettlementProgressionApprovedForSettlement || next == SettlementProgressionReviewNeeded {
			return nil
		}
	case SettlementProgressionPending:
		if next == SettlementProgressionApprovedForSettlement || next == SettlementProgressionReviewNeeded {
			return nil
		}
	case SettlementProgressionApprovedForSettlement:
		if next == SettlementProgressionInProgress || next == SettlementProgressionSettled || next == SettlementProgressionPartiallySettled {
			return nil
		}
	case SettlementProgressionReviewNeeded:
		if next == SettlementProgressionReviewNeeded || next == SettlementProgressionDisputeReady {
			return nil
		}
	case SettlementProgressionInProgress:
		if next == SettlementProgressionSettled || next == SettlementProgressionPartiallySettled || next == SettlementProgressionReviewNeeded {
			return nil
		}
	case SettlementProgressionPartiallySettled:
		if next == SettlementProgressionSettled || next == SettlementProgressionReviewNeeded || next == SettlementProgressionDisputeReady {
			return nil
		}
	}
	return fmt.Errorf("%w: %q -> %q", ErrInvalidSettlementProgressionState, current, next)
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

- [ ] **Step 5: Commit the settlement progression receipt slice**

Run:

```bash
git add internal/receipts/types.go internal/receipts/store.go internal/receipts/store_test.go
git -c commit.gpgsign=false commit -m "feat: add settlement progression receipt state"
```

### Task 2: Add The Settlement Progression Service

**Files:**
- Create: `internal/settlementprogression/types.go`
- Create: `internal/settlementprogression/service.go`
- Test: `internal/settlementprogression/service_test.go`

- [ ] **Step 1: Write the failing service tests**

Create `internal/settlementprogression/service_test.go` with:

```go
func TestService_ApplyReleaseOutcome_ApproveMovesToApprovedForSettlement(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-sp-1",
		Counterparty:   "did:lango:peer-1",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	svc := NewService(store)
	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              OutcomeApprove,
	})
	require.NoError(t, err)
	assert.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Status)
}

func TestService_ApplyReleaseOutcome_RejectMovesToReviewNeeded(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-sp-2",
		Counterparty:   "did:lango:peer-2",
		RequestedScope: "artifact/design-draft",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.80",
	})
	require.NoError(t, err)

	svc := NewService(store)
	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              OutcomeReject,
	})
	require.NoError(t, err)
	assert.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Status)
}
```

- [ ] **Step 2: Run the service tests and verify they fail**

Run:

```bash
go test ./internal/settlementprogression/... -count=1
```

Expected:

```text
FAIL
package github.com/langoai/lango/internal/settlementprogression: no Go files
```

- [ ] **Step 3: Implement the service**

Create `internal/settlementprogression/types.go`:

```go
package settlementprogression

import "github.com/langoai/lango/internal/receipts"

type Outcome string

const (
	OutcomeApprove         Outcome = "approve"
	OutcomeReject          Outcome = "reject"
	OutcomeRequestRevision Outcome = "request-revision"
	OutcomeEscalate        Outcome = "escalate"
)

type ApplyReleaseOutcomeRequest struct {
	TransactionReceiptID string
	Outcome              Outcome
	Reason               string
	PartialHint          string
}

type Result struct {
	TransactionReceiptID string
	Status               receipts.SettlementProgressionStatus
}
```

Create `internal/settlementprogression/service.go`:

```go
package settlementprogression

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	ApplySettlementProgression(context.Context, string, receipts.SettlementProgressionStatus, string, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) ApplyReleaseOutcome(ctx context.Context, req ApplyReleaseOutcomeRequest) (Result, error) {
	var next receipts.SettlementProgressionStatus
	var reason string

	switch req.Outcome {
	case OutcomeApprove:
		next = receipts.SettlementProgressionApprovedForSettlement
		reason = "approve"
	case OutcomeReject:
		next = receipts.SettlementProgressionReviewNeeded
		reason = "reject"
	case OutcomeRequestRevision:
		next = receipts.SettlementProgressionReviewNeeded
		reason = "request-revision"
	case OutcomeEscalate:
		next = receipts.SettlementProgressionReviewNeeded
		if req.Reason != "" {
			reason = req.Reason
		} else {
			reason = "higher approval needed"
		}
	default:
		return Result{}, fmt.Errorf("unsupported release outcome %q", req.Outcome)
	}

	tx, err := s.store.ApplySettlementProgression(ctx, req.TransactionReceiptID, next, reason, req.PartialHint)
	if err != nil {
		return Result{}, err
	}
	return Result{
		TransactionReceiptID: tx.TransactionReceiptID,
		Status:               tx.SettlementProgressionStatus,
	}, nil
}
```

- [ ] **Step 4: Re-run the service tests and verify they pass**

Run:

```bash
go test ./internal/settlementprogression/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the settlement progression service**

Run:

```bash
git add internal/settlementprogression/types.go internal/settlementprogression/service.go internal/settlementprogression/service_test.go
git -c commit.gpgsign=false commit -m "feat: add settlement progression service"
```

### Task 3: Add Settlement Progression Meta Tools

**Files:**
- Modify: `internal/app/tools_meta.go`
- Modify: `internal/app/tools_parity_test.go`
- Create: `internal/app/tools_meta_settlementprogression_test.go`

- [ ] **Step 1: Write the failing meta-tool tests**

Create `internal/app/tools_meta_settlementprogression_test.go`:

```go
func TestBuildMetaTools_IncludesSettlementProgressionTool(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, receipts.NewStore())
	require.NotNil(t, findTool(tools, "apply_settlement_progression"))
}

func TestApplySettlementProgression_ApprovePathReturnsCanonicalState(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-meta-sp-1",
		Counterparty:   "did:lango:peer-1",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	tool := findTool(buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil, store), "apply_settlement_progression")
	require.NotNil(t, tool)

	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_receipt_id": tx.TransactionReceiptID,
		"outcome":                "approve",
	})
	require.NoError(t, err)

	payload := got.(map[string]interface{})
	assert.Equal(t, tx.TransactionReceiptID, payload["transaction_receipt_id"])
	assert.Equal(t, string(receipts.SettlementProgressionApprovedForSettlement), payload["settlement_progression_status"])
}
```

- [ ] **Step 2: Run the meta-tool tests and verify they fail**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesSettlementProgressionTool|TestApplySettlementProgression_' -count=1
```

Expected:

```text
FAIL
tool not found
```

- [ ] **Step 3: Implement the settlement progression meta tool**

In `internal/app/tools_meta.go`, add:

```go
func newApplySettlementProgressionTool(receiptStore *receipts.Store) *agent.Tool {
	return &agent.Tool{
		Name:        "apply_settlement_progression",
		Description: "Apply artifact release outcomes to the transaction-level settlement progression state",
		SafetyLevel: agent.SafetyLevelModerate,
		Capability: agent.ToolCapability{
			Category: "knowledge",
			Activity: agent.ActivityWrite,
		},
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transaction_receipt_id": map[string]interface{}{"type": "string"},
				"outcome":                map[string]interface{}{"type": "string", "enum": []string{"approve", "reject", "request-revision", "escalate"}},
				"reason":                 map[string]interface{}{"type": "string"},
				"partial_hint":           map[string]interface{}{"type": "string"},
			},
			"required": []string{"transaction_receipt_id", "outcome"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			if receiptStore == nil {
				return nil, fmt.Errorf("receipts store dependency is not configured")
			}
			transactionReceiptID, err := toolparam.RequireString(params, "transaction_receipt_id")
			if err != nil {
				return nil, err
			}
			outcomeStr, err := toolparam.RequireString(params, "outcome")
			if err != nil {
				return nil, err
			}
			svc := settlementprogression.NewService(receiptStore)
			result, err := svc.ApplyReleaseOutcome(ctx, settlementprogression.ApplyReleaseOutcomeRequest{
				TransactionReceiptID: transactionReceiptID,
				Outcome:              settlementprogression.Outcome(outcomeStr),
				Reason:               toolparam.OptionalString(params, "reason", ""),
				PartialHint:          toolparam.OptionalString(params, "partial_hint", ""),
			})
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"transaction_receipt_id":      result.TransactionReceiptID,
				"settlement_progression_status": string(result.Status),
			}, nil
		},
	}
}
```

Register it in `buildMetaTools(...)` when `receiptStore != nil`, and update parity expectations with:

```go
"apply_settlement_progression",
```

- [ ] **Step 4: Re-run the meta-tool tests and verify they pass**

Run:

```bash
go test ./internal/app -run 'TestBuildMetaTools_IncludesSettlementProgressionTool|TestApplySettlementProgression_' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the meta-tool slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_parity_test.go internal/app/tools_meta_settlementprogression_test.go
git -c commit.gpgsign=false commit -m "feat: add settlement progression tools"
```

### Task 4: Document And Close Out The Slice

**Files:**
- Create: `docs/architecture/settlement-progression.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Create: `openspec/changes/settlement-progression/**`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`

- [ ] **Step 1: Write the public settlement progression doc**

Create `docs/architecture/settlement-progression.md`:

```md
# Settlement Progression

This page describes the first transaction-level settlement progression slice for `knowledge exchange v1`.

## What Ships

- transaction-level settlement progression state
- release outcome mapping
- review-needed progression
- approved-for-settlement progression
- dispute-ready opening rules
- a receipts-backed settlement progression meta tool

## Current Limits

- no settlement executor implementation here
- no partial settlement calculation formula
- no dispute engine
- no human adjudication UI
```

- [ ] **Step 2: Wire the new page into architecture docs and nav**

Add to `docs/architecture/index.md`:

```md
-   :material-bank-check-outline: **[Settlement Progression](settlement-progression.md)**
    Transaction-level settlement progression model for approve, revise, reject, escalate, and dispute-ready handoff.
```

In `docs/architecture/p2p-knowledge-exchange-track.md`, replace the payment/settlement follow-on item with:

```md
4. `settlement progression` first slice is now landed; the follow-on work is actual settlement execution, partial settlement rules, and dispute engine completion
```

In `zensical.toml`, add:

```toml
{ "Settlement Progression" = "architecture/settlement-progression.md" }
```

under the `Architecture` nav.

- [ ] **Step 3: Run verification**

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

Create `openspec/changes/settlement-progression/proposal.md`:

```md
## Why

The runtime design now needs a transaction-level settlement progression layer so release outcomes can map into canonical progression state before a full settlement executor or dispute engine exists.

## What Changes

- add a settlement progression design/implementation slice
- expose a transaction-level settlement progression page
- expose a receipts-backed settlement progression meta tool

## Impact

- `docs/architecture/settlement-progression.md`
- architecture landing and track docs
- docs navigation
- `meta-tools` spec
```

Create `openspec/changes/settlement-progression/specs/project-docs/spec.md`:

```md
## ADDED Requirements

### Requirement: Settlement progression page is published
The architecture docs SHALL include a dedicated `settlement-progression.md` page for the first transaction-level settlement progression slice.

#### Scenario: Settlement progression page exists
- **WHEN** a reader opens the architecture docs
- **THEN** they SHALL find the Settlement Progression page
```

Create `openspec/changes/settlement-progression/specs/docs-only/spec.md`:

```md
## ADDED Requirements

### Requirement: Architecture landing and track docs reference settlement progression
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed settlement progression slice.

#### Scenario: Landing page links settlement progression
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Settlement Progression page listed with the other architecture pages

#### Scenario: Track doc reflects landed settlement progression
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** settlement progression SHALL be described as landed slice work with follow-on execution/dispute work still remaining
```

Create `openspec/changes/settlement-progression/specs/meta-tools/spec.md`:

```md
## ADDED Requirements

### Requirement: Settlement progression meta tool
The system SHALL expose a receipts-backed meta tool for applying artifact release outcomes to transaction-level settlement progression state.

#### Scenario: Settlement progression tool available
- **WHEN** the meta tools are built with a receipts store
- **THEN** `apply_settlement_progression` SHALL be available
```

Then sync and archive:

```bash
cp openspec/changes/settlement-progression/specs/project-docs/spec.md openspec/specs/project-docs/spec.md
cp openspec/changes/settlement-progression/specs/docs-only/spec.md openspec/specs/docs-only/spec.md
cp openspec/changes/settlement-progression/specs/meta-tools/spec.md openspec/specs/meta-tools/spec.md
mkdir -p openspec/changes/archive
mv openspec/changes/settlement-progression openspec/changes/archive/2026-04-22-settlement-progression
git add docs/architecture/settlement-progression.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-22-settlement-progression
git -c commit.gpgsign=false commit -m "specs: archive settlement progression"
```

## Self-Review

- Spec coverage:
  - settlement progression receipt state: Task 1
  - settlement progression service: Task 2
  - settlement progression meta tool: Task 3
  - public docs and OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or deferred implementation markers remain
- Type/path consistency:
  - settlement progression remains transaction-owned
  - submission receipts remain evidence/hint providers
  - the public page path is consistently `docs/architecture/settlement-progression.md`
