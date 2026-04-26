package settlementprogression

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/approvalflow"
	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	ApplySettlementProgression(context.Context, string, receipts.SettlementProgressionStatus, receipts.SettlementProgressionReasonCode, string, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) ApplyReleaseOutcome(ctx context.Context, req ApplyReleaseOutcomeRequest) (ApplyReleaseOutcomeResult, error) {
	transactionReceiptID := strings.TrimSpace(req.TransactionReceiptID)
	if transactionReceiptID == "" {
		return ApplyReleaseOutcomeResult{}, fmt.Errorf("%w: transaction_receipt_id is required", ErrInvalidApplyReleaseOutcomeRequest)
	}
	if s == nil || s.store == nil {
		return ApplyReleaseOutcomeResult{}, fmt.Errorf("%w: receipt store is required", ErrInvalidApplyReleaseOutcomeRequest)
	}

	current, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		return ApplyReleaseOutcomeResult{}, err
	}

	mapped, err := mapReleaseOutcome(req.Outcome, current.SettlementProgressionStatus)
	if err != nil {
		return ApplyReleaseOutcomeResult{}, err
	}
	mapped.PartialHint = strings.TrimSpace(req.PartialHint)

	transaction, err := s.store.ApplySettlementProgression(
		ctx,
		transactionReceiptID,
		mapped.ProgressionStatus,
		mapped.ProgressionReasonCode,
		mapped.ProgressionReason,
		mapped.PartialHint,
	)
	if err != nil {
		return ApplyReleaseOutcomeResult{}, err
	}

	return ApplyReleaseOutcomeResult{
		Transaction: transaction,
		Outcome:     mapped,
	}, nil
}

func mapReleaseOutcome(
	outcome ReleaseOutcome,
	current receipts.SettlementProgressionStatus,
) (SettlementOutcome, error) {
	switch outcome.Decision {
	case approvalflow.DecisionApprove:
		return SettlementOutcome{
			ProgressionStatus:     receipts.SettlementProgressionApprovedForSettlement,
			ProgressionReasonCode: receipts.SettlementProgressionReasonCodeApprove,
			ProgressionReason:     progressionReason(outcome.Reason, "Artifact release approved."),
		}, nil
	case approvalflow.DecisionReject:
		return SettlementOutcome{
			ProgressionStatus:     receipts.SettlementProgressionReviewNeeded,
			ProgressionReasonCode: receipts.SettlementProgressionReasonCodeReject,
			ProgressionReason:     progressionReason(outcome.Reason, "Artifact release rejected."),
		}, nil
	case approvalflow.DecisionRequestRevision:
		return SettlementOutcome{
			ProgressionStatus:     receipts.SettlementProgressionReviewNeeded,
			ProgressionReasonCode: receipts.SettlementProgressionReasonCodeRequestRevision,
			ProgressionReason:     progressionReason(outcome.Reason, "Artifact release requires revision."),
		}, nil
	case approvalflow.DecisionEscalate:
		return SettlementOutcome{
			ProgressionStatus:     escalationProgressionStatus(current),
			ProgressionReasonCode: receipts.SettlementProgressionReasonCodeEscalate,
			ProgressionReason:     progressionReason(outcome.Reason, "higher approval needed"),
		}, nil
	default:
		return SettlementOutcome{}, fmt.Errorf("%w: %q", ErrUnsupportedReleaseDecision, outcome.Decision)
	}
}

func progressionReason(reason string, fallback string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fallback
	}
	return reason
}

func escalationProgressionStatus(
	current receipts.SettlementProgressionStatus,
) receipts.SettlementProgressionStatus {
	switch current {
	case receipts.SettlementProgressionReviewNeeded,
		receipts.SettlementProgressionApprovedForSettlement,
		receipts.SettlementProgressionPartiallySettled,
		receipts.SettlementProgressionDisputeReady:
		return receipts.SettlementProgressionDisputeReady
	default:
		return receipts.SettlementProgressionReviewNeeded
	}
}
