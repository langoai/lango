package settlementexecution

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/receipts"
)

func TestServiceExecute_DeniesMissingTransactionReceipt(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		getTransactionErr: receipts.ErrTransactionReceiptNotFound,
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{
		TransactionReceiptID: "tx-missing",
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonMissingReceipt)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-missing", result.TransactionReceiptID)
	require.Empty(t, result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionPending, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
	require.Equal(t, 0, store.markSettledCalls)
	require.Equal(t, 0, store.recordFailureCalls)
}

func TestServiceExecute_DeniesWhenCurrentSubmissionIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			PriceContext:                "quote:0.50-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{
		TransactionReceiptID: "tx-1",
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNoCurrentSubmission)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Empty(t, result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
	require.Equal(t, 0, store.markSettledCalls)
	require.Equal(t, 0, store.recordFailureCalls)
}

func TestServiceExecute_DeniesWhenProgressionIsNotApprovedForSettlement(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			PriceContext:                "quote:0.50-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionPending,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{
		TransactionReceiptID: "tx-1",
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNotApprovedForSettlement)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionPending, result.SettlementProgressionStatus)
	require.Equal(t, 0, runtime.calls)
	require.Equal(t, 0, store.markSettledCalls)
	require.Equal(t, 0, store.recordFailureCalls)
}

func TestServiceExecute_DeniesWhenAmountCannotBeResolved(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			PriceContext:                "quote:abc-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{
		TransactionReceiptID: "tx-1",
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonAmountUnresolved)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Empty(t, result.ResolvedAmount)
	require.Equal(t, 0, runtime.calls)
	require.Equal(t, 0, store.markSettledCalls)
	require.Equal(t, 0, store.recordFailureCalls)
}

func TestServiceExecute_ExecutesRuntimeAndReportsSettledTarget(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			Counterparty:                "did:lango:peer",
			PriceContext:                "quote:0.50-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
		markSettledResult: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			PriceContext:                "quote:0.50-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionSettled,
		},
	}
	runtime := &fakeDirectPaymentRuntime{
		result: DirectPaymentResult{
			Reference: "settlement-tx-123",
		},
	}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{
		TransactionReceiptID: "tx-1",
	})

	require.NoError(t, err)
	require.Equal(t, StatusSettledTarget, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionSettled, result.SettlementProgressionStatus)
	require.Equal(t, "0.50", result.ResolvedAmount)
	require.Equal(t, "settlement-tx-123", result.RuntimeReference)
	require.Nil(t, result.Failure)

	require.Equal(t, 1, runtime.calls)
	require.Equal(t, DirectPaymentRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		Counterparty:         "did:lango:peer",
		Amount:               "0.50",
	}, runtime.lastRequest)

	require.Equal(t, 1, store.markSettledCalls)
	require.Equal(t, CloseoutRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		ResolvedAmount:       "0.50",
		RuntimeReference:     "settlement-tx-123",
	}, store.lastCloseout)
	require.Equal(t, 0, store.recordFailureCalls)
}

func TestServiceExecute_RuntimeFailureReturnsFailureShape(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			Counterparty:                "did:lango:peer",
			PriceContext:                "quote:0.50-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{
		err: errors.New("rpc timeout"),
	}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{
		TransactionReceiptID: "tx-1",
	})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindExecutionFailed, "")
	require.Equal(t, StatusFailed, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.SettlementProgressionStatus)
	require.Equal(t, "0.50", result.ResolvedAmount)
	require.NotNil(t, result.Failure)
	require.Equal(t, FailureKindExecutionFailed, result.Failure.Kind)
	require.Equal(t, "rpc timeout", result.Failure.Message)
	require.Empty(t, result.RuntimeReference)

	require.Equal(t, 1, runtime.calls)
	require.Equal(t, 0, store.markSettledCalls)
	require.Equal(t, 1, store.recordFailureCalls)
	require.Equal(t, FailureRecordRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		ResolvedAmount:       "0.50",
		Reason:               "rpc timeout",
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
	markSettledErr     error
	recordFailureErr   error
	markSettledResult  receipts.TransactionReceipt
	calls              int
	markSettledCalls   int
	recordFailureCalls int
	lastCloseout       CloseoutRequest
	lastFailure        FailureRecordRequest
}

func (f *fakeReceiptStore) GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error) {
	f.calls++
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

func (f *fakeReceiptStore) MarkSettlementSettled(_ context.Context, req CloseoutRequest) (receipts.TransactionReceipt, error) {
	f.markSettledCalls++
	f.lastCloseout = req
	if f.markSettledErr != nil {
		return receipts.TransactionReceipt{}, f.markSettledErr
	}
	if f.markSettledResult.TransactionReceiptID == "" {
		return f.transaction, nil
	}
	return f.markSettledResult, nil
}

func (f *fakeReceiptStore) RecordSettlementFailure(_ context.Context, req FailureRecordRequest) error {
	f.recordFailureCalls++
	f.lastFailure = req
	return f.recordFailureErr
}

type fakeDirectPaymentRuntime struct {
	result      DirectPaymentResult
	err         error
	calls       int
	lastRequest DirectPaymentRequest
}

func (f *fakeDirectPaymentRuntime) ExecuteSettlement(_ context.Context, req DirectPaymentRequest) (DirectPaymentResult, error) {
	f.calls++
	f.lastRequest = req
	if f.err != nil {
		return DirectPaymentResult{}, f.err
	}
	return f.result, nil
}
