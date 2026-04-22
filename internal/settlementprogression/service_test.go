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
			Reason:   "Artifact release approved.",
		},
	})
	require.NoError(t, err)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Outcome.ProgressionStatus)
	require.Equal(t, "approve", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.Transaction.SettlementProgressionStatus)
	require.Equal(t, "approve", result.Transaction.SettlementProgressionReason)
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
	require.Equal(t, "reject", result.Outcome.ProgressionReason)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.Transaction.SettlementProgressionStatus)
	require.Equal(t, "reject", result.Transaction.SettlementProgressionReason)
}
