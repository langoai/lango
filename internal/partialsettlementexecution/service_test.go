package partialsettlementexecution

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
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "missing"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonMissingReceipt)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "missing", result.TransactionReceiptID)
}

func TestServiceExecute_DeniesWhenCurrentSubmissionIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			PriceContext:                "quote:1.00-usdc",
			PartialSettlementHint:       "settle:0.40-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNoCurrentSubmission)
	require.Equal(t, StatusDenied, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
}

func TestServiceExecute_DeniesWhenProgressionIsNotApprovedForSettlement(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			PriceContext:                "quote:1.00-usdc",
			PartialSettlementHint:       "settle:0.40-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonNotApprovedForSettlement)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_DeniesWhenPartialHintIsMissing(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			PriceContext:                "quote:1.00-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonPartialHintMissing)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_DeniesWhenPartialHintIsInvalid(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			PriceContext:                "quote:1.00-usdc",
			PartialSettlementHint:       "settle:forty%",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonPartialHintInvalid)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_DeniesWhenPartialHintAmountIsNotPositive(t *testing.T) {
	t.Parallel()

	cases := []string{"settle:0-usdc", "settle:-0.10-usdc"}
	for _, hint := range cases {
		hint := hint
		t.Run(hint, func(t *testing.T) {
			store := &fakeReceiptStore{
				transaction: receipts.TransactionReceipt{
					TransactionReceiptID:        "tx-1",
					CurrentSubmissionReceiptID:  "sub-1",
					PriceContext:                "quote:1.00-usdc",
					PartialSettlementHint:       hint,
					SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
				},
				submission: receipts.SubmissionReceipt{
					SubmissionReceiptID:  "sub-1",
					TransactionReceiptID: "tx-1",
				},
			}
			runtime := &fakeDirectPaymentRuntime{}
			svc := NewService(store, runtime)

			result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

			require.Error(t, err)
			assertExecutionError(t, err, FailureKindDenied, DenyReasonPartialHintInvalid)
			require.Equal(t, StatusDenied, result.Status)
		})
	}
}

func TestServiceExecute_DeniesWhenAlreadyPartiallySettled(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			PriceContext:                "quote:1.00-usdc",
			PartialSettlementHint:       "settle:0.40-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionPartiallySettled,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindDenied, DenyReasonAlreadyPartiallySettled)
	require.Equal(t, StatusDenied, result.Status)
}

func TestServiceExecute_ExecutesRuntimeAndReturnsPartialShape(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			Counterparty:                "did:lango:peer",
			PriceContext:                "quote:1.00-usdc",
			PartialSettlementHint:       "settle:0.40-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{result: DirectPaymentResult{Reference: "partial-tx-123"}}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.NoError(t, err)
	require.Equal(t, StatusPartiallySettledTarget, result.Status)
	require.Equal(t, "tx-1", result.TransactionReceiptID)
	require.Equal(t, "sub-1", result.SubmissionReceiptID)
	require.Equal(t, receipts.SettlementProgressionPartiallySettled, result.SettlementProgressionStatus)
	require.Equal(t, "0.40", result.ExecutedAmount)
	require.Equal(t, "0.60", result.RemainingAmount)
	require.Equal(t, "partial-tx-123", result.RuntimeReference)
	require.Equal(t, 1, runtime.calls)
	require.Equal(t, DirectPaymentRequest{
		TransactionReceiptID: "tx-1",
		SubmissionReceiptID:  "sub-1",
		Counterparty:         "did:lango:peer",
		Amount:               "0.40",
	}, runtime.lastRequest)
}

func TestServiceExecute_RuntimeFailureReturnsFailureShape(t *testing.T) {
	t.Parallel()

	store := &fakeReceiptStore{
		transaction: receipts.TransactionReceipt{
			TransactionReceiptID:        "tx-1",
			CurrentSubmissionReceiptID:  "sub-1",
			Counterparty:                "did:lango:peer",
			PriceContext:                "quote:1.00-usdc",
			PartialSettlementHint:       "settle:0.40-usdc",
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
		},
		submission: receipts.SubmissionReceipt{
			SubmissionReceiptID:  "sub-1",
			TransactionReceiptID: "tx-1",
		},
	}
	runtime := &fakeDirectPaymentRuntime{err: errors.New("rpc timeout")}
	svc := NewService(store, runtime)

	result, err := svc.Execute(context.Background(), Request{TransactionReceiptID: "tx-1"})

	require.Error(t, err)
	assertExecutionError(t, err, FailureKindExecutionFailed, "")
	require.Equal(t, StatusFailed, result.Status)
	require.Equal(t, receipts.SettlementProgressionApprovedForSettlement, result.SettlementProgressionStatus)
	require.Equal(t, "0.40", result.ExecutedAmount)
	require.Equal(t, "0.60", result.RemainingAmount)
	require.NotNil(t, result.Failure)
	require.Equal(t, FailureKindExecutionFailed, result.Failure.Kind)
}

func assertExecutionError(t *testing.T, err error, wantKind FailureKind, wantReason DenyReason) {
	t.Helper()

	var executionErr *ExecutionError
	require.ErrorAs(t, err, &executionErr)
	require.Equal(t, wantKind, executionErr.Kind)
	require.Equal(t, wantReason, executionErr.DenyReason)
}

type fakeReceiptStore struct {
	transaction       receipts.TransactionReceipt
	submission        receipts.SubmissionReceipt
	getTransactionErr error
	getSubmissionErr  error
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
