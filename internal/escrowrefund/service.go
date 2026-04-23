package escrowrefund

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
	RecordEscrowRefundSuccess(context.Context, receipts.EscrowRefundEvidenceRequest) error
	RecordEscrowRefundFailure(context.Context, receipts.SettlementFailureRequest) error
}

type refundRuntime interface {
	Refund(context.Context, RefundRequest) (RefundResult, error)
}

type Service struct {
	store   receiptStore
	runtime refundRuntime
}

func NewService(store receiptStore, runtime refundRuntime) *Service {
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
		return Result{}, fmt.Errorf("escrow refund runtime is required")
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

	if transaction.EscrowExecutionStatus != receipts.EscrowExecutionStatusFunded {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonEscrowNotFunded)
	}

	if transaction.SettlementProgressionStatus != receipts.SettlementProgressionReviewNeeded {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonNotReviewNeeded)
	}

	resolvedAmount, err := resolveAmountFromTransactionContext(transaction.PriceContext)
	if err != nil {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonAmountUnresolved)
	}

	runtimeResult, err := s.runtime.Refund(ctx, RefundRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		EscrowReference:      strings.TrimSpace(transaction.EscrowReference),
		Amount:               resolvedAmount,
	})
	if err != nil {
		result := Result{
			Status:                      ResultStatusFailed,
			TransactionReceiptID:        transaction.TransactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			ResolvedAmount:              resolvedAmount,
			Failure: &Failure{
				Kind:    FailureKindExecutionFailed,
				Message: err.Error(),
			},
		}
		failure := receipts.SettlementFailureRequest{
			TransactionReceiptID: transaction.TransactionReceiptID,
			SubmissionReceiptID:  submissionReceiptID,
			ResolvedAmount:       resolvedAmount,
			Reason:               err.Error(),
		}
		if recordErr := s.store.RecordEscrowRefundFailure(ctx, failure); recordErr != nil {
			return result, fmt.Errorf("record escrow refund failure: %w", recordErr)
		}
		return result, &ExecutionError{Kind: FailureKindExecutionFailed, Message: err.Error(), Err: err}
	}

	if err := s.store.RecordEscrowRefundSuccess(ctx, receipts.EscrowRefundEvidenceRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		RuntimeReference:     runtimeResult.Reference,
	}); err != nil {
		return Result{}, fmt.Errorf("record escrow refund success: %w", err)
	}

	return Result{
		Status:                      ResultStatusRefundExecuted,
		TransactionReceiptID:        transaction.TransactionReceiptID,
		SubmissionReceiptID:         submissionReceiptID,
		SettlementProgressionStatus: receipts.SettlementProgressionReviewNeeded,
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

func resolveAmountFromTransactionContext(priceContext string) (string, error) {
	trimmed := strings.TrimSpace(priceContext)
	if !strings.HasPrefix(trimmed, "quote:") || !strings.HasSuffix(trimmed, "-usdc") {
		return "", fmt.Errorf("unsupported price context %q", priceContext)
	}

	amount := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "quote:"), "-usdc"))
	if amount == "" {
		return "", fmt.Errorf("missing amount in price context %q", priceContext)
	}

	parsed, err := finance.ParseUSDC(amount)
	if err != nil {
		return "", err
	}
	if parsed.Sign() <= 0 {
		return "", fmt.Errorf("amount must be positive in price context %q", priceContext)
	}

	return finance.FormatUSDC(parsed), nil
}
