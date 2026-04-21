package escrowexecution

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/paymentapproval"
	"github.com/langoai/lango/internal/receipts"
	"github.com/langoai/lango/internal/wallet"
)

type receiptStore interface {
	GetTransactionReceipt(context.Context, string) (receipts.TransactionReceipt, error)
	ApplyEscrowExecutionProgress(context.Context, string, string, receipts.EscrowExecutionStatus, string, receipts.EventType, string) (receipts.TransactionReceipt, error)
}

type runtime interface {
	Create(context.Context, escrow.CreateRequest) (*escrow.EscrowEntry, error)
	Fund(context.Context, string) (*escrow.EscrowEntry, error)
}

type Service struct {
	store   receiptStore
	runtime runtime
}

func NewService(store receiptStore, runtime runtime) *Service {
	return &Service{
		store:   store,
		runtime: runtime,
	}
}

func (s *Service) ExecuteRecommendation(ctx context.Context, req Request) (Result, error) {
	transactionReceiptID := strings.TrimSpace(req.TransactionReceiptID)
	if transactionReceiptID == "" {
		return Result{}, fmt.Errorf("transaction receipt id is required")
	}

	transaction, err := s.store.GetTransactionReceipt(ctx, transactionReceiptID)
	if err != nil {
		return Result{}, fmt.Errorf("load transaction receipt %q: %w", transactionReceiptID, err)
	}

	submissionReceiptID := strings.TrimSpace(transaction.CurrentSubmissionReceiptID)
	if submissionReceiptID == "" {
		return Result{}, fmt.Errorf("transaction receipt %q has no current submission receipt", transactionReceiptID)
	}
	if transaction.CurrentPaymentApprovalStatus != receipts.PaymentApprovalApproved {
		return Result{}, fmt.Errorf("transaction receipt %q payment approval status is %q, want %q", transactionReceiptID, transaction.CurrentPaymentApprovalStatus, receipts.PaymentApprovalApproved)
	}
	if transaction.CanonicalSettlementHint != string(paymentapproval.ModeEscrow) {
		return Result{}, fmt.Errorf("transaction receipt %q settlement hint is %q, want %q", transactionReceiptID, transaction.CanonicalSettlementHint, paymentapproval.ModeEscrow)
	}
	if transaction.EscrowExecutionInput == nil {
		return Result{}, fmt.Errorf("transaction receipt %q is missing bound escrow execution input", transactionReceiptID)
	}

	createReq, err := buildCreateRequest(*transaction.EscrowExecutionInput)
	if err != nil {
		return Result{}, err
	}

	if _, err := s.store.ApplyEscrowExecutionProgress(
		ctx,
		transactionReceiptID,
		submissionReceiptID,
		receipts.EscrowExecutionStatusPending,
		"",
		receipts.EventEscrowExecutionStarted,
		"",
	); err != nil {
		return Result{}, fmt.Errorf("record escrow started progress for transaction receipt %q: %w", transactionReceiptID, err)
	}

	createdEntry, err := s.runtime.Create(ctx, createReq)
	if err != nil {
		opErr := fmt.Errorf("create escrow for transaction receipt %q: %w", transactionReceiptID, err)
		return Result{}, s.appendFailure(ctx, transactionReceiptID, submissionReceiptID, "", opErr)
	}

	createdEscrowID := escrowIDFromEntry(createdEntry)
	if createdEscrowID == "" {
		opErr := fmt.Errorf("create escrow for transaction receipt %q: runtime returned empty escrow id", transactionReceiptID)
		return Result{}, s.appendFailure(ctx, transactionReceiptID, submissionReceiptID, "", opErr)
	}

	if _, err := s.store.ApplyEscrowExecutionProgress(
		ctx,
		transactionReceiptID,
		submissionReceiptID,
		receipts.EscrowExecutionStatusCreated,
		createdEscrowID,
		receipts.EventEscrowExecutionCreated,
		"",
	); err != nil {
		return Result{}, fmt.Errorf("record escrow created progress for transaction receipt %q: %w", transactionReceiptID, err)
	}

	fundedEntry, err := s.runtime.Fund(ctx, createdEscrowID)
	if err != nil {
		opErr := fmt.Errorf("fund escrow %q for transaction receipt %q: %w", createdEscrowID, transactionReceiptID, err)
		return Result{}, s.appendFailure(ctx, transactionReceiptID, submissionReceiptID, createdEscrowID, opErr)
	}

	fundedEscrowID := escrowIDFromEntry(fundedEntry)
	if fundedEscrowID == "" {
		fundedEscrowID = createdEscrowID
	}

	updatedTx, err := s.store.ApplyEscrowExecutionProgress(
		ctx,
		transactionReceiptID,
		submissionReceiptID,
		receipts.EscrowExecutionStatusFunded,
		fundedEscrowID,
		receipts.EventEscrowExecutionFunded,
		"",
	)
	if err != nil {
		return Result{}, fmt.Errorf("record escrow funded progress for transaction receipt %q: %w", transactionReceiptID, err)
	}

	return Result{
		TransactionReceiptID:  transactionReceiptID,
		SubmissionReceiptID:   submissionReceiptID,
		EscrowReference:       fundedEscrowID,
		EscrowExecutionStatus: updatedTx.EscrowExecutionStatus,
	}, nil
}

func buildCreateRequest(input receipts.EscrowExecutionInput) (escrow.CreateRequest, error) {
	amount, err := wallet.ParseUSDC(input.Amount)
	if err != nil {
		return escrow.CreateRequest{}, fmt.Errorf("parse escrow amount %q: %w", input.Amount, err)
	}

	milestones := make([]escrow.MilestoneRequest, len(input.Milestones))
	for i, milestone := range input.Milestones {
		milestoneAmount, err := wallet.ParseUSDC(milestone.Amount)
		if err != nil {
			return escrow.CreateRequest{}, fmt.Errorf("parse escrow milestone %d amount %q: %w", i, milestone.Amount, err)
		}
		milestones[i] = escrow.MilestoneRequest{
			Description: milestone.Description,
			Amount:      milestoneAmount,
		}
	}

	return escrow.CreateRequest{
		BuyerDID:   input.BuyerDID,
		SellerDID:  input.SellerDID,
		Amount:     amount,
		Reason:     input.Reason,
		TaskID:     input.TaskID,
		Milestones: milestones,
	}, nil
}

func (s *Service) appendFailure(ctx context.Context, transactionReceiptID, submissionReceiptID, escrowID string, opErr error) error {
	_, err := s.store.ApplyEscrowExecutionProgress(
		ctx,
		transactionReceiptID,
		submissionReceiptID,
		receipts.EscrowExecutionStatusFailed,
		escrowID,
		receipts.EventEscrowExecutionFailed,
		opErr.Error(),
	)
	if err != nil {
		return fmt.Errorf("%w; record escrow execution failure: %v", opErr, err)
	}
	return opErr
}

func escrowIDFromEntry(entry *escrow.EscrowEntry) string {
	if entry == nil {
		return ""
	}
	return strings.TrimSpace(entry.ID)
}
