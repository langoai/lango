package receipts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWave42SettlementLifecycleRejectsMissingAndForeignCurrentSubmissions(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.ApplySettlementProgression(ctx, "missing-wave42-tx", SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	submission, tx := createSubmittedTransaction(t, store, ctx, "wave42-missing-current")
	wave42DeleteSubmission(store, submission.SubmissionReceiptID)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	store = newTestStore(t)
	current, tx := createSubmittedTransaction(t, store, ctx, "wave42-foreign-current")
	foreign, _ := createSubmittedTransaction(t, store, ctx, "wave42-foreign-submission")
	tx.CurrentSubmissionReceiptID = foreign.SubmissionReceiptID
	wave42PutTransaction(store, tx)

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	tx.CurrentSubmissionReceiptID = current.SubmissionReceiptID
	wave42PutTransaction(store, tx)
	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  foreign.SubmissionReceiptID,
		RuntimeReference:     "settlement-with-foreign-submission",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}

func TestWave42CloneIsolationAcrossEscrowInputBoundaries(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	opened, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "wave42-clone",
		Counterparty:   "did:lango:wave42-peer",
		RequestedScope: "artifact/wave42",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.80",
	})
	require.NoError(t, err)
	submission, tx := createSubmittedTransaction(t, store, ctx, "wave42-clone")
	require.Equal(t, opened.TransactionReceiptID, tx.TransactionReceiptID)

	input := wave42EscrowInput("original clone input")
	bound, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, input)
	require.NoError(t, err)
	input.Milestones[0].Amount = "999.00"
	bound.EscrowExecutionInput.Milestones[0].Description = "mutated returned milestone"

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, "1.00", stored.EscrowExecutionInput.Milestones[0].Amount)
	require.Equal(t, "draft", stored.EscrowExecutionInput.Milestones[0].Description)

	reopened, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "wave42-clone",
		Counterparty:   "did:lango:wave42-peer",
		RequestedScope: "artifact/wave42",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.80",
	})
	require.NoError(t, err)
	reopened.EscrowExecutionInput.Milestones[0].Amount = "777.00"

	stored, err = store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, "1.00", stored.EscrowExecutionInput.Milestones[0].Amount)
	require.Equal(t, "original clone input", stored.EscrowExecutionInput.Reason)
}

func TestWave42EscrowRefundAndHoldEvidenceCoversGuardsAndFallbackTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: "missing-wave42-refund-tx",
		SubmissionReceiptID:  "missing-wave42-refund-submission",
		RuntimeReference:     "refund-missing",
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	submission, tx := createSubmittedTransaction(t, store, ctx, "wave42-refund-missing-submission")
	err = store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  "missing-wave42-refund-submission",
		RuntimeReference:     "refund-missing-submission",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	second, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "wave42-refund-missing-submission",
		ArtifactLabel:       "artifact-wave42-refund-second",
		PayloadHash:         "hash-wave42-refund-second",
		SourceLineageDigest: "lineage-wave42-refund-second",
	})
	require.NoError(t, err)
	require.Equal(t, second.SubmissionReceiptID, tx.CurrentSubmissionReceiptID)
	err = store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RuntimeReference:     "refund-stale-submission",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	err = store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  second.SubmissionReceiptID,
		RuntimeReference:     "refund-before-funded",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	holdSubmission, holdTx := wave42FundedDisputeReadyTransaction(t, store, ctx, "wave42-hold-fallback")
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: holdTx.TransactionReceiptID,
		SubmissionReceiptID:  holdSubmission.SubmissionReceiptID,
		EscrowReference:      "escrow-wave42-hold-fallback",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, holdSubmission.SubmissionReceiptID)
	require.NoError(t, err)
	event := wave42RequireEvent(t, events, "dispute_hold", "held")
	assert.Equal(t, "escrow-wave42-hold-fallback", event.Reason)
}

func TestWave42AdjudicationAndRecoveryRejectCorruptTransactionState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	submission, tx := wave42FundedDisputeReadyTransaction(t, store, ctx, "wave42-adjudication-guards")

	_, err := store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: "missing-wave42-adjudication-tx",
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      tx.EscrowReference,
		Outcome:              EscrowAdjudicationRelease,
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      tx.EscrowReference,
		RuntimeReference:     "hold-wave42",
	})
	require.NoError(t, err)
	tx, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      tx.EscrowReference,
		Outcome:              EscrowAdjudicationRelease,
	})
	require.NoError(t, err)

	err = store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: "missing-wave42-retry-tx",
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         1,
		NextRetryAt:          time.Date(2026, 5, 19, 9, 0, 0, 0, time.UTC),
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	wave42DeleteSubmission(store, submission.SubmissionReceiptID)
	err = store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         1,
		NextRetryAt:          time.Date(2026, 5, 19, 9, 5, 0, 0, time.UTC),
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	foreign, _ := createSubmittedTransaction(t, store, ctx, "wave42-retry-foreign")
	tx.CurrentSubmissionReceiptID = foreign.SubmissionReceiptID
	wave42PutTransaction(store, tx)
	err = store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         1,
		NextRetryAt:          time.Date(2026, 5, 19, 9, 10, 0, 0, time.UTC),
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	err = store.RecordManualRetryRequested(ctx, ManualRetryRequestedRequest{
		TransactionReceiptID: "missing-wave42-manual-retry-tx",
		Outcome:              EscrowAdjudicationRelease,
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
}

func TestWave42PartialSettlementGuardsAndSuccessTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: "missing-wave42-partial-tx",
		SubmissionReceiptID:  "missing-wave42-partial-submission",
		RemainingAmount:      "0.50",
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	submission, tx := createSubmittedTransaction(t, store, ctx, "wave42-partial")
	_, err = store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  "missing-wave42-partial-submission",
		RemainingAmount:      "0.50",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	foreign, _ := createSubmittedTransaction(t, store, ctx, "wave42-partial-foreign")
	_, err = store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  foreign.SubmissionReceiptID,
		RemainingAmount:      "0.50",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RemainingAmount:      "0.50",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)
	_, err = store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RemainingAmount:      "0",
	})
	require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)

	partial, err := store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "0.50",
		RemainingAmount:      "0.25",
		RuntimeReference:     "partial-settlement-wave42",
	})
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionPartiallySettled, partial.SettlementProgressionStatus)

	err = store.RecordPartialSettlementSuccess(ctx, PartialSettlementExecutionEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RuntimeReference:     "partial-success-wave42",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	event := wave42RequireEvent(t, events, "partial_settlement_execution", "partially-settled")
	assert.Equal(t, "partial-success-wave42", event.Reason)
}

func TestWave42PaymentAndEscrowExecutionValidationBranches(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	submission, tx := createSubmittedTransaction(t, store, ctx, "wave42-execution-validation")

	err := store.appendPaymentExecutionEvent(ctx, submission.SubmissionReceiptID, EventType("unknown-wave42-event"), "invalid", "invalid")
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)

	_, err = store.BindEscrowExecutionInput(ctx, "missing-wave42-bind-tx", submission.SubmissionReceiptID, wave42EscrowInput("missing tx"))
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, "missing-wave42-bind-submission", wave42EscrowInput("missing submission"))
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventType("unknown-wave42-event"), "invalid event")
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)
	_, err = store.ApplyEscrowExecutionProgress(ctx, "missing-wave42-progress-tx", submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "missing tx")
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, "missing-wave42-progress-submission", EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "missing submission")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, wave42EscrowInput("execution validation"))
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "started")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "created")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "", EventEscrowExecutionFunded, "missing escrow reference")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-wave42-execution-validation", EventEscrowExecutionCreated, "event mismatch")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
}

func wave42FundedDisputeReadyTransaction(t *testing.T, store *Store, ctx context.Context, transactionID string) (SubmissionReceipt, TransactionReceipt) {
	t.Helper()

	submission, tx := createSubmittedTransaction(t, store, ctx, transactionID)
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, wave42EscrowInput("escrow "+transactionID))
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

func wave42EscrowInput(reason string) EscrowExecutionInput {
	return EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer-wave42",
		SellerDID: "did:lango:seller-wave42",
		Amount:    "1.00",
		Reason:    reason,
		TaskID:    "task-wave42",
		Milestones: []EscrowMilestoneInput{
			{Description: "draft", Amount: "1.00"},
		},
	}
}

func wave42DeleteSubmission(store *Store, submissionReceiptID string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.submissions, submissionReceiptID)
}

func wave42PutTransaction(store *Store, transaction TransactionReceipt) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.transactions[transaction.TransactionReceiptID] = transaction
}

func wave42RequireEvent(t *testing.T, events []ReceiptEvent, source, subtype string) ReceiptEvent {
	t.Helper()
	require.NotEmpty(t, events)
	for _, event := range events {
		if event.Source == source && event.Subtype == subtype {
			return event
		}
	}
	t.Fatalf("missing receipt event source=%s subtype=%s in %#v", source, subtype, events)
	return ReceiptEvent{}
}
