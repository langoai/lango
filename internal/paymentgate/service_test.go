package paymentgate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

func TestService_Decide_AllowsApprovedPrepay(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-allowed",
		ArtifactLabel:       "artifact/allowed",
		PayloadHash:         "hash-allowed",
		SourceLineageDigest: "lineage-allowed",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	service := NewService(store)
	result, err := service.Decide(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.NoError(t, err)
	require.Equal(t, DecisionAllow, result.Decision)
	require.Empty(t, result.DenyReason)
}

func TestService_Decide_DeniesMissingTransactionReceiptID(t *testing.T) {
	service := NewService(receipts.NewStore())

	result, err := service.Decide(context.Background(), Request{})
	require.NoError(t, err)
	require.Equal(t, DecisionDeny, result.Decision)
	require.Equal(t, DenyReasonMissingReceipt, result.DenyReason)
}

func TestService_Decide_DeniesMissingTransactionInStore(t *testing.T) {
	service := NewService(receipts.NewStore())

	result, err := service.Decide(context.Background(), Request{TransactionReceiptID: "missing"})
	require.NoError(t, err)
	require.Equal(t, DecisionDeny, result.Decision)
	require.Equal(t, DenyReasonMissingReceipt, result.DenyReason)
}

func TestService_Decide_DeniesWhenApprovalIsNotApproved(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-not-approved",
		ArtifactLabel:       "artifact/not-approved",
		PayloadHash:         "hash-not-approved",
		SourceLineageDigest: "lineage-not-approved",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionReject,
		Reason:        "Amount exceeds max prepay policy.",
		SuggestedMode: paymentapproval.ModeReject,
	})
	require.NoError(t, err)

	service := NewService(store)
	result, err := service.Decide(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.NoError(t, err)
	require.Equal(t, DecisionDeny, result.Decision)
	require.Equal(t, DenyReasonApprovalNotApproved, result.DenyReason)
}

func TestService_Decide_DeniesWhenExecutionModeIsNotPrepay(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-mode-mismatch",
		ArtifactLabel:       "artifact/mode-mismatch",
		PayloadHash:         "hash-mode-mismatch",
		SourceLineageDigest: "lineage-mode-mismatch",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.NoError(t, err)

	service := NewService(store)
	result, err := service.Decide(ctx, Request{TransactionReceiptID: tx.TransactionReceiptID})
	require.NoError(t, err)
	require.Equal(t, DecisionDeny, result.Decision)
	require.Equal(t, DenyReasonExecutionModeMismatch, result.DenyReason)
}
