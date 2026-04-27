package escrowrefund

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
	runtime := &fakeRefundRuntime{}
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
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			PriceContext:                "quote:1.00-usdc",
		},
	}
	runtime := &fakeRefundRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNoCurrentSubmission)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenEscrowIsNotFunded(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusCreated,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeRefundRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonEscrowNotFunded)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenProgressionIsNotReviewNeeded(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeRefundRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNotReviewNeeded)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenAdjudicationIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	svc := NewService(store, &fakeRefundRuntime{})

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonAdjudicationMissing)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_DeniesWhenAdjudicationMismatchesOrOppositeEvidenceExists(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		adjudication receipts.EscrowAdjudicationDecision
		events       []receipts.ReceiptEvent
	}{
		{
			name:         "mismatch",
			adjudication: receipts.EscrowAdjudicationRelease,
		},
		{
			name:         "opposite evidence",
			adjudication: receipts.EscrowAdjudicationRefund,
			events: []receipts.ReceiptEvent{
				{Source: "escrow_release", Subtype: "settled", Type: receipts.EventSettlementUpdated},
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeReceiptStore{
				transaction: receipts.TransactionReceipt{
					TransactionReceiptID:        "tx-1",
					CurrentSubmissionReceiptID:  "sub-1",
					EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
					SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
					EscrowAdjudication:          tt.adjudication,
					PriceContext:                "quote:1.00-usdc",
				},
				submission: receipts.SubmissionReceipt{
					SubmissionReceiptID:  "sub-1",
					TransactionReceiptID: "tx-1",
				},
				events: tt.events,
			}
			svc := NewService(store, &fakeRefundRuntime{})

			result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

			require.Error(t, err)
			assertExecutionError(t, err, FailureKindDenied, DenyReasonAdjudicationMismatch)
			require.Equal(t, StatusDenied, result.Status)
		})
	}
}

func TestServiceExecute_DeniesWhenAmountCannotBeResolved(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			EscrowAdjudication:          receipts.EscrowAdjudicationRefund,
			PriceContext:                "quote:abc-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeRefundRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonAmountUnresolved)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_ExecutesRuntimeAndReturnsRefundExecutedShape(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			EscrowAdjudication:          receipts.EscrowAdjudicationRefund,
			EscrowReference:             "escrow-123",
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeRefundRuntime{result: RefundResult{Reference: "refund-tx-123"}}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.NoError(t, err)
	require.Equal(t, StatusRefundExecuted, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.SettlementProgressionStatus)
	require.Equal(t, "1.00", result.ResolvedAmount)
	require.Equal(t, "refund-tx-123", result.RuntimeReference)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 1, store.recordSuccessCalls)
	require.Equal(t, 0, store.recordFailureCalls)
	require.Equal(t, receipts.EscrowRefundEvidenceRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		RuntimeReference:     "refund-tx-123",
	}, store.lastSuccess)
	require.Equal(t, RefundRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		EscrowReference:      "escrow-123",
		Amount:               "1.00",
	}, runtime.lastRequest)
}

func TestServiceExecute_RuntimeFailureReturnsFailureShape(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			EscrowAdjudication:          receipts.EscrowAdjudicationRefund,
			EscrowReference:             "escrow-123",
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeRefundRuntime{err: errors.New("escrow refund failed")}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindExecutionFailed, "")
	require.Equal(t, StatusFailed, result.Status)
	require.Equal(t, receipts.SettlementProgressionReviewNeeded, result.SettlementProgressionStatus)
	require.Equal(t, "1.00", result.ResolvedAmount)
	require.NotNil(t, result.Failure)
	require.Equal(t, FailureKindExecutionFailed, result.Failure.Kind)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 0, store.recordSuccessCalls)
	require.Equal(t, 1, store.recordFailureCalls)
	require.Equal(t, receipts.SettlementFailureRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		ResolvedAmount:       "1.00",
		Reason:               "escrow refund failed",
	}, store.lastFailure)
}

func TestServiceExecute_RecordRefundSuccessFailureReturnsWrappedError(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			EscrowAdjudication:          receipts.EscrowAdjudicationRefund,
			EscrowReference:             "escrow-123",
			PriceContext:                "quote:1.00-usdc",
		},
		submission:       receipts.SubmissionReceipt{SubmissionReceiptID: "sub-1", TransactionReceiptID: "tx-1"},
		recordSuccessErr: errors.New("write failed"),
	}
	runtime := &fakeRefundRuntime{result: RefundResult{Reference: "refund-tx-123"}}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	require.ErrorContains(t, err, "record escrow refund success")
	require.Equal(t, Result{}, result)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 1, store.recordSuccessCalls)
	require.Equal(t, 0, store.recordFailureCalls)
}

func TestServiceExecute_SerializesConcurrentRefundExecutionPerTransaction(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			EscrowAdjudication:          receipts.EscrowAdjudicationRefund,
			EscrowReference:             "escrow-123",
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &blockingRefundRuntime{result: RefundResult{Reference: "refund-tx-123"}}
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
	events             []receipts.ReceiptEvent
	getTransactionErr  error
	getSubmissionErr   error
	recordSuccessErr   error
	recordFailureErr   error
	recordSuccessCalls int
	recordFailureCalls int
	lastSuccess        receipts.EscrowRefundEvidenceRequest
	lastFailure        receipts.SettlementFailureRequest
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

func (f *fakeReceiptStore) RecordEscrowRefundSuccess(_ context.Context, req receipts.EscrowRefundEvidenceRequest) error {
	f.recordSuccessCalls++
	f.lastSuccess = req
	return f.recordSuccessErr
}

func (f *fakeReceiptStore) RecordEscrowRefundFailure(_ context.Context, req receipts.SettlementFailureRequest) error {
	f.recordFailureCalls++
	f.lastFailure = req
	return f.recordFailureErr
}

type fakeRefundRuntime struct {
	result      RefundResult
	err         error
	calls       int
	lastRequest RefundRequest
}

func (f *fakeRefundRuntime) Refund(_ context.Context, req RefundRequest) (RefundResult, error) {
	f.calls++
	f.lastRequest = req
	if f.err != nil {
		return RefundResult{}, f.err
	}
	return f.result, nil
}

type blockingRefundRuntime struct {
	result        RefundResult
	calls         int32
	active        int32
	maxConcurrent int32
}

func (b *blockingRefundRuntime) Refund(_ context.Context, _ RefundRequest) (RefundResult, error) {
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
