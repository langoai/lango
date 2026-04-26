package escrowadjudication

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/receipts"
)

func TestServiceAdjudicate_DeniesMissingTransactionReceipt(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{getTransactionErr: receipts.ErrTransactionReceiptNotFound}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "missing",
		Outcome:              OutcomeRelease,
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonMissingReceipt)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "missing", result.TransactionReceiptID)
	require.Equal(t, receipts.SettlementProgressionPending, result.SettlementProgressionStatus)
}

func TestServiceAdjudicate_DeniesWhenCurrentSubmissionIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
	}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "tx-1",
		Outcome:              OutcomeRelease,
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNoCurrentSubmission)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, receipts.SettlementProgressionDisputeReady, result.SettlementProgressionStatus)
}

func TestServiceAdjudicate_DeniesWhenEscrowIsNotFunded(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusCreated,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "tx-1",
		Outcome:              OutcomeRelease,
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonEscrowNotFunded)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceAdjudicate_DeniesWhenProgressionIsNotDisputeReady(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			EscrowReference:             "escrow-123",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "tx-1",
		Outcome:              OutcomeRelease,
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNotDisputeReady)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceAdjudicate_DeniesWhenHoldEvidenceIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		events: []receipts.ReceiptEvent{
			{Source: "settlement_progression", Subtype: "dispute-ready", Type: receipts.EventDisputed},
		},
	}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "tx-1",
		Outcome:              OutcomeRelease,
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonHoldEvidenceMissing)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, 1, store.recordFailureCalls)
}

func TestServiceAdjudicate_DeniesInvalidOutcome(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		events: []receipts.ReceiptEvent{
			{Source: "dispute_hold", Subtype: "held", Type: receipts.EventSettlementUpdated},
		},
	}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "tx-1",
		Outcome:              Outcome("other"),
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonInvalidOutcome)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, 1, store.recordFailureCalls)
}

func TestServiceAdjudicate_AppliesReleaseDecision(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		events: []receipts.ReceiptEvent{
			{Source: "dispute_hold", Subtype: "held", Type: receipts.EventSettlementUpdated},
		},
		applyResult: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			DisputeLifecycleStatus:      receipts.DisputeLifecycleHoldActive,
			EscrowReference:             "escrow-123",
			EscrowAdjudication:          receipts.EscrowAdjudicationRelease,
		},
	}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "tx-1",
		Outcome:              OutcomeRelease,
		Reason:               "fulfilled after review",
	})

	require.NoError(t, err)
	require.Equal(t, StatusAdjudicated, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.SettlementProgressionStatus)
	require.Equal(t, receipts.DisputeLifecycleHoldActive, result.DisputeLifecycleStatus)
	require.Equal(t, "escrow-123", result.EscrowReference)
	require.Equal(t, OutcomeRelease, result.Outcome)
	require.Equal(t, 1, store.applyCalls)
	require.Equal(t, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		EscrowReference:      "escrow-123",
		Outcome:              receipts.EscrowAdjudicationRelease,
		Reason:               "fulfilled after review",
	}, store.lastApply)
}

func TestServiceAdjudicate_ApplyFailureReturnsFailureShape(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		events: []receipts.ReceiptEvent{
			{Source: "dispute_hold", Subtype: "held", Type: receipts.EventSettlementUpdated},
		},
		applyErr: errors.New("write failed"),
	}
	svc := NewService(store)

	result, err := svc.Adjudicate(context.Background(), Request{
		TransactionReceiptID: "tx-1",
		Outcome:              OutcomeRefund,
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindApplyFailed, "")
	require.Equal(t, StatusFailed, result.Status)
	require.Equal(t, OutcomeRefund, result.Outcome)
	require.NotNil(t, result.Failure)
	require.Equal(t, 1, store.recordFailureCalls)
}

func assertExecutionError(t *testing.T, err error, wantKind FailureKind, wantReason DenyReason) {
	t.Helper()
	var executionErr *ExecutionError
	require.ErrorAs(t, err, &executionErr)
	require.Equal(t, wantKind, executionErr.Kind)
	require.Equal(t, wantReason, executionErr.DenyReason)
}

type fakeReceiptStore struct {
	transaction        receipts.TransactionReceipt
	submission         receipts.SubmissionReceipt
	events             []receipts.ReceiptEvent
	getTransactionErr  error
	getSubmissionErr   error
	applyErr           error
	recordFailureErr   error
	applyCalls         int
	recordFailureCalls int
	lastApply          receipts.EscrowAdjudicationRequest
	lastFailure        receipts.EscrowAdjudicationFailureRequest
	applyResult        receipts.TransactionReceipt
}

func (f *fakeReceiptStore) GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error) {
	if f.getTransactionErr != nil {
		return receipts.TransactionReceipt{}, f.getTransactionErr
	}
	return f.transaction, nil
}

func (f *fakeReceiptStore) GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error) {
	if f.getSubmissionErr != nil {
		return receipts.SubmissionReceipt{}, nil, f.getSubmissionErr
	}
	return f.submission, f.events, nil
}

func (f *fakeReceiptStore) ApplyEscrowAdjudication(_ context.Context, req receipts.EscrowAdjudicationRequest) (receipts.TransactionReceipt, error) {
	f.applyCalls++
	f.lastApply = req
	if f.applyErr != nil {
		return receipts.TransactionReceipt{}, f.applyErr
	}
	return f.applyResult, nil
}

func (f *fakeReceiptStore) RecordEscrowAdjudicationFailure(_ context.Context, req receipts.EscrowAdjudicationFailureRequest) error {
	f.recordFailureCalls++
	f.lastFailure = req
	return f.recordFailureErr
}
