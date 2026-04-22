package settlementprogression

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/approvalflow"
	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	ApplySettlementProgression(context.Context, string, receipts.SettlementProgressionStatus, string, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) ApplyReleaseOutcome(ctx context.Context, req ApplyReleaseOutcomeRequest) (ApplyReleaseOutcomeResult, error) {
	if strings.TrimSpace(req.TransactionReceiptID) == "" {
		return ApplyReleaseOutcomeResult{}, fmt.Errorf("%w: transaction_receipt_id is required", ErrInvalidApplyReleaseOutcomeRequest)
	}
	if s == nil || s.store == nil {
		return ApplyReleaseOutcomeResult{}, fmt.Errorf("%w: receipt store is required", ErrInvalidApplyReleaseOutcomeRequest)
	}

	mapped, err := mapReleaseOutcome(req.Outcome)
	if err != nil {
		return ApplyReleaseOutcomeResult{}, err
	}

	transaction, err := s.store.ApplySettlementProgression(
		ctx,
		req.TransactionReceiptID,
		mapped.ProgressionStatus,
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

func mapReleaseOutcome(outcome ReleaseOutcome) (SettlementOutcome, error) {
	switch outcome.Decision {
	case approvalflow.DecisionApprove:
		return SettlementOutcome{
			ProgressionStatus: receipts.SettlementProgressionApprovedForSettlement,
			ProgressionReason: string(approvalflow.DecisionApprove),
		}, nil
	case approvalflow.DecisionReject:
		return SettlementOutcome{
			ProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			ProgressionReason: string(approvalflow.DecisionReject),
		}, nil
	case approvalflow.DecisionRequestRevision:
		return SettlementOutcome{
			ProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			ProgressionReason: string(approvalflow.DecisionRequestRevision),
		}, nil
	case approvalflow.DecisionEscalate:
		return SettlementOutcome{
			ProgressionStatus: receipts.SettlementProgressionReviewNeeded,
			ProgressionReason: escalationReason(outcome.Reason),
		}, nil
	default:
		return SettlementOutcome{}, fmt.Errorf("%w: %q", ErrUnsupportedReleaseDecision, outcome.Decision)
	}
}

func escalationReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "higher approval needed"
	}
	return reason
}
