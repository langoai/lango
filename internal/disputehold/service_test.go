package disputehold

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
	require.Equal(t, receipts.DisputeLifecycleHoldActive, result.DisputeLifecycleStatus)
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

func TestServiceExecute_RecordHoldSuccessFailureReturnsWrappedError(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			EscrowReference:             "escrow-123",
		},
		submission:       receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1", TransactionReceiptID: "tx-1"},
		recordSuccessErr: errors.New("write failed"),
	}
	runtime := &fakeHoldRuntime{result: HoldResult{Reference: "hold-123"}}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	require.ErrorContains(t, err, "record dispute hold success")
	require.Equal(t, Result{}, result)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 1, store.recordSuccessCalls)
	require.Equal(t, 0, store.recordFailureCalls)
}

func TestServiceExecute_SerializesConcurrentHoldExecutionPerTransaction(t *testing.T) {
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
	runtime := &blockingHoldRuntime{result: HoldResult{Reference: "hold-123"}}
	svc := NewService(store, runtime)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _ = svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})
		}()
	}

	close(start)
	wg.Wait()

	require.Equal(t, int32(2), atomic.LoadInt32(&runtime.calls))
	require.Equal(t, int32(1), atomic.LoadInt32(&runtime.maxConcurrent))
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

type blockingHoldRuntime struct {
	result        HoldResult
	calls         int32
	active        int32
	maxConcurrent int32
}

func (b *blockingHoldRuntime) Hold(_ context.Context, _ EscrowHoldRequest) (HoldResult, error) {
	atomic.AddInt32(&b.calls, 1)
	active := atomic.AddInt32(&b.active, 1)
	defer atomic.AddInt32(&b.active, -1)

	for {
		current := atomic.LoadInt32(&b.maxConcurrent)
		if active <= current || atomic.CompareAndSwapInt32(&b.maxConcurrent, current, active) {
			break
		}
	}

	time.Sleep(50 * time.Millisecond)
	return b.result, nil
}
