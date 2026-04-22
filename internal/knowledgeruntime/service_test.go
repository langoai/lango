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
	require.Equal(t, submission.SubmissionReceiptID, stored.CurrentSubmissionReceiptID)
}

func TestService_SelectExecutionPath_UsesEscrowBranch(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	svc := NewService(store)

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-rt-3",
		Counterparty:   "did:lango:peer-3",
		RequestedScope: "artifact/design-final",
		PriceContext:   "quote:5.00-usdc",
		TrustContext:   "trust:0.86",
	})
	require.NoError(t, err)

	submission, _, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "deal-rt-3",
		ArtifactLabel:       "artifact/design-final-v1",
		PayloadHash:         "hash-deal-rt-3",
		SourceLineageDigest: "lineage-deal-rt-3",
	})
	require.NoError(t, err)

	updated, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved with escrow",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.NoError(t, err)

	branch, err := svc.SelectExecutionPath(ctx, updated.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, BranchEscrow, branch.Branch)
}

func TestService_SelectExecutionPath_RejectsStaleApprovalForNonCurrentSubmission(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	svc := NewService(store)

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-rt-4",
		Counterparty:   "did:lango:peer-4",
		RequestedScope: "artifact/runtime-multi",
		PriceContext:   "quote:3.00-usdc",
		TrustContext:   "trust:0.88",
	})
	require.NoError(t, err)

	first, _, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "deal-rt-4",
		ArtifactLabel:       "artifact/runtime-a",
		PayloadHash:         "hash-deal-rt-4-a",
		SourceLineageDigest: "lineage-deal-rt-4-a",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved on stale submission",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	second, updatedTx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "deal-rt-4",
		ArtifactLabel:       "artifact/runtime-b",
		PayloadHash:         "hash-deal-rt-4-b",
		SourceLineageDigest: "lineage-deal-rt-4-b",
	})
	require.NoError(t, err)
	require.Equal(t, second.SubmissionReceiptID, updatedTx.CurrentSubmissionReceiptID)

	_, err = svc.SelectExecutionPath(ctx, updatedTx.TransactionReceiptID)
	require.Error(t, err)

	stored, err := store.GetTransactionReceipt(ctx, updatedTx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, receipts.RuntimeStatusOpened, stored.KnowledgeExchangeRuntimeStatus)
	require.Equal(t, second.SubmissionReceiptID, stored.CurrentSubmissionReceiptID)
}

func TestService_SelectExecutionPath_IsIdempotentForRepeatCalls(t *testing.T) {
	ctx := context.Background()
	store := receipts.NewStore()
	svc := NewService(store)

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  "deal-rt-5",
		Counterparty:   "did:lango:peer-5",
		RequestedScope: "artifact/repeat-call",
		PriceContext:   "quote:2.50-usdc",
		TrustContext:   "trust:0.93",
	})
	require.NoError(t, err)

	submission, _, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "deal-rt-5",
		ArtifactLabel:       "artifact/repeat-call-v1",
		PayloadHash:         "hash-deal-rt-5",
		SourceLineageDigest: "lineage-deal-rt-5",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "approved",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	firstSelection, err := svc.SelectExecutionPath(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)

	secondSelection, err := svc.SelectExecutionPath(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, firstSelection, secondSelection)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, receipts.RuntimeStatusPaymentApproved, stored.KnowledgeExchangeRuntimeStatus)
	require.Equal(t, submission.SubmissionReceiptID, stored.CurrentSubmissionReceiptID)
}
