package knowledgeruntime

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	OpenKnowledgeExchangeTransaction(context.Context, receipts.OpenTransactionInput) (receipts.TransactionReceipt, error)
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	ApplyKnowledgeExchangeRuntimeProgression(context.Context, string, receipts.KnowledgeExchangeRuntimeStatus, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) OpenTransaction(ctx context.Context, req OpenTransactionRequest) (OpenTransactionResult, error) {
	tx, err := s.store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
		TransactionID:  req.TransactionID,
		Counterparty:   req.Counterparty,
		RequestedScope: req.RequestedScope,
		PriceContext:   req.PriceContext,
		TrustContext:   req.TrustContext,
	})
	if err != nil {
		return OpenTransactionResult{}, err
	}

	return OpenTransactionResult{
		TransactionReceiptID: tx.TransactionReceiptID,
		RuntimeStatus:        tx.KnowledgeExchangeRuntimeStatus,
	}, nil
}

func (s *Service) SelectExecutionPath(ctx context.Context, transactionReceiptID string) (BranchSelection, error) {
	tx, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		return BranchSelection{}, err
	}

	var branch Branch
	switch tx.CanonicalSettlementHint {
	case string(paymentapproval.ModePrepay):
		branch = BranchPrepay
	case string(paymentapproval.ModeEscrow):
		branch = BranchEscrow
	default:
		return BranchSelection{}, fmt.Errorf("transaction %q has unsupported settlement hint %q", transactionReceiptID, tx.CanonicalSettlementHint)
	}

	if _, err := s.store.ApplyKnowledgeExchangeRuntimeProgression(ctx, transactionReceiptID, receipts.RuntimeStatusPaymentApproved, ""); err != nil {
		return BranchSelection{}, err
	}

	return BranchSelection{
		TransactionReceiptID: transactionReceiptID,
		Branch:               branch,
	}, nil
}
