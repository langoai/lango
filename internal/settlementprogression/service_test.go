package settlementprogression

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/approvalflow"
	"github.com/langoai/lango/internal/receipts"
)

func TestApplyReleaseOutcome_ApproveMapsToApprovedForSettlement(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-approve",
		Counterparty:   "did:lango:peer-approve",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionApprove,
			Reason:   "Approved after final review.",
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Outcome.ProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeApprove, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "Approved after final review.", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Transaction.SettlementProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeApprove, result.Transaction.SettlementProgressionReasonCode)
	require.Equal(t, "Approved after final review.", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_ApproveDefaultsReasonWhenMissing(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-approve-default",
		Counterparty:   "did:lango:peer-approve-default",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionApprove,
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Outcome.ProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeApprove, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "Artifact release approved.", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Transaction.SettlementProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeApprove, result.Transaction.SettlementProgressionReasonCode)
	require.Equal(t, "Artifact release approved.", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_TrimsTransactionReceiptIDBeforeLookup(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-trimmed-id",
		Counterparty:   "did:lango:peer-trimmed-id",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: "  " + tx.TransactionReceiptID + "  ",
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionApprove,
			Reason:   "Approved after final review.",
		},
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, result.Transaction.TransactionReceiptID)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Transaction.SettlementProgressionStatus)
}

func TestApplyReleaseOutcome_RejectMapsToReviewNeeded(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-reject",
		Counterparty:   "did:lango:peer-reject",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionReject,
			Reason:   "Artifact release blocked by policy.",
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Outcome.ProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeReject, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "Artifact release blocked by policy.", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Transaction.SettlementProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeReject, result.Transaction.SettlementProgressionReasonCode)
	require.Equal(t, "Artifact release blocked by policy.", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_RejectDefaultsReasonWhenMissing(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-reject-default",
		Counterparty:   "did:lango:peer-reject-default",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionReject,
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionReasonCodeReject, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "Artifact release rejected.", result.Outcome.ProgressionReason)
	require.Equal(t, "Artifact release rejected.", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_EscalateUsesProvidedReason(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-escalate-provided",
		Counterparty:   "did:lango:peer-escalate",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionEscalate,
			Reason:   "manual approval required",
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Outcome.ProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeEscalate, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "manual approval required", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Transaction.SettlementProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeEscalate, result.Transaction.SettlementProgressionReasonCode)
	require.NotEqual(t, receipts.SettlementProgressionReasonCodeReject, result.Transaction.SettlementProgressionReasonCode)
	require.NotEqual(t, receipts.SettlementProgressionReasonCodeRequestRevision, result.Transaction.SettlementProgressionReasonCode)
	require.Equal(t, "manual approval required", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_EscalateDefaultsReason(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-escalate-default",
		Counterparty:   "did:lango:peer-escalate-default",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionEscalate,
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Outcome.ProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeEscalate, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "higher approval needed", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Transaction.SettlementProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeEscalate, result.Transaction.SettlementProgressionReasonCode)
	require.Equal(t, "higher approval needed", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_RequestRevisionPreservesReason(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-request-revision",
		Counterparty:   "did:lango:peer-request-revision",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionRequestRevision,
			Reason:   "Need updated evidence.",
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Outcome.ProgressionStatus)
	require.Equal(t, receipts.SettlementProgressionReasonCodeRequestRevision, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "Need updated evidence.", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionReasonCodeRequestRevision, result.Transaction.SettlementProgressionReasonCode)
	require.Equal(t, "Need updated evidence.", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_RequestRevisionDefaultsReasonWhenMissing(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-request-revision-default",
		Counterparty:   "did:lango:peer-request-revision-default",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionRequestRevision,
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionReasonCodeRequestRevision, result.Outcome.ProgressionReasonCode)
	require.Equal(t, "Artifact release requires revision.", result.Outcome.ProgressionReason)
	require.Equal(t, "Artifact release requires revision.", result.Transaction.SettlementProgressionReason)
}

func TestApplyReleaseOutcome_RequestRevisionPersistsPartialHint(t *testing.T) {
	store := receipts.NewStore()
	svc := NewService(store)
	ctx := context.Background()

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-release-request-revision-partial",
		Counterparty:   "did:lango:peer-request-revision-partial",
		RequestedScope: "artifact/research-note",
		PriceContext:   "quote:0.50-usdc",
		TrustContext:   "trust:0.72",
	})
	require.NoError(t, err)

	result, err := svc.ApplyReleaseOutcome(ctx, ApplyReleaseOutcomeRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		PartialHint:          "settle 40% now, defer the rest",
		Outcome: ReleaseOutcome{
			Decision: approvalflow.DecisionRequestRevision,
			Reason:   "Need updated evidence.",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "settle 40% now, defer the rest", result.Outcome.PartialHint)
	require.Equal(t, "settle 40% now, defer the rest", result.Transaction.PartialSettlementHint)
}
