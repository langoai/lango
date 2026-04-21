package paymentgate

import (
	"context"
	"strings"

	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) Decide(ctx context.Context, req Request) (Result, error) {
	if strings.TrimSpace(req.TransactionReceiptID) == "" {
		return Result{Decision: DecisionDeny, DenyReason: DenyReasonMissingReceipt}, nil
	}

	transaction, err := s.store.GetTransactionReceipt(ctx, req.TransactionReceiptID)
	if err != nil {
		if err == receipts.ErrTransactionReceiptNotFound {
			return Result{Decision: DecisionDeny, DenyReason: DenyReasonMissingReceipt}, nil
		}
		return Result{}, err
	}

	if transaction.CurrentPaymentApprovalStatus != receipts.PaymentApprovalApproved {
		return Result{Decision: DecisionDeny, DenyReason: DenyReasonApprovalNotApproved}, nil
	}

	if transaction.CanonicalSettlementHint != "prepay" {
		return Result{Decision: DecisionDeny, DenyReason: DenyReasonExecutionModeMismatch}, nil
	}

	return Result{Decision: DecisionAllow}, nil
}
