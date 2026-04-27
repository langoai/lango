package escrowrelease

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
	MarkSettlementSettled(context.Context, receipts.SettlementCloseoutRequest) (receipts.TransactionReceipt, error)
	RecordSettlementFailure(context.Context, receipts.SettlementFailureRequest) error
}

type releaseRuntime interface {
	Release(context.Context, ReleaseRequest) (ReleaseResult, error)
}

type Service struct {
	store   receiptStore
	runtime releaseRuntime
}

func NewService(store receiptStore, runtime releaseRuntime) *Service {
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
		return Result{}, fmt.Errorf("escrow runtime is required")
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

	var events []receipts.ReceiptEvent
	submission, events, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
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

	if transaction.SettlementProgressionStatus != receipts.SettlementProgressionApprovedForSettlement {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonNotApprovedForSettlement)
	}
	if transaction.EscrowAdjudication == "" {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonAdjudicationMissing)
	}
	if transaction.EscrowAdjudication != receipts.EscrowAdjudicationRelease {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonAdjudicationMismatch)
	}
	if hasOppositeRefundEvidence(events) {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonAdjudicationMismatch)
	}

	resolvedAmount, err := resolveAmountFromTransactionContext(transaction.PriceContext)
	if err != nil {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, DenyReasonAmountUnresolved)
	}

	runtimeResult, err := s.runtime.Release(ctx, ReleaseRequest{
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
			SettlementProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
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
		if recordErr := s.store.RecordSettlementFailure(ctx, failure); recordErr != nil {
			return result, fmt.Errorf("record escrow release failure: %w", recordErr)
		}
		return result, &ExecutionError{Kind: FailureKindExecutionFailed, Message: err.Error(), Err: err}
	}

	updated, err := s.store.MarkSettlementSettled(ctx, receipts.SettlementCloseoutRequest{
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

func hasOppositeRefundEvidence(events []receipts.ReceiptEvent) bool {
	for _, event := range events {
		if event.Source == "escrow_refund" && event.Subtype == "refunded" {
			return true
		}
	}
	return false
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
	amount := strings.TrimSpace(priceContext)
	if amount == "" {
		return "", fmt.Errorf("missing price context")
	}
	if !strings.HasPrefix(amount, "quote:") || !strings.HasSuffix(amount, "-usdc") {
		return "", fmt.Errorf("unsupported price context %q", priceContext)
	}
	amount = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(amount, "quote:"), "-usdc"))
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
