package settlementexecution

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/finance"
	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error)
	MarkSettlementSettled(context.Context, closeoutRequest) (receipts.TransactionReceipt, error)
	RecordSettlementFailure(context.Context, failureRequest) error
}

type directPaymentRuntime interface {
	ExecuteSettlement(context.Context, DirectPaymentRequest) (DirectPaymentResult, error)
}

type Service struct {
	store   receiptStore
	runtime directPaymentRuntime
}

func NewService(store receiptStore, runtime directPaymentRuntime) *Service {
	return &Service{store: store, runtime: runtime}
}

func (s *Service) Execute(ctx context.Context, req ExecuteRequest) (Result, error) {
	transactionReceiptID := strings.TrimSpace(req.TransactionReceiptID)
	if transactionReceiptID == "" {
		return deniedResult("", "", receipts.SettlementProgressionPending, DenyReasonMissingReceipt)
	}
	if s == nil || s.store == nil {
		return Result{}, fmt.Errorf("receipt store is required")
	}
	if s.runtime == nil {
		return Result{}, fmt.Errorf("direct payment runtime is required")
	}

	transaction, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrTransactionReceiptNotFound) {
			return deniedResult(transactionReceiptID, "", receipts.SettlementProgressionPending, DenyReasonMissingReceipt)
		}
		return Result{}, err
	}

	submissionReceiptID := strings.TrimSpace(transaction.CurrentSubmissionReceiptID)
	if submissionReceiptID == "" {
		return deniedResult(transaction.TransactionReceiptID, "", transaction.SettlementProgressionStatus, DenyReasonNoCurrentSubmission)
	}

	submission, _, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
			return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonNoCurrentSubmission)
		}
		return Result{}, err
	}
	if submission.TransactionReceiptID != transaction.TransactionReceiptID {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonNoCurrentSubmission)
	}

	if transaction.SettlementProgressionStatus != receipts.SettlementProgressionApprovedForSettlement {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonNotApprovedForSettlement)
	}

	resolvedAmount, err := resolveAmountFromPriceContext(transaction.PriceContext)
	if err != nil {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonAmountUnresolved)
	}

	runtimeResult, err := s.runtime.ExecuteSettlement(ctx, DirectPaymentRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		Counterparty:         transaction.Counterparty,
		Amount:               resolvedAmount,
	})
	if err != nil {
		result := Result{
			Status:                      ResultStatusFailed,
			TransactionReceiptID:        transaction.TransactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			ResolvedAmount:              resolvedAmount,
			Failure: &Failure{
				Kind:    FailureKindExecutionFailed,
				Message: err.Error(),
			},
		}
		failure := failureRequest{
			TransactionReceiptID: transaction.TransactionReceiptID,
			SubmissionReceiptID:  submissionReceiptID,
			ResolvedAmount:       resolvedAmount,
			Reason:               err.Error(),
		}
		if recordErr := s.store.RecordSettlementFailure(ctx, failure); recordErr != nil {
			return result, fmt.Errorf("record settlement failure: %w", recordErr)
		}
		return result, &ExecutionError{Kind: FailureKindExecutionFailed, Message: err.Error(), Err: err}
	}

	updated, err := s.store.MarkSettlementSettled(ctx, closeoutRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		ResolvedAmount:       resolvedAmount,
		RuntimeReference:     runtimeResult.Reference,
	})
	if err != nil {
		return Result{}, err
	}

	return Result{
		Status:                      ResultStatusSettledTarget,
		TransactionReceiptID:        updated.TransactionReceiptID,
		SubmissionReceiptID:         submissionReceiptID,
		SettlementProgressionStatus: updated.SettlementProgressionStatus,
		ResolvedAmount:              resolvedAmount,
		RuntimeReference:            runtimeResult.Reference,
	}, nil
}

func deniedResult(transactionReceiptID, submissionReceiptID string, status receipts.SettlementProgressionStatus, reason DenyReason) (Result, error) {
	return Result{
			Status:                      ResultStatusDenied,
			TransactionReceiptID:        transactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: status,
			Failure: &Failure{
				Kind:       FailureKindDenied,
				DenyReason: reason,
				Message:    string(reason),
			},
		},
		&ExecutionError{Kind: FailureKindDenied, DenyReason: reason, Message: string(reason)}
}

func resolveAmountFromPriceContext(priceContext string) (string, error) {
	trimmed := strings.TrimSpace(priceContext)
	if !strings.HasPrefix(trimmed, "quote:") || !strings.HasSuffix(trimmed, "-usdc") {
		return "", fmt.Errorf("unsupported price context %q", priceContext)
	}
	amount := strings.TrimSuffix(strings.TrimPrefix(trimmed, "quote:"), "-usdc")
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return "", fmt.Errorf("missing amount in price context %q", priceContext)
	}
	parsed, err := finance.ParseUSDC(amount)
	if err != nil {
		return "", err
	}
	return finance.FormatUSDC(parsed), nil
}
