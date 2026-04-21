package paymentgate

import (
	"context"
	"errors"
	"strings"

	"github.com/langoai/lango/internal/paymentapproval"
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
		if errors.Is(err, receipts.ErrTransactionReceiptNotFound) {
			return Result{Decision: Deny, Reason: ReasonMissingReceipt}, nil
		}
		return Result{}, err
	}

	if strings.TrimSpace(req.SubmissionReceiptID) == "" {
		return Result{Decision: Deny, Reason: ReasonMissingReceipt}, nil
	}

	if transaction.CurrentPaymentApprovalStatus != receipts.PaymentApprovalApproved {
		return Result{Decision: Deny, Reason: ReasonApprovalNotApproved}, nil
	}

	if transaction.CanonicalSettlementHint != string(paymentapproval.ModePrepay) {
		return Result{Decision: Deny, Reason: ReasonExecutionModeMismatch}, nil
	}

	return Result{Decision: Allow}, nil
}
