package receipts

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/ctxkeys"
)

func TestWave27LifecycleQueryFilteringAndCloneIsolation(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	opened, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "wave27-lifecycle",
		Counterparty:   "did:lango:wave27-peer",
		RequestedScope: "artifact/wave27-report",
		PriceContext:   "quote:7.50-usdc",
		TrustContext:   "trust:0.94",
	})
	require.NoError(t, err)

	first, tx := createSubmittedTransaction(t, store, ctx, "wave27-lifecycle")
	require.Equal(t, opened.TransactionReceiptID, tx.TransactionReceiptID)
	require.Equal(t, "did:lango:wave27-peer", tx.Counterparty)
	require.Equal(t, RuntimeStatusOpened, tx.KnowledgeExchangeRuntimeStatus)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, wave27EscrowInput("wave27 original escrow"))
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, first.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "created")
	require.NoError(t, err)

	require.NoError(t, store.AppendPaymentExecutionDenied(ctx, first.SubmissionReceiptID, "stale submission policy denial"))

	second, nextTx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "wave27-lifecycle",
		ArtifactLabel:       "artifact/wave27-report-v2",
		PayloadHash:         "hash-wave27-v2",
		SourceLineageDigest: "lineage-wave27-v2",
	})
	require.NoError(t, err)
	require.Equal(t, tx.TransactionReceiptID, nextTx.TransactionReceiptID)
	require.Equal(t, second.SubmissionReceiptID, nextTx.CurrentSubmissionReceiptID)
	require.Empty(t, nextTx.EscrowExecutionStatus)
	require.Empty(t, nextTx.EscrowReference)
	require.Empty(t, nextTx.EscrowAdjudication)
	require.Nil(t, nextTx.EscrowExecutionInput)
	require.NoError(t, store.AppendReceiptEvent(ctx, second.SubmissionReceiptID, EventFinalExportability))

	submissions, err := store.ListSubmissionReceipts(ctx)
	require.NoError(t, err)
	require.Len(t, submissions, 2)
	sort.Slice(submissions, func(i, j int) bool {
		return submissions[i].ArtifactLabel < submissions[j].ArtifactLabel
	})
	assert.Equal(t, []string{"artifact-wave27-lifecycle", "artifact/wave27-report-v2"}, []string{submissions[0].ArtifactLabel, submissions[1].ArtifactLabel})
	for _, submission := range submissions {
		assert.Equal(t, tx.TransactionReceiptID, submission.TransactionReceiptID)
	}

	transactions, err := store.ListTransactionReceipts(ctx)
	require.NoError(t, err)
	require.Len(t, transactions, 1)
	listed := transactions[0]
	require.Equal(t, second.SubmissionReceiptID, listed.CurrentSubmissionReceiptID)
	require.Equal(t, "artifact/wave27-report", listed.RequestedScope)
	require.Nil(t, listed.EscrowExecutionInput)

	_, firstEvents, err := store.GetSubmissionReceipt(ctx, first.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, firstEvents, 2)
	assert.Equal(t, "payment_execution", firstEvents[1].Source)
	assert.Equal(t, "stale submission policy denial", firstEvents[1].Reason)

	_, secondEvents, err := store.GetSubmissionReceipt(ctx, second.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, secondEvents, 1)
	assert.Equal(t, EventFinalExportability, secondEvents[0].Type)

	_, _, err = store.GetSubmissionReceipt(ctx, "missing-wave27-submission")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
	_, err = store.GetTransactionReceipt(ctx, "missing-wave27-transaction")
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
}

func TestWave27OpenTransactionAndSettlementCloseoutErrorBranches(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "  ",
		Counterparty:   "did:lango:peer",
		RequestedScope: "artifact/scope",
	})
	require.ErrorIs(t, err, ErrInvalidSubmissionInput)

	current, tx := createSubmittedTransaction(t, store, ctx, "wave27-closeout")

	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: "missing-wave27-tx",
		SubmissionReceiptID:  current.SubmissionReceiptID,
		RuntimeReference:     "settlement-missing-tx",
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  "missing-wave27-submission",
		RuntimeReference:     "settlement-missing-sub",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  current.SubmissionReceiptID,
		RuntimeReference:     "settlement-before-approval",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	stale := current
	current, tx, err = store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "wave27-closeout",
		ArtifactLabel:       "wave27-current",
		PayloadHash:         "hash-wave27-current",
		SourceLineageDigest: "lineage-wave27-current",
	})
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)

	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  stale.SubmissionReceiptID,
		RuntimeReference:     "settlement-stale-submission",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	err = store.RecordSettlementFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  stale.SubmissionReceiptID,
		Reason:               "failure belongs to stale submission",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	closed, err := store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  current.SubmissionReceiptID,
		RuntimeReference:     "settlement-wave27-success",
	})
	require.NoError(t, err)
	assert.Equal(t, SettlementProgressionSettled, closed.SettlementProgressionStatus)

	_, events, err := store.GetSubmissionReceipt(ctx, current.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "settlement-wave27-success", events[1].Reason)
}

func TestWave27EscrowEvidenceFilteringRejectsMismatchedReceipts(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	submission, tx := wave27FundedDisputeReadyTransaction(t, store, ctx, "wave27-escrow-errors")

	foreignSubmission, foreignTx := createSubmittedTransaction(t, store, ctx, "wave27-foreign")

	err := store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  foreignSubmission.SubmissionReceiptID,
		RuntimeReference:     "foreign-refund",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "wrong-escrow",
		RuntimeReference:     "wrong-hold",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	err = store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "wrong-escrow",
		Reason:               "wrong hold reference",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-wave27-escrow-errors",
		Outcome:              EscrowAdjudicationDecision("split"),
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	err = store.RecordEscrowAdjudicationFailure(ctx, EscrowAdjudicationFailureRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "wrong-escrow",
		Reason:               "mismatched adjudication reference",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	_, err = store.BindEscrowExecutionInput(ctx, foreignTx.TransactionReceiptID, submission.SubmissionReceiptID, wave27EscrowInput("foreign bind"))
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, SettlementProgressionDisputeReady, stored.SettlementProgressionStatus)
	assert.Equal(t, EscrowExecutionStatusFunded, stored.EscrowExecutionStatus)
	assert.Empty(t, stored.EscrowAdjudication)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	assert.False(t, wave27HasEvent(events, "dispute_hold", "held"))
}

func TestWave27PostAdjudicationRecoveryErrorsAndActorTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	submission, tx := wave27AdjudicatedReleaseTransaction(t, store, ctx, "wave27-post-adjudication")

	err := store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRefund,
		AttemptCount:         2,
		NextRetryAt:          time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC),
		DispatchReference:    "dispatch-wrong-outcome",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	err = store.RecordPostAdjudicationDeadLetter(ctx, PostAdjudicationDeadLetterRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRefund,
		AttemptCount:         2,
		Reason:               "wrong outcome",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	nextRetryAt := time.Date(2026, 5, 19, 12, 30, 0, 0, time.UTC)
	err = store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         3,
		NextRetryAt:          nextRetryAt,
		DispatchReference:    "dispatch-release",
	})
	require.NoError(t, err)

	operatorCtx := ctxkeys.WithPrincipal(ctx, "operator:wave27")
	err = store.RecordManualRetryRequested(operatorCtx, ManualRetryRequestedRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		Reason:               "operator requested replay",
	})
	require.NoError(t, err)

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, EscrowAdjudicationRelease, stored.EscrowAdjudication)
	assert.Equal(t, SettlementProgressionApprovedForSettlement, stored.SettlementProgressionStatus)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.True(t, len(events) >= 2)
	retryEvent := events[len(events)-2]
	assert.Equal(t, PostAdjudicationRecoveryEventSource, retryEvent.Source)
	assert.Equal(t, PostAdjudicationRetryScheduledSubtype, retryEvent.Subtype)
	assert.Contains(t, retryEvent.Reason, "attempt=3")
	assert.Contains(t, retryEvent.Reason, "next_retry_at=2026-05-19T12:30:00Z")
	assert.Contains(t, retryEvent.Reason, "dispatch_reference=dispatch-release")

	manualEvent := events[len(events)-1]
	assert.Equal(t, PostAdjudicationManualRetryRequestedSubtype, manualEvent.Subtype)
	assert.True(t, strings.HasPrefix(manualEvent.Reason, "actor=operator:wave27 manual_replay_at="), manualEvent.Reason)
	assert.Contains(t, manualEvent.Reason, "reason=operator requested replay")
}

func TestWave27ValidationAndPaymentExecutionErrorBranches(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	submission, tx := createSubmittedTransaction(t, store, ctx, "wave27-validation")

	err := store.AppendPaymentExecutionDenied(ctx, submission.SubmissionReceiptID, "policy denied")
	require.NoError(t, err)
	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	require.Len(t, events, 1)

	events[0].Reason = "mutated by caller"
	_, freshEvents, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, "policy denied", freshEvents[0].Reason)

	require.NoError(t, validatePaymentApprovalStatus(PaymentApprovalPending))
	require.ErrorIs(t, validatePaymentApprovalStatus(PaymentApprovalStatus("unknown")), ErrInvalidPaymentApprovalStatus)

	err = store.AppendPaymentExecutionDenied(ctx, "missing-wave27-submission", "policy denied")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.ApplyKnowledgeExchangeRuntimeProgression(ctx, "missing-wave27-transaction", RuntimeStatusOpened, "")
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventSettlementUpdated, "")
	require.True(t, errors.Is(err, ErrInvalidEscrowExecutionState) || errors.Is(err, ErrInvalidEscrowExecutionStatus), "unexpected error: %v", err)
}

func wave27FundedDisputeReadyTransaction(t *testing.T, store *Store, ctx context.Context, transactionID string) (SubmissionReceipt, TransactionReceipt) {
	t.Helper()

	submission, tx := createSubmittedTransaction(t, store, ctx, transactionID)
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, wave27EscrowInput("wave27 escrow "+transactionID))
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "escrow started")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "escrow created")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-"+transactionID, EventEscrowExecutionFunded, "escrow funded")
	require.NoError(t, err)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	tx, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)

	return submission, tx
}

func wave27AdjudicatedReleaseTransaction(t *testing.T, store *Store, ctx context.Context, transactionID string) (SubmissionReceipt, TransactionReceipt) {
	t.Helper()

	submission, tx := wave27FundedDisputeReadyTransaction(t, store, ctx, transactionID)
	err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-" + transactionID,
		RuntimeReference:     "hold-" + transactionID,
	})
	require.NoError(t, err)
	tx, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      "escrow-" + transactionID,
		Outcome:              EscrowAdjudicationRelease,
		Reason:               "release after wave27 review",
	})
	require.NoError(t, err)

	return submission, tx
}

func wave27EscrowInput(reason string) EscrowExecutionInput {
	return EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer-wave27",
		SellerDID: "did:lango:seller-wave27",
		Amount:    "7.50",
		Reason:    reason,
		TaskID:    "task-wave27",
		Milestones: []EscrowMilestoneInput{
			{Description: "deliverable", Amount: "7.50"},
		},
	}
}

func wave27HasEvent(events []ReceiptEvent, source string, subtype string) bool {
	for _, event := range events {
		if event.Source == source && event.Subtype == subtype {
			return true
		}
	}
	return false
}
