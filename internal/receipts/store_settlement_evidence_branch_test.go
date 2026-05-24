package receipts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscrowDisputeHoldEvidenceCoversFallbackReasonAndFailureGuards(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, submission, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-success")

	err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		EscrowReference:      escrowRef,
	})
	require.NoError(t, err)
	events := store.events[submission.SubmissionReceiptID]
	last := events[len(events)-1]
	assert.Equal(t, "dispute_hold", last.Source)
	assert.Equal(t, "held", last.Subtype)
	assert.Equal(t, escrowRef, last.Reason)

	updated, err := store.GetTransactionReceipt(ctx, tx.TransactionReceiptID)
	require.NoError(t, err)
	assert.Equal(t, DisputeLifecycleHoldActive, updated.DisputeLifecycleStatus)

	t.Run("failure success appends event", func(t *testing.T) {
		store, submission, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-failure")
		err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			EscrowReference:      escrowRef,
			Reason:               "hold execution reverted",
		})
		require.NoError(t, err)
		events := store.events[submission.SubmissionReceiptID]
		last := events[len(events)-1]
		assert.Equal(t, "dispute_hold", last.Source)
		assert.Equal(t, "failed", last.Subtype)
		assert.Equal(t, EventSettlementExecutionFailed, last.Type)
	})

	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "missing transaction",
			run: func(t *testing.T) {
				store, submission, _, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-missing-tx")
				err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
					TransactionReceiptID: "missing-tx",
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
			},
		},
		{
			name: "missing submission",
			run: func(t *testing.T) {
				store, _, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-missing-sub")
				err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  "missing-submission",
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
			},
		},
		{
			name: "foreign submission",
			run: func(t *testing.T) {
				store, _, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-foreign")
				foreignSubmission, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-foreign-other")
				err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  foreignSubmission.SubmissionReceiptID,
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
			},
		},
		{
			name: "non current submission",
			run: func(t *testing.T) {
				store := newTestStore(t)
				first, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-non-current")
				_, latest := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-non-current")
				err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
					TransactionReceiptID: latest.TransactionReceiptID,
					SubmissionReceiptID:  first.SubmissionReceiptID,
					EscrowReference:      "escrow-non-current",
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
		{
			name: "wrong escrow status",
			run: func(t *testing.T) {
				store := newTestStore(t)
				submission, tx := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-wrong-escrow")
				err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      "escrow-wrong-status",
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
		{
			name: "wrong settlement status",
			run: func(t *testing.T) {
				store, submission, tx, escrowRef := receiptsEvidenceBranchFunded(t, ctx, "receipts-evidence-hold-wrong-settlement")
				err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
		{
			name: "escrow reference mismatch",
			run: func(t *testing.T) {
				store, submission, tx, _ := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-ref-mismatch")
				err := store.RecordEscrowDisputeHoldFailure(ctx, EscrowDisputeHoldFailureRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      "escrow-other",
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestEscrowDisputeHoldSuccessRejectsInvalidEvidenceState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "missing transaction",
			run: func(t *testing.T) {
				store, submission, _, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-success-missing-tx")
				err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
					TransactionReceiptID: "missing-tx",
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
			},
		},
		{
			name: "missing submission",
			run: func(t *testing.T) {
				store, _, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-success-missing-sub")
				err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  "missing-submission",
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
			},
		},
		{
			name: "foreign submission",
			run: func(t *testing.T) {
				store, _, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-success-foreign")
				foreignSubmission, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-success-foreign-other")
				err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  foreignSubmission.SubmissionReceiptID,
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
			},
		},
		{
			name: "non current submission",
			run: func(t *testing.T) {
				store := newTestStore(t)
				first, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-success-non-current")
				_, latest := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-success-non-current")
				err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
					TransactionReceiptID: latest.TransactionReceiptID,
					SubmissionReceiptID:  first.SubmissionReceiptID,
					EscrowReference:      "escrow-non-current",
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
		{
			name: "wrong escrow status",
			run: func(t *testing.T) {
				store := newTestStore(t)
				submission, tx := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-hold-success-wrong-escrow")
				err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      "escrow-wrong-status",
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
		{
			name: "wrong settlement status",
			run: func(t *testing.T) {
				store, submission, tx, escrowRef := receiptsEvidenceBranchFunded(t, ctx, "receipts-evidence-hold-success-wrong-settlement")
				err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      escrowRef,
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
		{
			name: "escrow reference mismatch",
			run: func(t *testing.T) {
				store, submission, tx, _ := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-hold-success-ref-mismatch")
				err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  submission.SubmissionReceiptID,
					EscrowReference:      "escrow-other",
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestEscrowRefundFailureCoversSuccessAndGuardBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, submission, tx, _ := receiptsEvidenceBranchFunded(t, ctx, "receipts-evidence-refund-failure-success")
	reviewNeeded, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	err = store.RecordEscrowRefundFailure(ctx, SettlementFailureRequest{
		TransactionReceiptID: reviewNeeded.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		Reason:               "refund execution reverted",
	})
	require.NoError(t, err)
	last := store.events[submission.SubmissionReceiptID][len(store.events[submission.SubmissionReceiptID])-1]
	assert.Equal(t, "escrow_refund", last.Source)
	assert.Equal(t, "failed", last.Subtype)
	assert.Equal(t, EventSettlementExecutionFailed, last.Type)

	tests := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "missing transaction",
			run: func(t *testing.T) {
				store, submission, _, _ := receiptsEvidenceBranchFunded(t, ctx, "receipts-evidence-refund-failure-missing-tx")
				err := store.RecordEscrowRefundFailure(ctx, SettlementFailureRequest{
					TransactionReceiptID: "missing-tx",
					SubmissionReceiptID:  submission.SubmissionReceiptID,
				})
				require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
			},
		},
		{
			name: "missing submission",
			run: func(t *testing.T) {
				store, _, tx, _ := receiptsEvidenceBranchFunded(t, ctx, "receipts-evidence-refund-failure-missing-sub")
				err := store.RecordEscrowRefundFailure(ctx, SettlementFailureRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  "missing-submission",
				})
				require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
			},
		},
		{
			name: "foreign submission",
			run: func(t *testing.T) {
				store, _, tx, _ := receiptsEvidenceBranchFunded(t, ctx, "receipts-evidence-refund-failure-foreign")
				foreignSubmission, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-refund-failure-foreign-other")
				err := store.RecordEscrowRefundFailure(ctx, SettlementFailureRequest{
					TransactionReceiptID: tx.TransactionReceiptID,
					SubmissionReceiptID:  foreignSubmission.SubmissionReceiptID,
				})
				require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
			},
		},
		{
			name: "non current submission",
			run: func(t *testing.T) {
				store := newTestStore(t)
				first, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-refund-failure-non-current")
				_, latest := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-refund-failure-non-current")
				err := store.RecordEscrowRefundFailure(ctx, SettlementFailureRequest{
					TransactionReceiptID: latest.TransactionReceiptID,
					SubmissionReceiptID:  first.SubmissionReceiptID,
				})
				require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestEscrowAdjudicationResidualValidationBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	t.Run("invalid outcome", func(t *testing.T) {
		store, submission, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-adjudication-invalid-outcome")
		err := store.RecordEscrowDisputeHoldSuccess(ctx, EscrowDisputeHoldEvidenceRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			EscrowReference:      escrowRef,
		})
		require.NoError(t, err)
		_, err = store.ApplyEscrowAdjudication(ctx, EscrowAdjudicationRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			EscrowReference:      escrowRef,
			Outcome:              EscrowAdjudicationDecision("split"),
		})
		require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
	})

	t.Run("failure missing transaction", func(t *testing.T) {
		store, submission, _, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-adjudication-failure-missing-tx")
		err := store.RecordEscrowAdjudicationFailure(ctx, EscrowAdjudicationFailureRequest{
			TransactionReceiptID: "missing-tx",
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			EscrowReference:      escrowRef,
		})
		require.ErrorIs(t, err, ErrTransactionReceiptNotFound)
	})

	t.Run("failure missing submission", func(t *testing.T) {
		store, _, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-adjudication-failure-missing-sub")
		err := store.RecordEscrowAdjudicationFailure(ctx, EscrowAdjudicationFailureRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  "missing-submission",
			EscrowReference:      escrowRef,
		})
		require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
	})

	t.Run("failure foreign submission", func(t *testing.T) {
		store, _, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-adjudication-failure-foreign")
		foreignSubmission, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-adjudication-failure-foreign-other")
		err := store.RecordEscrowAdjudicationFailure(ctx, EscrowAdjudicationFailureRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  foreignSubmission.SubmissionReceiptID,
			EscrowReference:      escrowRef,
		})
		require.ErrorIs(t, err, ErrSubmissionReceiptNotFound)
	})

	t.Run("failure non current submission", func(t *testing.T) {
		store := newTestStore(t)
		first, _ := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-adjudication-failure-non-current")
		_, latest := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, "receipts-evidence-adjudication-failure-non-current")
		err := store.RecordEscrowAdjudicationFailure(ctx, EscrowAdjudicationFailureRequest{
			TransactionReceiptID: latest.TransactionReceiptID,
			SubmissionReceiptID:  first.SubmissionReceiptID,
			EscrowReference:      "escrow-non-current",
		})
		require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
	})

	t.Run("failure escrow reference mismatch", func(t *testing.T) {
		store, submission, tx, escrowRef := receiptsEvidenceBranchFundedDisputeReady(t, ctx, "receipts-evidence-adjudication-failure-ref-mismatch")
		err := store.RecordEscrowAdjudicationFailure(ctx, EscrowAdjudicationFailureRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			EscrowReference:      escrowRef + "-other",
		})
		require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
	})
}

func TestPartialSettlementEvidenceGuardsAndEvents(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, submission, tx := receiptsEvidenceBranchApproved(t, ctx, "receipts-evidence-partial-success")

	partial, err := store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
		TransactionReceiptID: tx.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		ExecutedAmount:       "3.00",
		RemainingAmount:      "1.00",
		RuntimeReference:     "partial-exec-1",
	})
	require.NoError(t, err)
	err = store.RecordPartialSettlementSuccess(ctx, PartialSettlementExecutionEvidenceRequest{
		TransactionReceiptID: partial.TransactionReceiptID,
		SubmissionReceiptID:  submission.SubmissionReceiptID,
		RuntimeReference:     "partial-proof",
	})
	require.NoError(t, err)
	last := store.events[submission.SubmissionReceiptID][len(store.events[submission.SubmissionReceiptID])-1]
	assert.Equal(t, "partial_settlement_execution", last.Source)
	assert.Equal(t, "partially-settled", last.Subtype)
	assert.Equal(t, EventSettlementUpdated, last.Type)

	t.Run("success rejects wrong state", func(t *testing.T) {
		store, submission, tx := receiptsEvidenceBranchApproved(t, ctx, "receipts-evidence-partial-success-wrong-state")
		err := store.RecordPartialSettlementSuccess(ctx, PartialSettlementExecutionEvidenceRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
		})
		require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
	})

	t.Run("failure success appends event", func(t *testing.T) {
		store, submission, tx := receiptsEvidenceBranchApproved(t, ctx, "receipts-evidence-partial-failure")
		err := store.RecordPartialSettlementFailure(ctx, PartialSettlementFailureRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			ExecutedAmount:       "2.00",
			RemainingAmount:      "3.00",
			Reason:               "partial settlement failed",
		})
		require.NoError(t, err)
		last := store.events[submission.SubmissionReceiptID][len(store.events[submission.SubmissionReceiptID])-1]
		assert.Equal(t, "partial_settlement_execution", last.Source)
		assert.Equal(t, "failed", last.Subtype)
		assert.Equal(t, EventSettlementExecutionFailed, last.Type)
	})

	t.Run("failure rejects partially settled state", func(t *testing.T) {
		store, submission, tx := receiptsEvidenceBranchApproved(t, ctx, "receipts-evidence-partial-failure-wrong-state")
		_, err := store.MarkPartialSettlementSettled(ctx, PartialSettlementCloseoutRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			ExecutedAmount:       "3.00",
			RemainingAmount:      "1.00",
		})
		require.NoError(t, err)
		err = store.RecordPartialSettlementFailure(ctx, PartialSettlementFailureRequest{
			TransactionReceiptID: tx.TransactionReceiptID,
			SubmissionReceiptID:  submission.SubmissionReceiptID,
			Reason:               "late failure",
		})
		require.ErrorIs(t, err, ErrInvalidSettlementProgressionState)
	})
}

func receiptsEvidenceBranchCreateSubmitted(t *testing.T, store *Store, ctx context.Context, transactionID string) (SubmissionReceipt, TransactionReceipt) {
	t.Helper()
	if store == nil {
		store = newTestStore(t)
	}
	submission, tx, err := store.CreateSubmissionReceipt(ctx, CreateSubmissionInput{
		TransactionID:       transactionID,
		ArtifactLabel:       "artifact-" + transactionID,
		PayloadHash:         "hash-" + transactionID,
		SourceLineageDigest: "lineage-" + transactionID,
	})
	require.NoError(t, err)
	return submission, tx
}

func receiptsEvidenceBranchApproved(t *testing.T, ctx context.Context, transactionID string) (*Store, SubmissionReceipt, TransactionReceipt) {
	t.Helper()
	store := newTestStore(t)
	submission, tx := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, transactionID)
	approved, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionApprovedForSettlement, SettlementProgressionReasonCodeApprove, "approved", "")
	require.NoError(t, err)
	return store, submission, approved
}

func receiptsEvidenceBranchFunded(t *testing.T, ctx context.Context, transactionID string) (*Store, SubmissionReceipt, TransactionReceipt, string) {
	t.Helper()
	store := newTestStore(t)
	submission, tx := receiptsEvidenceBranchCreateSubmitted(t, store, ctx, transactionID)
	_, err := store.BindEscrowExecutionInput(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionInput{
		BuyerDID:  "did:lango:buyer",
		SellerDID: "did:lango:seller",
		Amount:    "4.00",
		Reason:    "escrow evidence branch",
		Milestones: []EscrowMilestoneInput{{
			Description: "complete work",
			Amount:      "4.00",
		}},
	})
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusPending, "", EventEscrowExecutionStarted, "started")
	require.NoError(t, err)
	_, err = store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusCreated, "", EventEscrowExecutionCreated, "created")
	require.NoError(t, err)
	escrowRef := "escrow-" + transactionID
	funded, err := store.ApplyEscrowExecutionProgress(ctx, tx.TransactionReceiptID, submission.SubmissionReceiptID, EscrowExecutionStatusFunded, escrowRef, EventEscrowExecutionFunded, "funded")
	require.NoError(t, err)
	return store, submission, funded, escrowRef
}

func receiptsEvidenceBranchFundedDisputeReady(t *testing.T, ctx context.Context, transactionID string) (*Store, SubmissionReceipt, TransactionReceipt, string) {
	t.Helper()
	store, submission, tx, escrowRef := receiptsEvidenceBranchFunded(t, ctx, transactionID)
	_, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionReviewNeeded, SettlementProgressionReasonCodeReject, "review needed", "")
	require.NoError(t, err)
	disputeReady, err := store.ApplySettlementProgression(ctx, tx.TransactionReceiptID, SettlementProgressionDisputeReady, SettlementProgressionReasonCodeEscalate, "dispute ready", "")
	require.NoError(t, err)
	return store, submission, disputeReady, escrowRef
}
