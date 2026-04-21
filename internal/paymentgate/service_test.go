package paymentgate

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

func TestService_EvaluateDirectPayment_AllowsApprovedPrepay(t *testing.T) {
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
	result, err := service.EvaluateDirectPayment(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ToolName:             "payment_send",
		Context:              map[string]interface{}{"source": "test"},
	})
	require.NoError(t, err)
	require.Equal(t, Allow, result.Decision)
	require.Empty(t, result.Reason)
}

func TestService_EvaluateDirectPayment_DeniesMissingTransactionReceiptID(t *testing.T) {
	service := NewService(receipts.NewStore())

	result, err := service.EvaluateDirectPayment(context.Background(), Request{})
	require.NoError(t, err)
	require.Equal(t, Deny, result.Decision)
	require.Equal(t, ReasonMissingReceipt, result.Reason)
}

func TestService_EvaluateDirectPayment_DeniesMissingTransactionInStore(t *testing.T) {
	service := NewService(receipts.NewStore())

	result, err := service.EvaluateDirectPayment(context.Background(), Request{
		TransactionReceiptID: "missing",
		SubmissionReceiptID:  "submission-missing",
	})
	require.NoError(t, err)
	require.Equal(t, Deny, result.Decision)
	require.Equal(t, ReasonMissingReceipt, result.Reason)
}

func TestService_EvaluateDirectPayment_DeniesMissingSubmissionReceiptID(t *testing.T) {
	store := receipts.NewStore()
	ctx := context.Background()

	submission, tx, err := store.CreateSubmissionReceipt(ctx, receipts.CreateSubmissionInput{
		TransactionID:       "tx-missing-submission",
		ArtifactLabel:       "artifact/missing-submission",
		PayloadHash:         "hash-missing-submission",
		SourceLineageDigest: "lineage-missing-submission",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	})
	require.NoError(t, err)

	service := NewService(store)
	result, err := service.EvaluateDirectPayment(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
	})
	require.NoError(t, err)
	require.Equal(t, Deny, result.Decision)
	require.Equal(t, ReasonMissingReceipt, result.Reason)
}

func TestService_EvaluateDirectPayment_DeniesWhenApprovalIsNotApproved(t *testing.T) {
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
	result, err := service.EvaluateDirectPayment(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
	})
	require.NoError(t, err)
	require.Equal(t, Deny, result.Decision)
	require.Equal(t, ReasonApprovalNotApproved, result.Reason)
}

func TestService_EvaluateDirectPayment_DeniesWhenExecutionModeIsNotPrepay(t *testing.T) {
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
	result, err := service.EvaluateDirectPayment(ctx, Request{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
	})
	require.NoError(t, err)
	require.Equal(t, Deny, result.Decision)
	require.Equal(t, ReasonExecutionModeMismatch, result.Reason)
}

func TestService_EvaluateDirectPayment_PropagatesUnexpectedStoreErrors(t *testing.T) {
	storeErr := errors.New("boom")
	service := NewService(&fakeReceiptStore{err: storeErr})

	result, err := service.EvaluateDirectPayment(context.Background(), Request{
		TransactionReceiptID: "tx-error",
		SubmissionReceiptID:  "submission-error",
	})
	require.ErrorIs(t, err, storeErr)
	require.Equal(t, Result{}, result)
}

type fakeReceiptStore struct {
	err error
}

func (f *fakeReceiptStore) GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error) {
	return receipts.TransactionReceipt{}, f.err
}
