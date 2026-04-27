package partialsettlementexecution

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/langoai/lango/internal/finance"
	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error)
	MarkPartialSettlementSettled(context.Context, receipts.PartialSettlementCloseoutRequest) (receipts.TransactionReceipt, error)
	RecordPartialSettlementSuccess(context.Context, receipts.PartialSettlementExecutionEvidenceRequest) error
	RecordPartialSettlementFailure(context.Context, receipts.PartialSettlementFailureRequest) error
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

	if transaction.SettlementProgressionStatus == receipts.SettlementProgressionPartiallySettled {
		return deniedResult(
			transaction.TransactionReceiptID,
			strings.TrimSpace(transaction.CurrentSubmissionReceiptID),
			transaction.SettlementProgressionStatus,
			DenyReasonAlreadyPartiallySettled,
		)
	}
	if transaction.SettlementProgressionStatus != receipts.SettlementProgressionApprovedForSettlement {
		return deniedResult(
			transaction.TransactionReceiptID,
			strings.TrimSpace(transaction.CurrentSubmissionReceiptID),
			transaction.SettlementProgressionStatus,
			DenyReasonNotApprovedForSettlement,
		)
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

	partialHint := strings.TrimSpace(transaction.PartialSettlementHint)
	if partialHint == "" {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonPartialHintMissing)
	}

	executedAmount, err := parsePartialSettlementHint(partialHint)
	if err != nil {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonPartialHintInvalid)
	}

	totalAmount, err := resolveQuoteAmount(transaction.PriceContext)
	if err != nil {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonPartialHintInvalid)
	}

	remainingAmount, err := calculateRemainingAmount(totalAmount, executedAmount)
	if err != nil {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonPartialHintInvalid)
	}

	runtimeResult, err := s.runtime.ExecuteSettlement(ctx, DirectPaymentRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		Counterparty:         transaction.Counterparty,
		Amount:               executedAmount,
	})
	if err != nil {
		result := Result{
			Status:                      ResultStatusFailed,
			TransactionReceiptID:        transaction.TransactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			ExecutedAmount:              executedAmount,
			RemainingAmount:             remainingAmount,
			Failure: &Failure{
				Kind:    FailureKindExecutionFailed,
				Message: err.Error(),
			},
		}
		failure := receipts.PartialSettlementFailureRequest{
			TransactionReceiptID: transaction.TransactionReceiptID,
			SubmissionReceiptID:  submissionReceiptID,
			ExecutedAmount:       executedAmount,
			RemainingAmount:      remainingAmount,
			Reason:               err.Error(),
		}
		if recordErr := s.store.RecordPartialSettlementFailure(ctx, failure); recordErr != nil {
			return result, fmt.Errorf("record partial settlement failure: %w", recordErr)
		}
		return result, &ExecutionError{Kind: FailureKindExecutionFailed, Message: err.Error(), Err: err}
	}

	updated, err := s.store.MarkPartialSettlementSettled(ctx, receipts.PartialSettlementCloseoutRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		ExecutedAmount:       executedAmount,
		RemainingAmount:      remainingAmount,
		RuntimeReference:     runtimeResult.Reference,
	})
	if err != nil {
		return Result{}, err
	}
	if err := s.store.RecordPartialSettlementSuccess(ctx, receipts.PartialSettlementExecutionEvidenceRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		RuntimeReference:     runtimeResult.Reference,
	}); err != nil {
		return Result{}, fmt.Errorf("record partial settlement success: %w", err)
	}

	return Result{
		Status:                      ResultStatusPartiallySettledTarget,
		TransactionReceiptID:        updated.TransactionReceiptID,
		SubmissionReceiptID:         submissionReceiptID,
		SettlementProgressionStatus: updated.SettlementProgressionStatus,
		ExecutedAmount:              executedAmount,
		RemainingAmount:             remainingAmount,
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

func parsePartialSettlementHint(hint string) (string, error) {
	return parseUSDCContextValue(hint, "settle:", "-usdc")
}

func resolveQuoteAmount(priceContext string) (string, error) {
	return parseUSDCContextValue(priceContext, "quote:", "-usdc")
}

func parseUSDCContextValue(raw, prefix, suffix string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, prefix) || !strings.HasSuffix(trimmed, suffix) {
		return "", fmt.Errorf("unsupported amount context %q", raw)
	}

	amount := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, prefix), suffix))
	if amount == "" {
		return "", fmt.Errorf("missing amount in %q", raw)
	}

	parsed, err := finance.ParseUSDC(amount)
	if err != nil {
		return "", err
	}
	if parsed.Sign() <= 0 {
		return "", fmt.Errorf("amount must be positive in %q", raw)
	}

	return finance.FormatUSDC(parsed), nil
}

func calculateRemainingAmount(totalAmount, executedAmount string) (string, error) {
	total, err := finance.ParseUSDC(totalAmount)
	if err != nil {
		return "", err
	}
	executed, err := finance.ParseUSDC(executedAmount)
	if err != nil {
		return "", err
	}
	if executed.Cmp(total) > 0 {
		return "", fmt.Errorf("executed amount %s exceeds total amount %s", executedAmount, totalAmount)
	}
	if executed.Cmp(total) == 0 {
		return "", fmt.Errorf("executed amount %s must leave a positive remaining amount from total %s", executedAmount, totalAmount)
	}

	remaining := new(big.Int).Sub(total, executed)
	return finance.FormatUSDC(remaining), nil
}
