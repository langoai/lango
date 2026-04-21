package receipts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/paymentapproval"
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
	require.Empty(t, tx.CanonicalDecision)
	require.Empty(t, tx.CanonicalSettlementHint)
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
	require.Equal(t, "manual", events[0].Source)
	require.Equal(t, string(EventApprovalRequested), events[0].Subtype)
}

func TestApplyUpfrontPaymentApproval_ApprovesAndAppendsEventToSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-approval",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-approval",
		SourceLineageDigest: "lineage-approval",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, updatedTx.TransactionReceiptID)
	require.Equal(t, PaymentApprovalApproved, updatedTx.CurrentPaymentApprovalStatus)
	require.Equal(t, "approve", updatedTx.CanonicalDecision)
	require.Equal(t, "prepay", updatedTx.CanonicalSettlementHint)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, EventPaymentApproval, events[0].Type)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
	require.Equal(t, "approval", events[0].Source)
	require.Equal(t, "approval.upfront_payment", events[0].Subtype)
}

func TestApplyUpfrontPaymentApproval_RejectsAndAppendsEventToSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-reject",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-reject",
		SourceLineageDigest: "lineage-reject",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionReject,
		Reason:        "Amount exceeds max prepay policy.",
		SuggestedMode: paymentapproval.ModeReject,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, PaymentApprovalRejected, updatedTx.CurrentPaymentApprovalStatus)
	require.Equal(t, "reject", updatedTx.CanonicalDecision)
	require.Equal(t, "reject", updatedTx.CanonicalSettlementHint)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
	require.Equal(t, EventPaymentApproval, events[0].Type)
}

func TestApplyUpfrontPaymentApproval_EscalatesAndAppendsEventToSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escalate",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-escalate",
		SourceLineageDigest: "lineage-escalate",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionEscalate,
		Reason:        "Trust score is in an edge-case range for upfront payment.",
		SuggestedMode: paymentapproval.ModeEscalate,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, PaymentApprovalEscalated, updatedTx.CurrentPaymentApprovalStatus)
	require.Equal(t, "escalate", updatedTx.CanonicalDecision)
	require.Equal(t, "escalate", updatedTx.CanonicalSettlementHint)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Len(t, events, 1)
	require.Equal(t, sub.SubmissionReceiptID, events[0].SubmissionReceiptID)
	require.Equal(t, EventPaymentApproval, events[0].Type)
}

func TestApplyUpfrontPaymentApproval_RejectsInvalidTransactionOrSubmissionIdentifiers(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-ids",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-ids",
		SourceLineageDigest: "lineage-ids",
	})
	require.NoError(t, err)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	}

	_, err = store.ApplyUpfrontPaymentApproval(ctx, "missing-transaction", sub.SubmissionReceiptID, outcome)
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, "missing-submission", outcome)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	otherSub, otherTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-other",
		ArtifactLabel:       "memo-other",
		PayloadHash:         "hash-other",
		SourceLineageDigest: "lineage-other",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, otherSub.SubmissionReceiptID, outcome)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, otherTx.TransactionReceiptID, sub.SubmissionReceiptID, outcome)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}

func TestApplyUpfrontPaymentApproval_RejectsInvalidOutcomeWithoutMutatingStorage(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-invalid-outcome",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-invalid-outcome",
		SourceLineageDigest: "lineage-invalid-outcome",
	})
	require.NoError(t, err)

	_, err = store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, paymentapproval.Outcome{
		Decision:      paymentapproval.Decision("bogus"),
		Reason:        "Impossible state.",
		SuggestedMode: paymentapproval.ModeEscrow,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidPaymentApprovalStatus)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Empty(t, events)
	require.Equal(t, ApprovalPending, gotSub.CanonicalApprovalStatus)

	updatedTx := store.transactions[tx.TransactionReceiptID]
	require.Equal(t, PaymentApprovalPending, updatedTx.CurrentPaymentApprovalStatus)
	require.Empty(t, updatedTx.CanonicalDecision)
	require.Empty(t, updatedTx.CanonicalSettlementHint)
}

func TestApplyUpfrontPaymentApproval_UsesExplicitSubmissionInMultiSubmissionTransaction(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	first, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-multi",
		ArtifactLabel:       "memo-a",
		PayloadHash:         "hash-a",
		SourceLineageDigest: "lineage-a",
	})
	require.NoError(t, err)

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-multi",
		ArtifactLabel:       "memo-b",
		PayloadHash:         "hash-b",
		SourceLineageDigest: "lineage-b",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)

	outcome := paymentapproval.Outcome{
		Decision:      paymentapproval.DecisionApprove,
		Reason:        "Upfront payment approved.",
		SuggestedMode: paymentapproval.ModePrepay,
	}

	updatedTx, err := store.ApplyUpfrontPaymentApproval(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, outcome)
	require.NoError(t, err)
	require.Equal(t, PaymentApprovalApproved, updatedTx.CurrentPaymentApprovalStatus)

	_, firstEvents, err := store.GetSubmissionReceipt(ctx, first.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, firstEvents, 1)
	require.Equal(t, first.SubmissionReceiptID, firstEvents[0].SubmissionReceiptID)

	_, secondEvents, err := store.GetSubmissionReceipt(ctx, second.SubmissionReceiptID)
	require.NoError(t, err)
	require.Empty(t, secondEvents)
}

func TestBindEscrowExecutionInput_PersistsCanonicalInputOnTransaction(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-bind",
		ArtifactLabel:       "artifact/escrow-bind",
		PayloadHash:         "hash-escrow-bind",
		SourceLineageDigest: "lineage-escrow-bind",
	})
	require.NoError(t, err)

	updated, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "3.50",
		Reason:    "knowledge exchange",
		TaskID:    "task-escrow-bind",
		Milestones: []EscrowMilestoneInput{
			{Description: "draft", Amount: "1.50"},
			{Description: "final", Amount: "2.00"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updated.EscrowExecutionInput)
	require.Equal(t, EscrowExecutionStatusPending, updated.EscrowExecutionStatus)
	require.Equal(t, "did:lango:buyer", updated.EscrowExecutionInput.BuyerDID)
	require.Equal(t, "3.50", updated.EscrowExecutionInput.Amount)

	gotSub, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Equal(t, sub.SubmissionReceiptID, gotSub.SubmissionReceiptID)
	require.Empty(t, events)
}

func TestApplyEscrowExecutionProgress_RecordsCreatedFundedAndFailed(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-progress",
		ArtifactLabel:       "artifact/escrow-progress",
		PayloadHash:         "hash-escrow-progress",
		SourceLineageDigest: "lineage-escrow-progress",
	})
	require.NoError(t, err)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "knowledge exchange",
		TaskID:    "task-escrow-progress",
		Milestones: []EscrowMilestoneInput{
			{Description: "delivery", Amount: "4.00"},
		},
	})
	require.NoError(t, err)

	created, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "")
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusCreated, created.EscrowExecutionStatus)
	require.Empty(t, created.EscrowReference)

	updated, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-1", EventEscrowExecutionFunded, "")
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFunded, updated.EscrowExecutionStatus)
	require.Equal(t, "escrow-1", updated.EscrowReference)

	failed, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, sub.SubmissionReceiptID, EscrowExecutionStatusFailed, "escrow-1", EventEscrowExecutionFailed, "funding reverted")
	require.NoError(t, err)
	require.Equal(t, EscrowExecutionStatusFailed, failed.EscrowExecutionStatus)
	require.Equal(t, "escrow-1", failed.EscrowReference)

	_, events, err := store.GetSubmissionReceipt(ctx, sub.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 3)
	require.Equal(t, EventEscrowExecutionCreated, events[0].Type)
	require.Equal(t, EventEscrowExecutionFunded, events[1].Type)
	require.Equal(t, EventEscrowExecutionFailed, events[2].Type)
	require.Equal(t, "funding reverted", events[2].Reason)
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

func TestAppendReceiptEvent_AllowsEscrowExecutionEventTypes(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sub, _, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "tx-escrow-events",
		ArtifactLabel:       "memo",
		PayloadHash:         "hash-escrow-events",
		SourceLineageDigest: "lineage-escrow-events",
	})
	require.NoError(t, err)

	for _, eventType := range []EventType{
		EventEscrowExecutionStarted,
		EventEscrowExecutionCreated,
		EventEscrowExecutionFunded,
		EventEscrowExecutionFailed,
	} {
		err = store.AppendReceiptEvent(ctx, sub.SubmissionReceiptID, eventType)
		require.NoError(t, err)
	}
}

func TestAppendReceiptEvent_RejectsMissingSubmission(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.AppendReceiptEvent(ctx, "missing-submission", EventApprovalRequested)
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}
