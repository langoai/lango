package knowledgeruntime

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
)

type receiptStore interface {
	OpenKnowledgeExchangeTransaction(context.Context, receipts.OpenTransactionInput) (receipts.TransactionReceipt, error)
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	GetSubmissionReceipt(context.Context, string) (receipts.SubmissionReceipt, []receipts.ReceiptEvent, error)
	ApplyKnowledgeExchangeRuntimeProgression(context.Context, string, receipts.KnowledgeExchangeRuntimeStatus, string) (receipts.TransactionReceipt, error)
}

type Service struct {
	store receiptStore
}

func NewService(store receiptStore) *Service {
	return &Service{store: store}
}

func (s *Service) OpenTransaction(ctx context.Context, req OpenTransactionRequest) (OpenTransactionResult, error) {
	store, err := s.requireStore("open transaction")
	if err != nil {
		return OpenTransactionResult{}, err
	}

	tx, err := store.OpenKnowledgeExchangeTransaction(ctx, receipts.OpenTransactionInput{
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
	store, err := s.requireStore("select execution path")
	if err != nil {
		return BranchSelection{}, err
	}
	transactionReceiptID = strings.TrimSpace(transactionReceiptID)
	if transactionReceiptID == "" {
		return BranchSelection{}, fmt.Errorf("transaction_receipt_id is required")
	}

	tx, err := store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		return BranchSelection{}, err
	}
	if tx.CurrentSubmissionReceiptID == "" {
		return BranchSelection{}, fmt.Errorf("transaction %q has no current submission receipt bound", transactionReceiptID)
	}

	_, events, err := store.GetSubmissionReceipt(ctx, tx.CurrentSubmissionReceiptID)
	if err != nil {
		return BranchSelection{}, err
	}
	if !hasPaymentApprovalEvent(events, tx.CurrentSubmissionReceiptID) {
		return BranchSelection{}, fmt.Errorf("transaction %q current submission %q has no canonical payment approval state", transactionReceiptID, tx.CurrentSubmissionReceiptID)
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

	if tx.KnowledgeExchangeRuntimeStatus != receipts.RuntimeStatusPaymentApproved {
		if _, err := store.ApplyKnowledgeExchangeRuntimeProgression(ctx, transactionReceiptID, receipts.RuntimeStatusPaymentApproved, tx.CurrentSubmissionReceiptID); err != nil {
			return BranchSelection{}, err
		}
	}

	return BranchSelection{
		TransactionReceiptID: transactionReceiptID,
		Branch:               branch,
	}, nil
}

func hasPaymentApprovalEvent(events []receipts.ReceiptEvent, submissionReceiptID string) bool {
	for _, event := range events {
		if event.SubmissionReceiptID == submissionReceiptID && event.Type == receipts.EventPaymentApproval {
			return true
		}
	}

	return false
}

func (s *Service) requireStore(action string) (receiptStore, error) {
	if s == nil || s.store == nil {
		return nil, fmt.Errorf("%s: knowledge runtime receipt store is required", action)
	}
	return s.store, nil
}
