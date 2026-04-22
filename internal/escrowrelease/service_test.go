package escrowrelease

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
	runtime := &fakeReleaseRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "missing"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonMissingReceipt)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "missing", result.TransactionReceiptID)
	require.Equal(t, 0, runtime.calls)
}

func TestServiceExecute_DeniesWhenCurrentSubmissionIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			PriceContext:                "quote:1.00-usdc",
		},
	}
	runtime := &fakeReleaseRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNoCurrentSubmission)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_DeniesWhenEscrowIsNotFunded(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusCreated,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeReleaseRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonEscrowNotFunded)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_DeniesWhenProgressionIsNotApprovedForSettlement(t *testing.T) {
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
	runtime := &fakeReleaseRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNotApprovedForSettlement)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_DeniesWhenAmountCannotBeResolved(t *testing.T) {
	t.Parallel()

	cases := []string{"quote:abc-usdc", "quote:0-usdc", "quote:-1.00-usdc"}
	for _, priceContext := range cases {
		priceContext := priceContext
		t.Run(priceContext, func(t *testing.T) {
			store := &fakeReceiptStore{
				transaction: receipts.TransactionReceipt{
					TransactionReceiptID:        "tx-1",
					CurrentSubmissionReceiptID:  "sub-1",
					EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
					SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
					PriceContext:                priceContext,
				},
				submission: receipts.SubmissionReceipt{
					SubmissionReceiptID:  "sub-1",
					TransactionReceiptID: "tx-1",
				},
			}
			runtime := &fakeReleaseRuntime{}
			svc := NewService(store, runtime)

			result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

			require.Error(t, err)
			assertExecutionError(t, err, FailureKindDenied, DenyReasonAmountUnresolved)
			require.Equal(t, StatusDenied, result.Status)
		})
	}
}

func TestServiceExecute_ExecutesReleaseAndReturnsSettledTarget(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			EscrowReference:             "escrow-123",
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		markSettledResult: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			SettlementProgressionStatus: receipts.SettlementProgressionSettled,
		},
	}
	runtime := &fakeReleaseRuntime{result: ReleaseResult{Reference: "release-tx-123"}}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.NoError(t, err)
	require.Equal(t, StatusSettledTarget, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionSettled, result.SettlementProgressionStatus)
	require.Equal(t, "1.00", result.ResolvedAmount)
	require.Equal(t, "release-tx-123", result.RuntimeReference)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, ReleaseRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		EscrowReference:      "escrow-123",
		Amount:               "1.00",
	}, runtime.lastRequest)
	require.Equal(t, 1, store.markSettledCalls)
	require.Equal(t, receipts.SettlementCloseoutRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		ResolvedAmount:       "1.00",
		RuntimeReference:     "release-tx-123",
	}, store.lastCloseout)
}

func TestServiceExecute_RuntimeFailureReturnsFailureShape(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			EscrowReference:             "escrow-123",
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeReleaseRuntime{err: errors.New("escrow release failed")}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindExecutionFailed, "")
	require.Equal(t, StatusFailed, result.Status)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.SettlementProgressionStatus)
	require.Equal(t, "1.00", result.ResolvedAmount)
	require.NotNil(t, result.Failure)
	require.Equal(t, FailureKindExecutionFailed, result.Failure.Kind)
	require.Equal(t, 1, store.recordFailureCalls)
	require.Equal(t, receipts.SettlementFailureRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		ResolvedAmount:       "1.00",
		Reason:               "escrow release failed",
	}, store.lastFailure)
}

func TestServiceExecute_RuntimeFailureStillReturnsFailureShapeWhenFailureRecordingFails(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			EscrowExecutionStatus:       receipts.EscrowExecutionStatusFunded,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			EscrowReference:             "escrow-123",
			PriceContext:                "quote:1.00-usdc",
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		recordFailureErr: errors.New("trail unavailable"),
	}
	runtime := &fakeReleaseRuntime{err: errors.New("escrow release failed")}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	require.Equal(t, StatusFailed, result.Status)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.SettlementProgressionStatus)
	require.Equal(t, "1.00", result.ResolvedAmount)
	require.NotNil(t, result.Failure)
	require.Equal(t, FailureKindExecutionFailed, result.Failure.Kind)
	require.ErrorContains(t, err, "record escrow release failure")
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
	markSettledErr     error
	recordFailureErr   error
	markSettledResult  receipts.TransactionReceipt
	markSettledCalls   int
	recordFailureCalls int
	lastCloseout       receipts.SettlementCloseoutRequest
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
	return f.submission, nil, nil
}

func (f *fakeReceiptStore) MarkSettlementSettled(_ context.Context, req receipts.SettlementCloseoutRequest) (receipts.TransactionReceipt, error) {
	f.markSettledCalls++
	f.lastCloseout = req
	if f.markSettledErr != nil {
		return receipts.TransactionReceipt{}, f.markSettledErr
	}
	if f.markSettledResult.TransactionReceiptID == "" {
		tx := f.transaction
		tx.SettlementProgressionStatus = receipts.SettlementProgressionSettled
		return tx, nil
	}
	return f.markSettledResult, nil
}

func (f *fakeReceiptStore) RecordSettlementFailure(_ context.Context, req receipts.SettlementFailureRequest) error {
	f.recordFailureCalls++
	f.lastFailure = req
	return f.recordFailureErr
}

type fakeReleaseRuntime struct {
	result      ReleaseResult
	err         error
	calls       int
	lastRequest ReleaseRequest
}

func (f *fakeReleaseRuntime) Release(_ context.Context, req ReleaseRequest) (ReleaseResult, error) {
	f.calls++
	f.lastRequest = req
	if f.err != nil {
		return ReleaseResult{}, f.err
	}
	return f.result, nil
}
