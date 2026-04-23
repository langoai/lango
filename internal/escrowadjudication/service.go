package escrowadjudication

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
	ApplyEscrowAdjudication(context.Context, receipts.EscrowAdjudicationRequest) (receipts.TransactionReceipt, error)
	RecordEscrowAdjudicationFailure(context.Context, receipts.EscrowAdjudicationFailureRequest) error
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) Adjudicate(ctx context.Context, req AdjudicateRequest) (Result, error) {
	transactionReceiptID := strings.TrimSpace(req.TransactionReceiptID)
	if transactionReceiptID == "" {
		return deniedResult("", "", receipts.SettlementProgressionPending, "", "", DenyReasonMissingReceipt)
	}
	if s == nil || s.store == nil {
		return Result{}, fmt.Errorf("receipt store is required")
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
		return deniedResult(transaction.TransactionReceiptID, "", transaction.SettlementProgressionStatus, strings.TrimSpace(transaction.EscrowReference), "", DenyReasonNoCurrentSubmission)
	}

	submission, events, err := s.store.GetSubmissionReceipt(ctx, submissionReceiptID)
	if err != nil {
		if errors.Is(err, receipts.ErrSubmissionReceiptNotFound) {
			return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, strings.TrimSpace(transaction.EscrowReference), "", DenyReasonNoCurrentSubmission)
		}
		return Result{}, err
	}
	if submission.TransactionReceiptID != transaction.TransactionReceiptID {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, strings.TrimSpace(transaction.EscrowReference), "", DenyReasonNoCurrentSubmission)
	}

	escrowReference := strings.TrimSpace(transaction.EscrowReference)
	if transaction.EscrowExecutionStatus != receipts.EscrowExecutionStatusFunded {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, escrowReference, "", DenyReasonEscrowNotFunded)
	}
	if transaction.SettlementProgressionStatus != receipts.SettlementProgressionDisputeReady {
		return deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, escrowReference, "", DenyReasonNotDisputeReady)
	}
	if !hasDisputeHoldEvidence(events) {
		result, err := deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, escrowReference, "", DenyReasonHoldEvidenceMissing)
		if recordErr := s.recordFailure(ctx, transaction.TransactionReceiptID, submissionReceiptID, escrowReference, string(DenyReasonHoldEvidenceMissing)); recordErr != nil {
			return result, recordErr
		}
		return result, err
	}

	outcome, ok := mapOutcome(req.Outcome)
	if !ok {
		result, err := deniedResult(transaction.TransactionReceiptID, submissionReceiptID, transaction.SettlementProgressionStatus, escrowReference, string(req.Outcome), DenyReasonInvalidOutcome)
		if recordErr := s.recordFailure(ctx, transaction.TransactionReceiptID, submissionReceiptID, escrowReference, string(DenyReasonInvalidOutcome)); recordErr != nil {
			return result, recordErr
		}
		return result, err
	}

	updated, err := s.store.ApplyEscrowAdjudication(ctx, receipts.EscrowAdjudicationRequest{
		TransactionReceiptID: transaction.TransactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		EscrowReference:      escrowReference,
		Outcome:              outcome,
		Reason:               strings.TrimSpace(req.Reason),
	})
	if err != nil {
		result := Result{
			Status:                      ResultStatusFailed,
			TransactionReceiptID:        transaction.TransactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: transaction.SettlementProgressionStatus,
			EscrowReference:             escrowReference,
			Outcome:                     req.Outcome,
			Failure: &Failure{
				Kind:    FailureKindApplyFailed,
				Message: err.Error(),
			},
		}
		if recordErr := s.recordFailure(ctx, transaction.TransactionReceiptID, submissionReceiptID, escrowReference, err.Error()); recordErr != nil {
			return result, fmt.Errorf("record adjudication failure: %w", recordErr)
		}
		return result, &ExecutionError{Kind: FailureKindApplyFailed, Message: err.Error(), Err: err}
	}

	return Result{
		Status:                      ResultStatusAdjudicated,
		TransactionReceiptID:        updated.TransactionReceiptID,
		SubmissionReceiptID:         submissionReceiptID,
		SettlementProgressionStatus: updated.SettlementProgressionStatus,
		EscrowReference:             escrowReference,
		Outcome:                     req.Outcome,
	}, nil
}

func (s *Service) recordFailure(ctx context.Context, transactionReceiptID, submissionReceiptID, escrowReference, reason string) error {
	return s.store.RecordEscrowAdjudicationFailure(ctx, receipts.EscrowAdjudicationFailureRequest{
		TransactionReceiptID: transactionReceiptID,
		SubmissionReceiptID:  submissionReceiptID,
		EscrowReference:      escrowReference,
		Reason:               reason,
	})
}

func hasDisputeHoldEvidence(events []receipts.ReceiptEvent) bool {
	for _, event := range events {
		if event.Source == "dispute_hold" && event.Subtype == "held" {
			return true
		}
	}
	return false
}

func mapOutcome(outcome Outcome) (receipts.EscrowAdjudicationDecision, bool) {
	switch outcome {
	case OutcomeRelease:
		return receipts.EscrowAdjudicationRelease, true
	case OutcomeRefund:
		return receipts.EscrowAdjudicationRefund, true
	default:
		return "", false
	}
}

func deniedResult(transactionReceiptID, submissionReceiptID string, status receipts.SettlementProgressionStatus, escrowReference, outcome string, reason DenyReason) (Result, error) {
	return Result{
			Status:                      ResultStatusDenied,
			TransactionReceiptID:        transactionReceiptID,
			SubmissionReceiptID:         submissionReceiptID,
			SettlementProgressionStatus: status,
			EscrowReference:             escrowReference,
			Outcome:                     Outcome(outcome),
			Failure: &Failure{
				Kind:       FailureKindDenied,
				DenyReason: reason,
				Message:    string(reason),
			},
		},
		&ExecutionError{Kind: FailureKindDenied, DenyReason: reason, Message: string(reason)}
}
