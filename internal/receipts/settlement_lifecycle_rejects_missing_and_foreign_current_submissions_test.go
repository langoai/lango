package receipts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettlementLifecycleRejectsMissingAndForeignCurrentSubmissions(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.ApplySettlementProgression(ctx, "missing-serverConnectionToolsReturnsSliceCopy2-tx", SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	submission, tx := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-missing-current")
	settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsDeleteSubmission(store, submission.SubmissionReceiptID)
	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	store = newTestStore(t)
	current, tx := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-foreign-current")
	foreign, _ := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-foreign-submission")
	tx.CurrentSubmissionReceiptID = foreign.SubmissionReceiptID
	settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsPutTransaction(store, tx)

	_, err = store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	tx.CurrentSubmissionReceiptID = current.SubmissionReceiptID
	settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsPutTransaction(store, tx)
	_, err = store.MarkSettlementSettled(ctx, SettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  foreign.SubmissionReceiptID,
		RuntimeReference:     "settlement-with-foreign-submission",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
}

func TestCloneIsolationAcrossEscrowInputBoundaries(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	opened, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "serverConnectionToolsReturnsSliceCopy2-clone",
		Counterparty:   "did:lango:serverConnectionToolsReturnsSliceCopy2-peer",
		RequestedScope: "artifact/serverConnectionToolsReturnsSliceCopy2",
		PriceContext:   "quote:1.00-usdc",
		TrustContext:   "trust:0.80",
	})
	require.NoError(t, err)
	submission, tx := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-clone")
	require.Equal(t, opened.TransactionReceiptID, tx.TransactionReceiptID)

	input := settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsEscrowInput("original clone input")
	bound, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, input)
	require.NoError(t, err)
	input.Milestones[0].Amount = "999.00"
	bound.EscrowExecutionInput.Milestones[0].Description = "mutated returned milestone"

	stored, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	require.Equal(t, "1.00", stored.EscrowExecutionInput.Milestones[0].Amount)
	require.Equal(t, "draft", stored.EscrowExecutionInput.Milestones[0].Description)

	reopened, err := store.OpenKnowledgeExchangeTransaction(ctx, OpenTransactionInput{
		TransactionID:  "serverConnectionToolsReturnsSliceCopy2-clone",
		Counterparty:   "did:lango:serverConnectionToolsReturnsSliceCopy2-peer",
		RequestedScope: "artifact/serverConnectionToolsReturnsSliceCopy2",
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

func TestEscrowRefundAndHoldEvidenceCoversGuardsAndFallbackTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	err := store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: "missing-serverConnectionToolsReturnsSliceCopy2-refund-tx",
		SubmissionReceiptID:  "missing-serverConnectionToolsReturnsSliceCopy2-refund-submission",
		RuntimeReference:     "refund-missing",
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	submission, tx := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-refund-missing-submission")
	err = store.RecordEscrowRefundSuccess(ctx, EscrowRefundEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  "missing-serverConnectionToolsReturnsSliceCopy2-refund-submission",
		RuntimeReference:     "refund-missing-submission",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	second, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       "serverConnectionToolsReturnsSliceCopy2-refund-missing-submission",
		ArtifactLabel:       "artifact-serverConnectionToolsReturnsSliceCopy2-refund-second",
		PayloadHash:         "hash-serverConnectionToolsReturnsSliceCopy2-refund-second",
		SourceLineageDigest: "lineage-serverConnectionToolsReturnsSliceCopy2-refund-second",
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

	holdSubmission, holdTx := settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsFundedDisputeReadyTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-hold-fallback")
	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: holdTx.TransactionReceiptID,
		SubmissionReceiptID:  holdSubmission.SubmissionReceiptID,
		EscrowReference:      "escrow-serverConnectionToolsReturnsSliceCopy2-hold-fallback",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, holdSubmission.SubmissionReceiptID)
	require.NoError(t, err)
	event := settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsRequireEvent(t, events, "dispute_hold", "held")
	assert.Equal(t, "escrow-serverConnectionToolsReturnsSliceCopy2-hold-fallback", event.Reason)
}

func TestAdjudicationAndRecoveryRejectCorruptTransactionState(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	submission, tx := settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsFundedDisputeReadyTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-adjudication-guards")

	_, err := store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
		TransactionReceiptID: "missing-serverConnectionToolsReturnsSliceCopy2-adjudication-tx",
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      tx.EscrowReference,
		Outcome:              EscrowAdjudicationRelease,
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	err = store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      tx.EscrowReference,
		RuntimeReference:     "hold-serverConnectionToolsReturnsSliceCopy2",
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
		TransactionReceiptID: "missing-serverConnectionToolsReturnsSliceCopy2-retry-tx",
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         1,
		NextRetryAt:          time.Date(2026, 5, 19, 9, 0, 0, 0, time.UTC),
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsDeleteSubmission(store, submission.SubmissionReceiptID)
	err = store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         1,
		NextRetryAt:          time.Date(2026, 5, 19, 9, 5, 0, 0, time.UTC),
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	foreign, _ := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-retry-foreign")
	tx.CurrentSubmissionReceiptID = foreign.SubmissionReceiptID
	settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsPutTransaction(store, tx)
	err = store.RecordPostAdjudicationRetryScheduled(ctx, PostAdjudicationRetryScheduledRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		Outcome:              EscrowAdjudicationRelease,
		AttemptCount:         1,
		NextRetryAt:          time.Date(2026, 5, 19, 9, 10, 0, 0, time.UTC),
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	err = store.RecordManualRetryRequested(ctx, ManualRetryRequestedRequest{
		TransactionReceiptID: "missing-serverConnectionToolsReturnsSliceCopy2-manual-retry-tx",
		Outcome:              EscrowAdjudicationRelease,
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
}

func TestPartialSettlementGuardsAndSuccessTrail(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: "missing-serverConnectionToolsReturnsSliceCopy2-partial-tx",
		SubmissionReceiptID:  "missing-serverConnectionToolsReturnsSliceCopy2-partial-submission",
		RemainingAmount:      "0.50",
	})
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)

	submission, tx := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-partial")
	_, err = store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  "missing-serverConnectionToolsReturnsSliceCopy2-partial-submission",
		RemainingAmount:      "0.50",
	})
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	foreign, _ := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-partial-foreign")
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
		RuntimeReference:     "partial-settlement-serverConnectionToolsReturnsSliceCopy2",
	})
	require.NoError(t, err)
	require.Equal(t, SettlementProgressionPartiallySettled, partial.SettlementProgressionStatus)

	err = store.RecordPartialSettlementSuccess(ctx, PartialSettlementExecutionEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RuntimeReference:     "partial-success-serverConnectionToolsReturnsSliceCopy2",
	})
	require.NoError(t, err)

	_, events, err := store.GetSubmissionReceipt(ctx, submission.SubmissionReceiptID)
	require.NoError(t, err)
	event := settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsRequireEvent(t, events, "partial_settlement_execution", "partially-settled")
	assert.Equal(t, "partial-success-serverConnectionToolsReturnsSliceCopy2", event.Reason)
}

func TestPaymentAndEscrowExecutionValidationBranches(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	submission, tx := createSubmittedTransaction(t, store, ctx, "serverConnectionToolsReturnsSliceCopy2-execution-validation")

	err := store.appendPaymentExecutionEvent(ctx, submission.SubmissionReceiptID, EventType("unknown-serverConnectionToolsReturnsSliceCopy2-event"), "invalid", "invalid")
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)

	_, err = store.BindEscrowExecutionInput(ctx, "missing-serverConnectionToolsReturnsSliceCopy2-bind-tx", submission.SubmissionReceiptID, settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsEscrowInput("missing tx"))
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, "missing-serverConnectionToolsReturnsSliceCopy2-bind-submission", settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsEscrowInput("missing submission"))
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventType("unknown-serverConnectionToolsReturnsSliceCopy2-event"), "invalid event")
	require.ErrorIs(t, err, ErrInvalidReceiptEventType)
	_, err = store.ApplyEscrowExecutionProgress(ctx, "missing-serverConnectionToolsReturnsSliceCopy2-progress-tx", submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "missing tx")
	require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, "missing-serverConnectionToolsReturnsSliceCopy2-progress-submission", EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "missing submission")
	require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)

	_, err = store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsEscrowInput("execution validation"))
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "started")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "created")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "", EventEscrowExecutionFunded, "missing escrow reference")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, "escrow-serverConnectionToolsReturnsSliceCopy2-execution-validation", EventEscrowExecutionCreated, "event mismatch")
	require.ErrorIs(t, err, ErrInvalidEscrowExecutionState)
}

func settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsFundedDisputeReadyTransaction(t *testing.T, store *Store, ctx context.Context, transactionID string) (SubmissionReceipt, TransactionReceipt) {
	t.Helper()

	submission, tx := createSubmittedTransaction(t, store, ctx, transactionID)
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsEscrowInput("escrow "+transactionID))
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

func settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsEscrowInput(reason string) EscrowExecutionInput {
	return EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer-serverConnectionToolsReturnsSliceCopy2",
		SellerDID: "did:lango:seller-serverConnectionToolsReturnsSliceCopy2",
		Amount:    "1.00",
		Reason:    reason,
		TaskID:    "task-serverConnectionToolsReturnsSliceCopy2",
		Milestones: []EscrowMilestoneInput{
			{Description: "draft", Amount: "1.00"},
		},
	}
}

func settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsDeleteSubmission(store *Store, submissionReceiptID string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.submissions, submissionReceiptID)
}

func settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsPutTransaction(store *Store, transaction TransactionReceipt) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.transactions[transaction.TransactionReceiptID] = transaction
}

func settlementLifecycleRejectsMissingAndForeignCurrentSubmissionsRequireEvent(t *testing.T, events []ReceiptEvent, source, subtype string) ReceiptEvent {
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
