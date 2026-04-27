# Dispute-Ready Receipts First Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a first `dispute-ready receipt lite` slice with dedicated submission and transaction receipts, canonical state, append-only event trail, and links to exportability, approval, settlement, and provenance context.

**Architecture:** This slice intentionally does not build a dispute engine. It adds a lightweight receipt domain and storage model that sits above existing exportability and approval flow outputs, stores canonical current state separately from event history, and references audit/provenance/settlement surfaces rather than embedding them completely.

**Tech Stack:** Go, Ent, Cobra CLI, existing exportability and approval flow packages (`internal/exportability/*`, `internal/approvalflow/*`), audit/provenance surfaces, MkDocs/Markdown docs

---

## Scope Split

This first slice covers only:

- dedicated `submission receipt` and `transaction receipt` models
- canonical approval / settlement status
- `current submission` pointer
- append-only event trail
- lite provenance summary
- references to audit / provenance / settlement context

This slice does **not** implement:

- dispute adjudication
- human dispute UI
- full settlement execution
- full provenance embedding
- full evidence graph

## OpenSpec Precondition

Before touching implementation code, create or refresh an OpenSpec change for this slice. Use a narrow change name such as `dispute-ready-receipts-first-slice`.

The implementation session must end with the repository's required OpenSpec workflow:

- `ff`
- `apply`
- `verify`
- `sync`
- `archive`

## File Map

- Create: `internal/receipts/types.go`
  - Receipt IDs, canonical statuses, event types, and lite summary structs.
- Create: `internal/receipts/store.go`
  - Receipt persistence and mutation helpers.
- Create: `internal/receipts/store_test.go`
  - Tests for submission creation, transaction linkage, current-submission pointer updates, and event trail append.
- Modify: `internal/ent/schema/*.go`
  - Add dedicated Ent schemas for transaction receipts, submission receipts, and receipt events.
- Modify: generated Ent files under `internal/ent/...`
  - Regenerated after schema changes.
- Modify: `internal/app/tools_meta.go`
  - Add a narrow receipt-generation/update tool or hook surface for the first slice.
- Create: `internal/app/tools_meta_receipts_test.go`
  - Real-store tests for receipt creation/update paths.
- Modify: `docs/security/approval-flow.md`
  - Reference dispute-ready receipt lite and its current limits.
- Create: `docs/security/dispute-ready-receipts.md`
  - Canonical operator doc for the first receipt slice.
- Modify: `docs/security/index.md`
  - Link the new receipt doc.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark dispute-ready receipt lite as landed once implemented.
- Modify: `docs/architecture/trust-security-policy-audit.md`
  - Add post-implementation notes under the auditability row.
- Modify: `README.md`
  - Add a short truthful note about dispute-ready receipt lite.
- Modify: `mkdocs.yml`
  - Add the new receipt doc to nav.

## Task 1: Introduce The Receipt Domain And Storage Model

**Files:**
- Create: `internal/receipts/types.go`
- Create: `internal/receipts/store.go`
- Create: `internal/receipts/store_test.go`
- Modify: `internal/ent/schema/*.go`
- Modify: generated Ent files under `internal/ent/...`

- [ ] **Step 1: Write the failing receipt-store tests**

Create `internal/receipts/store_test.go`:

```go
package receipts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSubmissionReceipt_CreatesTransactionAndCurrentPointer(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-1",
		ArtifactLabel:       "research-memo-v1",
		PayloadHash:         "hash-1",
		SourceLineageDigest: "lineage-1",
	})
	require.NoError(t, err)
	require.NotEmpty(t, sub.SubmissionReceiptID)
	require.NotEmpty(t, tx.TransactionReceiptID)
	require.Equal(t, sub.SubmissionReceiptID, tx.CurrentSubmissionReceiptID)
}

func TestAppendReceiptEvent_PreservesCanonicalReceiptAndTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-2",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-2",
		SourceLineageDigest: "lineage-2",
	})
	require.NoError(t, err)

	err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, EventApprovalRequested)
	require.NoError(t, err)

	got, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, ApprovalPending, got.CanonicalApprovalStatus)
	require.Len(t, events, 1)
}
```

- [ ] **Step 2: Run the new tests and confirm they fail**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
FAIL
```

with missing package / missing symbol errors.

- [ ] **Step 3: Implement the minimal receipt domain**

Create `internal/receipts/types.go`:

```go
package receipts

type ApprovalStatus string

const (
	ApprovalPending          ApprovalStatus = "pending"
	ApprovalApproved         ApprovalStatus = "approved"
	ApprovalRejected         ApprovalStatus = "rejected"
	ApprovalRevisionRequested ApprovalStatus = "revision-requested"
	ApprovalEscalated        ApprovalStatus = "escalated"
)

type SettlementStatus string

const (
	SettlementPending          SettlementStatus = "pending"
	SettlementPartiallySettled SettlementStatus = "partially-settled"
	SettlementSettled          SettlementStatus = "settled"
	SettlementDisputed         SettlementStatus = "disputed"
)

type EventType string

const (
	EventDraftExportability   EventType = "draft_exportability"
	EventFinalExportability   EventType = "final_exportability"
	EventApprovalRequested    EventType = "approval_requested"
	EventApprovalResolved     EventType = "approval_resolved"
	EventSettlementUpdated    EventType = "settlement_updated"
	EventEscalated            EventType = "escalated"
	EventDisputed             EventType = "disputed"
)

type ProvenanceSummary struct {
	ReferenceID        string `json:"reference_id"`
	ConfigFingerprint  string `json:"config_fingerprint,omitempty"`
	SignerSummary      string `json:"signer_summary,omitempty"`
	AttributionSummary string `json:"attribution_summary,omitempty"`
}

type SubmissionReceipt struct {
	SubmissionReceiptID     string           `json:"submission_receipt_id"`
	TransactionReceiptID    string           `json:"transaction_receipt_id"`
	ArtifactLabel           string           `json:"artifact_label"`
	PayloadHash             string           `json:"payload_hash"`
	SourceLineageDigest     string           `json:"source_lineage_digest"`
	CanonicalApprovalStatus ApprovalStatus   `json:"canonical_approval_status"`
	CanonicalSettlementHint string           `json:"canonical_settlement_hint,omitempty"`
	ProvenanceSummary       ProvenanceSummary `json:"provenance_summary"`
}

type TransactionReceipt struct {
	TransactionReceiptID    string           `json:"transaction_receipt_id"`
	TransactionID           string           `json:"transaction_id"`
	CurrentSubmissionReceiptID string        `json:"current_submission_receipt_id,omitempty"`
	CanonicalApprovalStatus ApprovalStatus   `json:"canonical_approval_status"`
	CanonicalSettlementStatus SettlementStatus `json:"canonical_settlement_status"`
}
```

Create `internal/receipts/store.go` with a minimal in-memory store first:

```go
package receipts

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type CreateSubmissionInput struct {
	TransactionID       string
	ArtifactLabel       string
	PayloadHash         string
	SourceLineageDigest string
}

type ReceiptEvent struct {
	SubmissionReceiptID string
	Type                EventType
}

type Store struct {
	mu           sync.Mutex
	submissions   map[string]SubmissionReceipt
	transactions  map[string]TransactionReceipt
	events        map[string][]ReceiptEvent
	txByExternalID map[string]string
}

func NewStore() *Store {
	return &Store{
		submissions:   map[string]SubmissionReceipt{},
		transactions:  map[string]TransactionReceipt{},
		events:        map[string][]ReceiptEvent{},
		txByExternalID: map[string]string{},
	}
}

func (s *Store) CreateSubmissionReceipt(_ context.Context, in CreateSubmissionInput) (SubmissionReceipt, TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	txID, ok := s.txByExternalID[in.TransactionID]
	if !ok {
		txID = uuid.NewString()
		s.txByExternalID[in.TransactionID] = txID
		s.transactions[txID] = TransactionReceipt{
			TransactionReceiptID:    txID,
			TransactionID:           in.TransactionID,
			CanonicalApprovalStatus: ApprovalPending,
			CanonicalSettlementStatus: SettlementPending,
		}
	}

	subID := uuid.NewString()
	sub := SubmissionReceipt{
		SubmissionReceiptID:      subID,
		TransactionReceiptID:     txID,
		ArtifactLabel:            in.ArtifactLabel,
		PayloadHash:              in.PayloadHash,
		SourceLineageDigest:      in.SourceLineageDigest,
		CanonicalApprovalStatus:  ApprovalPending,
	}
	s.submissions[subID] = sub

	tx := s.transactions[txID]
	tx.CurrentSubmissionReceiptID = subID
	s.transactions[txID] = tx

	return sub, tx, nil
}

func (s *Store) AppendReceiptEvent(_ context.Context, submissionID string, eventType EventType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.submissions[submissionID]
	if !ok {
		return fmt.Errorf("submission receipt not found")
	}
	switch eventType {
	case EventApprovalRequested:
		sub.CanonicalApprovalStatus = ApprovalPending
	case EventApprovalResolved:
		// first slice keeps canonical change external to event append
	}
	s.submissions[submissionID] = sub
	s.events[submissionID] = append(s.events[submissionID], ReceiptEvent{
		SubmissionReceiptID: submissionID,
		Type:                eventType,
	})
	return nil
}

func (s *Store) GetSubmissionReceipt(_ context.Context, submissionID string) (SubmissionReceipt, []ReceiptEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.submissions[submissionID]
	if !ok {
		return SubmissionReceipt{}, nil, fmt.Errorf("submission receipt not found")
	}
	return sub, append([]ReceiptEvent(nil), s.events[submissionID]...), nil
}
```

- [ ] **Step 4: Run the targeted tests and make sure they pass**

Run:

```bash
go test ./internal/receipts/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the receipt-domain slice**

Run:

```bash
git add internal/receipts/types.go internal/receipts/store.go internal/receipts/store_test.go
git -c commit.gpgsign=false commit -m "feat: add dispute-ready receipt model"
```

## Task 2: Add Minimal Receipt Integration Surface

**Files:**
- Modify: `internal/app/tools_meta.go`
- Create: `internal/app/tools_meta_receipts_test.go`

- [ ] **Step 1: Write the failing receipt integration tests**

Create `internal/app/tools_meta_receipts_test.go`:

```go
package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/receipts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMetaTools_IncludesCreateDisputeReadyReceipt(t *testing.T) {
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil)
	names := toolNamesUnsorted(tools)
	assert.Contains(t, names, "create_dispute_ready_receipt")
}

func TestCreateDisputeReadyReceiptTool_CreatesSubmissionAndTransaction(t *testing.T) {
	store := receipts.NewStore()
	tools := buildMetaTools(nil, nil, nil, config.SkillConfig{}, nil)

	var tool *agent.Tool
	for _, tdef := range tools {
		if tdef.Name == "create_dispute_ready_receipt" {
			tool = tdef
			break
		}
	}
	require.NotNil(t, tool)

	ctx := context.Background()
	got, err := tool.Handler(ctx, map[string]interface{}{
		"transaction_id": "tx-1",
		"artifact_label": "memo-v1",
		"payload_hash": "hash-1",
		"source_lineage_digest": "lineage-1",
	})
	require.NoError(t, err)
	payload := got.(map[string]interface{})
	assert.NotEmpty(t, payload["submission_receipt_id"])
	assert.NotEmpty(t, payload["transaction_receipt_id"])
}
```

- [ ] **Step 2: Run the new tests and confirm they fail**

Run:

```bash
go test ./internal/app/... -run 'Test(BuildMetaTools_IncludesCreateDisputeReadyReceipt|CreateDisputeReadyReceiptTool_CreatesSubmissionAndTransaction)' -count=1
```

Expected:

```text
FAIL
```

because the tool does not exist yet.

- [ ] **Step 3: Implement a narrow creation tool**

Modify `internal/app/tools_meta.go` to add:

```go
{
	Name:        "create_dispute_ready_receipt",
	Description: "Create a lite dispute-ready receipt for an artifact submission",
	SafetyLevel: agent.SafetyLevelModerate,
	Capability: agent.ToolCapability{
		Category: "knowledge",
		Activity: agent.ActivityWrite,
	},
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"transaction_id": map[string]interface{}{"type": "string"},
			"artifact_label": map[string]interface{}{"type": "string"},
			"payload_hash": map[string]interface{}{"type": "string"},
			"source_lineage_digest": map[string]interface{}{"type": "string"},
		},
		"required": []string{"transaction_id", "artifact_label", "payload_hash", "source_lineage_digest"},
	},
	Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		transactionID, err := toolparam.RequireString(params, "transaction_id")
		if err != nil {
			return nil, err
		}
		artifactLabel, err := toolparam.RequireString(params, "artifact_label")
		if err != nil {
			return nil, err
		}
		payloadHash, err := toolparam.RequireString(params, "payload_hash")
		if err != nil {
			return nil, err
		}
		lineageDigest, err := toolparam.RequireString(params, "source_lineage_digest")
		if err != nil {
			return nil, err
		}

		sub, tx, err := receiptStore.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
			TransactionID:       transactionID,
			ArtifactLabel:       artifactLabel,
			PayloadHash:         payloadHash,
			SourceLineageDigest: lineageDigest,
		})
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"submission_receipt_id": sub.SubmissionReceiptID,
			"transaction_receipt_id": tx.TransactionReceiptID,
			"current_submission_receipt_id": tx.CurrentSubmissionReceiptID,
		}, nil
	},
},
```

Keep this first slice narrow: do not yet wire full exportability/approval/settlement/provenance population.

- [ ] **Step 4: Run the targeted tests and make sure they pass**

Run:

```bash
go test ./internal/app/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the integration slice**

Run:

```bash
git add internal/app/tools_meta.go internal/app/tools_meta_receipts_test.go
git -c commit.gpgsign=false commit -m "feat: add dispute-ready receipt creation tool"
```

## Task 3: Add Minimal Operator Surface And Docs

**Files:**
- Create: `docs/security/dispute-ready-receipts.md`
- Modify: `docs/security/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `docs/architecture/trust-security-policy-audit.md`
- Modify: `README.md`
- Modify: `mkdocs.yml`

- [ ] **Step 1: Write the new operator doc**

Create `docs/security/dispute-ready-receipts.md`:

```md
# Dispute-Ready Receipts

Lango's first dispute-ready receipt slice introduces dedicated submission and transaction receipts.

## Current First Slice

- `submission receipt`
- `transaction receipt`
- `current submission pointer`
- canonical approval / settlement status
- append-only event trail
- lite provenance summary and external references

## Not Yet Included

- dispute adjudication
- human dispute UI
- full settlement execution
- full evidence graph
```

- [ ] **Step 2: Link and truth-align docs**

Modify `docs/security/index.md` quick links:

```md
- [Dispute-Ready Receipts](dispute-ready-receipts.md) -- Lite submission and transaction receipts for later dispute handling
```

Modify `docs/architecture/p2p-knowledge-exchange-track.md` so dispute-ready receipts are no longer entirely pending; note that the lite receipt model has landed and the remaining work is deeper provenance/settlement/dispute integration.

Modify `docs/architecture/trust-security-policy-audit.md` under the auditability row with post-implementation notes:

```md
### Post-Implementation Notes

- A lite dispute-ready receipt model now exists with submission and transaction receipt structure.
- Full dispute adjudication, deeper provenance embedding, and full evidence-graph linkage remain follow-on work.
```

Modify `README.md` with one short truthful note that knowledge exchange now has lite dispute-ready receipts above exportability and approval flow.

Modify `mkdocs.yml`:

```yaml
  - Security:
    - security/index.md
    - Encryption & Secrets: security/encryption.md
    - PII Redaction: security/pii-redaction.md
    - Exportability: security/exportability.md
    - Approval Flow: security/approval-flow.md
    - Dispute-Ready Receipts: security/dispute-ready-receipts.md
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
git add docs/security/dispute-ready-receipts.md docs/security/index.md docs/architecture/p2p-knowledge-exchange-track.md docs/architecture/trust-security-policy-audit.md README.md mkdocs.yml
git -c commit.gpgsign=false commit -m "docs: add dispute-ready receipt operator surface"
```

## Task 4: Full Verification And OpenSpec Closeout

**Files:**
- Modify: `openspec/changes/dispute-ready-receipts-first-slice/*` or create the change if missing
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

- receipt domain model
- receipt creation surface
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
git -c commit.gpgsign=false commit -m "specs: archive dispute-ready receipts first slice"
```
