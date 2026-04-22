package knowledgeruntime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

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
	require.NotEmpty(t, tx.TransactionReceiptID)
	require.Equal(t, receipts.RuntimeStatusOpened, tx.RuntimeStatus)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, receipts.RuntimeStatusOpened, stored.KnowledgeExchangeRuntimeStatus)
	require.Equal(t, "did:lango:peer-1", stored.Counterparty)
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

	submission, _, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "deal-rt-2",
		ArtifactLabel:       "artifact/design-draft-v1",
		PayloadHash:         "hash-deal-rt-2",
		SourceLineageDigest: "lineage-deal-rt-2",
	})
	require.NoError(t, err)

	updated, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)
	require.Equal(t, string(paymentapproval.ModePrepay), updated.CanonicalSettlementHint)

	branch, err := svc.SelectExecutionPath(ctx, updated.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, BranchPrepay, branch.Branch)
	require.Equal(t, updated.TransactionReceiptID, branch.TransactionReceiptID)

	stored, err := store.GetTransactionReceipt(ctx, updated.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, receipts.RuntimeStatusPaymentApproved, stored.KnowledgeExchangeRuntimeStatus)
}
