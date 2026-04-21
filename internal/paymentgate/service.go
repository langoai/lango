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

func (s *Service) EvaluateDirectPayment(ctx context.Context, req Request) (Result, error) {
	if strings.TrimSpace(req.TransactionReceiptID) == "" {
		return Result{Decision: Deny, Reason: ReasonMissingReceipt}, nil
	}

	transaction, err := s.store.GetTransactionReceipt(ctx, req.TransactionReceiptID)
	if err != nil {
		if err == receipts.ErrTransactionReceiptNotFound {
			return Result{Decision: Deny, Reason: ReasonMissingReceipt}, nil
		}
		return Result{}, err
	}

	if transaction.CurrentPaymentApprovalStatus != receipts.PaymentApprovalApproved {
		return Result{Decision: Deny, Reason: ReasonApprovalNotApproved}, nil
	}

	if transaction.CanonicalSettlementHint != "prepay" {
		return Result{Decision: Deny, Reason: ReasonExecutionModeMismatch}, nil
	}

	return Result{Decision: Allow}, nil
}
