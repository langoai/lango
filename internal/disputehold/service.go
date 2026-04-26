package disputehold

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error)
	RecordEscrowDisputeHoldSuccess(context.Context, receipts.EscrowDisputeHoldEvidenceRequest) error
	RecordEscrowDisputeHoldFailure(context.Context, receipts.EscrowDisputeHoldFailureRequest) error
}

type holdRuntime interface {
	Hold(context.Context, EscrowHoldRequest) (HoldResult, error)
}

type Service struct {
	store   receiptStore
	runtime holdRuntime
}

func NewService(store receiptStore, runtime holdRuntime) *Service {
	return &Service{store: store, runtime: runtime}
}

func (s *Service) Execute(ctx context.Context, req ExecuteRequest) (Result, error) {
	transactionReceiptID := strings.TrimSpace(req.TransactionReceiptID)
	if transactionReceiptID == "" {
		return deniedResult("", "", receipts.SettlementProgressionPending, "", "", DenyReasonMissingReceipt)
	}
	if s == nil || s.store == nil {
		return Result{}, fmt.Errorf("receipt store is required")
	}
	if s.runtime == nil {
		return Result{}, fmt.Errorf("dispute hold runtime is required")
	}

	transaction, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrTransactionReceiptNotFound) {
			return deniedResult(transactionReceiptID, "", receipts.SettlementProgressionPending, "", "", DenyReasonMissingReceipt)
		}
		return Result{}, err
	}

	submissionReceiptID := strings.TrimSpace(transaction.CurrentSubmissionReceiptID)
	if submissionReceiptID == "" {
		return deniedResult(transaction.TransactionReceiptID, "", transaction.SettlementProgressionStatus, transaction.DisputeLifecycleStatus, strings.TrimSpace(transaction.EscrowReference), DenyReasonNoCurrentSubmission)
	}

	submission, _, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
			return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, transaction.DisputeLifecycleStatus, strings.TrimSpace(transaction.EscrowReference), DenyReasonNoCurrentSubmission)
		}
		return Result{}, err
	}
	if submission.TransactionReceiptID != transaction.TransactionReceiptID {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, transaction.DisputeLifecycleStatus, strings.TrimSpace(transaction.EscrowReference), DenyReasonNoCurrentSubmission)
	}

	if transaction.EscrowExecutionStatus != receipts.EscrowExecutionStatusFunded {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, transaction.DisputeLifecycleStatus, strings.TrimSpace(transaction.EscrowReference), DenyReasonEscrowNotFunded)
	}
	if transaction.SettlementProgressionStatus != receipts.SettlementProgressionDisputeReady {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, transaction.DisputeLifecycleStatus, strings.TrimSpace(transaction.EscrowReference), DenyReasonNotDisputeReady)
	}

	escrowReference := strings.TrimSpace(transaction.EscrowReference)
	if escrowReference == "" {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, transaction.DisputeLifecycleStatus, "", DenyReasonEscrowReferenceMissing)
	}

	runtimeResult, err := s.runtime.Hold(ctx, EscrowHoldRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		EscrowReference:      escrowReference,
	})
	if err != nil {
		result := Result{
			Status:                      ResultStatusFailed,
			TransactionReceiptID:        transaction.TransactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
			DisputeLifecycleStatus:      transaction.DisputeLifecycleStatus,
			EscrowReference:             escrowReference,
			Failure: &Failure{
				Kind:    FailureKindExecutionFailed,
				Message: err.Error(),
			},
		}
		failure := receipts.EscrowDisputeHoldFailureRequest{
			TransactionReceiptID: transaction.TransactionReceiptID,
			SubmissionReceiptID:  submissionReceiptID,
			EscrowReference:      escrowReference,
			Reason:               err.Error(),
		}
		if recordErr := s.store.RecordEscrowDisputeHoldFailure(ctx, failure); recordErr != nil {
			return result, fmt.Errorf("record dispute hold failure: %w", recordErr)
		}
		return result, &ExecutionError{Kind: FailureKindExecutionFailed, Message: err.Error(), Err: err}
	}

	if err := s.store.RecordEscrowDisputeHoldSuccess(ctx, receipts.EscrowDisputeHoldEvidenceRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		EscrowReference:      escrowReference,
		RuntimeReference:     runtimeResult.Reference,
	}); err != nil {
		return Result{}, fmt.Errorf("record dispute hold success: %w", err)
	}

	return Result{
		Status:                      ResultStatusHoldApplied,
		TransactionReceiptID:        transaction.TransactionReceiptID,
		SubmissionReceiptID:         submissionReceiptID,
		SettlementProgressionStatus: receipts.SettlementProgressionDisputeReady,
		DisputeLifecycleStatus:      receipts.DisputeLifecycleHoldActive,
		EscrowReference:             escrowReference,
		RuntimeReference:            runtimeResult.Reference,
	}, nil
}

func deniedResult(
	transactionReceiptID, submissionReceiptID string,
	status receipts.SettlementProgressionStatus,
	disputeLifecycleStatus receipts.DisputeLifecycleStatus,
	escrowReference string,
	reason DenyReason,
) (Result, error) {
	return Result{
			Status:                      ResultStatusDenied,
			TransactionReceiptID:        transactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: status,
			DisputeLifecycleStatus:      disputeLifecycleStatus,
			EscrowReference:             escrowReference,
			Failure: &Failure{
				Kind:       FailureKindDenied,
				DenyReason: reason,
				Message:    string(reason),
			},
		},
		&ExecutionError{Kind: FailureKindDenied, DenyReason: reason, Message: string(reason)}
}
