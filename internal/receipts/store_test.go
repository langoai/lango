package receipts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	return NewStore()
}

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
	require.Equal(t, ApprovalPending, sub.CanonicalApprovalStatus)
	require.Equal(t, SettlementPending, tx.CanonicalSettlementStatus)
	require.Equal(t, PaymentApprovalPending, tx.CurrentPaymentApprovalStatus)
	require.Empty(t, tx.CanonicalPaymentApprovalDecision)
	require.Empty(t, tx.CanonicalPaymentSettlementHint)
}

func TestCreateSubmissionReceipt_RejectsEmptyInput(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{})
	require.ErrorIs(t, err, ErrInvalidSubmissionInput)
}

func TestCreateSubmissionReceipt_UpdatesCurrentPointerOnSecondSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-3",
		ArtifactLabel:       "memo-a",
		PayloadHash:         "hash-a",
		SourceLineageDigest: "lineage-a",
	})
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-3",
		ArtifactLabel:       "memo-b",
		PayloadHash:         "hash-b",
		SourceLineageDigest: "lineage-b",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)
	require.NotEqual(t, first.SubmissionReceiptID, second.SubmissionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)

	got, events, err := store.GetSubmissionReceipt(ctx, second.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, "memo-b", got.ArtifactLabel)
	require.Empty(t, events)
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
	require.Equal(t, EventApprovalRequested, events[0].Type)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
}

func TestApplyUpfrontPaymentApproval_UpdatesTransactionAndAppendsEvent(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-approval",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-approval",
		SourceLineageDigest: "lineage-approval",
	})
	require.NoError(t, err)

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, PaymentApprovalApproved, "approve", "prepay")
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, updatedTx.TransactionReceiptID)
	require.Equal(t, PaymentApprovalApproved, updatedTx.CurrentPaymentApprovalStatus)
	require.Equal(t, "approve", updatedTx.CanonicalPaymentApprovalDecision)
	require.Equal(t, "prepay", updatedTx.CanonicalPaymentSettlementHint)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, EventPaymentApproval, events[0].Type)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
}

func TestAppendReceiptEvent_RejectsInvalidEventType(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-4",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-4",
		SourceLineageDigest: "lineage-4",
	})
	require.NoError(t, err)

	err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, "")
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)

	err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, EventType("unknown"))
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)
}

func TestAppendReceiptEvent_RejectsMissingSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.AppendReceiptEvent(ctx, "missing-submission", EventApprovalRequested)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}
