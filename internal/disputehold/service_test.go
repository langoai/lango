package disputehold

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/receipts"
)

func TestServiceExecute_DeniesMissingTransactionReceipt(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{getTransactionErr: receipts.ErrTransactionReceiptNotFound}
	runtime := &fakeHoldRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "missing"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonMissingReceipt)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "missing", result.TransactionReceiptID)
	require.Equal(t, receipts.SettlementProgressionPending, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenCurrentSubmissionIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
	}
	runtime := &fakeHoldRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNoCurrentSubmission)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, receipts.SettlementProgressionDisputeReady, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenEscrowIsNotFunded(t *testing.T) {
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
	runtime := &fakeHoldRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonEscrowNotFunded)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, receipts.SettlementProgressionDisputeReady, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenProgressionIsNotDisputeReady(t *testing.T) {
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
	runtime := &fakeHoldRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNotDisputeReady)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenEscrowReferenceIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeHoldRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonEscrowReferenceMissing)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, receipts.SettlementProgressionDisputeReady, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_RecordsHoldEvidenceWhileKeepingStateUnchanged(t *testing.T) {
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
	}
	runtime := &fakeHoldRuntime{result: HoldResult{Reference: "hold-123"}}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.NoError(t, err)
	require.Equal(t, StatusHoldApplied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionDisputeReady, result.SettlementProgressionStatus)
	require.Equal(t, "escrow-123", result.EscrowReference)
	require.Equal(t, "hold-123", result.RuntimeReference)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 1, store.recordSuccessCalls)
	require.Equal(t, 0, store.recordFailureCalls)
	require.Equal(t, EscrowHoldRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		EscrowReference:      "escrow-123",
	}, runtime.lastRequest)
	require.Equal(t, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		EscrowReference:      "escrow-123",
		RuntimeReference:     "hold-123",
	}, store.lastSuccess)
}

func TestServiceExecute_RuntimeFailureReturnsFailureShapeWhileKeepingStateUnchanged(t *testing.T) {
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
	}
	runtime := &fakeHoldRuntime{err: errors.New("hold failed")}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindExecutionFailed, "")
	require.Equal(t, StatusFailed, result.Status)
	require.Equal(t, receipts.SettlementProgressionDisputeReady, result.SettlementProgressionStatus)
	require.Equal(t, "escrow-123", result.EscrowReference)
	require.NotNil(t, result.Failure)
	require.Equal(t, FailureKindExecutionFailed, result.Failure.Kind)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 0, store.recordSuccessCalls)
	require.Equal(t, 1, store.recordFailureCalls)
	require.Equal(t, receipts.EscrowDisputeHoldFailureRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		EscrowReference:      "escrow-123",
		Reason:               "hold failed",
	}, store.lastFailure)
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
	getTransactionErr  error
	getSubmissionErr   error
	recordSuccessErr   error
	recordFailureErr   error
	recordSuccessCalls int
	recordFailureCalls int
	lastSuccess        receipts.EscrowDisputeHoldEvidenceRequest
	lastFailure        receipts.EscrowDisputeHoldFailureRequest
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
	return f.submission, nil, nil
}

func (f *fakeReceiptStore) RecordEscrowDisputeHoldSuccess(_ context.Context, req receipts.EscrowDisputeHoldEvidenceRequest) error {
	f.recordSuccessCalls++
	f.lastSuccess = req
	return f.recordSuccessErr
}

func (f *fakeReceiptStore) RecordEscrowDisputeHoldFailure(_ context.Context, req receipts.EscrowDisputeHoldFailureRequest) error {
	f.recordFailureCalls++
	f.lastFailure = req
	return f.recordFailureErr
}

type fakeHoldRuntime struct {
	result      HoldResult
	err         error
	calls       int
	lastRequest EscrowHoldRequest
}

func (f *fakeHoldRuntime) Hold(_ context.Context, req EscrowHoldRequest) (HoldResult, error) {
	f.calls++
	f.lastRequest = req
	if f.err != nil {
		return HoldResult{}, f.err
	}
	return f.result, nil
}
