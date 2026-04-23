package receipts

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/finance"
	"github.com/langoai/lango/internal/paymentapproval"
)

type CreateSubmissionInput struct {
	TransactionID       string
	ArtifactLabel       string
	PayloadHash         string
	SourceLineageDigest string
}

type ReceiptEvent struct {
	SubmissionReceiptID string
	Source              string
	Subtype             string
	Reason              string
	Type                EventType
}

type Store struct {
	mu sync.Mutex

	submissions    map[string]SubmissionReceipt
	transactions   map[string]TransactionReceipt
	events         map[string][]ReceiptEvent
	txByExternalID map[string]string
}

func NewStore() *Store {
	return &Store{
		submissions:    make(map[string]SubmissionReceipt),
		transactions:   make(map[string]TransactionReceipt),
		events:         make(map[string][]ReceiptEvent),
		txByExternalID: make(map[string]string),
	}
}

func (s *Store) CreateSubmissionReceipt(_ context.Context, in CreateSubmissionInput) (SubmissionReceipt, TransactionReceipt, error) {
	if err := validateCreateSubmissionInput(in); err != nil {
		return SubmissionReceipt{}, TransactionReceipt{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	txReceiptID, ok := s.txByExternalID[in.TransactionID]
	if !ok {
		txReceiptID = uuid.NewString()
		s.txByExternalID[in.TransactionID] = txReceiptID
		s.transactions[txReceiptID] = TransactionReceipt{
			TransactionReceiptID:         txReceiptID,
			TransactionID:                in.TransactionID,
			SettlementProgressionStatus:  SettlementProgressionPending,
			CanonicalApprovalStatus:      ApprovalPending,
			CanonicalSettlementStatus:    SettlementPending,
			CurrentPaymentApprovalStatus: PaymentApprovalPending,
		}
	}

	submissionReceiptID := uuid.NewString()
	submission := SubmissionReceipt{
		SubmissionReceiptID:     submissionReceiptID,
		TransactionReceiptID:    txReceiptID,
		ArtifactLabel:           in.ArtifactLabel,
		PayloadHash:             in.PayloadHash,
		SourceLineageDigest:     in.SourceLineageDigest,
		CanonicalApprovalStatus: ApprovalPending,
		ProvenanceSummary: ProvenanceSummary{
			ReferenceID: submissionReceiptID,
		},
	}
	s.submissions[submissionReceiptID] = submission

	transaction := s.transactions[txReceiptID]
	transaction.CurrentSubmissionReceiptID = submissionReceiptID
	transaction.EscrowExecutionStatus = ""
	transaction.EscrowReference = ""
	transaction.EscrowAdjudication = ""
	transaction.EscrowExecutionInput = nil
	s.transactions[txReceiptID] = transaction

	return submission, cloneTransactionReceipt(transaction), nil
}

func (s *Store) OpenKnowledgeExchangeTransaction(_ context.Context, in OpenTransactionInput) (TransactionReceipt, error) {
	if strings.TrimSpace(in.TransactionID) == "" || strings.TrimSpace(in.Counterparty) == "" || strings.TrimSpace(in.RequestedScope) == "" {
		return TransactionReceipt{}, fmt.Errorf("%w: transaction_id, counterparty, and requested_scope are required", ErrInvalidSubmissionInput)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	txReceiptID, ok := s.txByExternalID[in.TransactionID]
	if !ok {
		txReceiptID = uuid.NewString()
		s.txByExternalID[in.TransactionID] = txReceiptID
	}

	tx := TransactionReceipt{
		TransactionReceiptID:           txReceiptID,
		TransactionID:                  in.TransactionID,
		Counterparty:                   in.Counterparty,
		RequestedScope:                 in.RequestedScope,
		PriceContext:                   in.PriceContext,
		TrustContext:                   in.TrustContext,
		KnowledgeExchangeRuntimeStatus: RuntimeStatusOpened,
		SettlementProgressionStatus:    SettlementProgressionPending,
		CanonicalApprovalStatus:        ApprovalPending,
		CanonicalSettlementStatus:      SettlementPending,
		CurrentPaymentApprovalStatus:   PaymentApprovalPending,
	}

	if existing, exists := s.transactions[txReceiptID]; exists {
		if err := validateCanonicalOpenInputConflict(existing, in); err != nil {
			return TransactionReceipt{}, err
		}
		tx.CurrentSubmissionReceiptID = existing.CurrentSubmissionReceiptID
		tx.CanonicalApprovalStatus = existing.CanonicalApprovalStatus
		tx.CanonicalSettlementStatus = existing.CanonicalSettlementStatus
		tx.CurrentPaymentApprovalStatus = existing.CurrentPaymentApprovalStatus
		tx.SettlementProgressionStatus = existing.SettlementProgressionStatus
		tx.SettlementProgressionReasonCode = existing.SettlementProgressionReasonCode
		tx.SettlementProgressionReason = existing.SettlementProgressionReason
		tx.PartialSettlementHint = existing.PartialSettlementHint
		tx.DisputeReady = existing.DisputeReady
		tx.CanonicalDecision = existing.CanonicalDecision
		tx.CanonicalSettlementHint = existing.CanonicalSettlementHint
		tx.EscrowExecutionStatus = existing.EscrowExecutionStatus
		tx.EscrowReference = existing.EscrowReference
		tx.EscrowAdjudication = existing.EscrowAdjudication
		tx.EscrowExecutionInput = cloneEscrowExecutionInput(existing.EscrowExecutionInput)
	}

	s.transactions[txReceiptID] = tx
	return cloneTransactionReceipt(tx), nil
}

func (s *Store) ApplySettlementProgression(_ context.Context, transactionReceiptID string, next SettlementProgressionStatus, reasonCode SettlementProgressionReasonCode, reason string, partialHint string) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	if tx.CurrentSubmissionReceiptID == "" {
		return TransactionReceipt{}, fmt.Errorf("%w: current submission receipt is required", ErrInvalidSettlementProgressionState)
	}
	submissionReceiptID := tx.CurrentSubmissionReceiptID
	submission, ok := s.submissions[submissionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != transactionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if err := validateSettlementProgressionTransition(tx.SettlementProgressionStatus, next); err != nil {
		return TransactionReceipt{}, err
	}
	if err := validateSettlementProgressionReasonCode(next, reasonCode); err != nil {
		return TransactionReceipt{}, err
	}

	tx.SettlementProgressionStatus = next
	tx.CanonicalSettlementStatus = canonicalSettlementStatusForProgression(next)
	tx.SettlementProgressionReasonCode = reasonCode
	tx.SettlementProgressionReason = reason
	tx.PartialSettlementHint = partialHint
	tx.DisputeReady = next == SettlementProgressionDisputeReady
	s.transactions[transactionReceiptID] = tx

	s.events[submissionReceiptID] = append(s.events[submissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: submissionReceiptID,
		Source:              "settlement_progression",
		Subtype:             string(next),
		Reason:              reason,
		Type:                EventSettlementUpdated,
	})
	if next == SettlementProgressionDisputeReady {
		s.events[submissionReceiptID] = append(s.events[submissionReceiptID], ReceiptEvent{
			SubmissionReceiptID: submissionReceiptID,
			Source:              "settlement_progression",
			Subtype:             string(next),
			Reason:              reason,
			Type:                EventDisputed,
		})
	}

	return cloneTransactionReceipt(tx), nil
}

func (s *Store) MarkSettlementSettled(ctx context.Context, req SettlementCloseoutRequest) (TransactionReceipt, error) {
	return s.markTransactionSettled(ctx, req, "settlement_execution", "settlement executed")
}

func (s *Store) MarkEscrowReleaseSettled(ctx context.Context, req SettlementCloseoutRequest) (TransactionReceipt, error) {
	return s.markTransactionSettled(ctx, req, "escrow_release", "escrow release executed")
}

func (s *Store) markTransactionSettled(_ context.Context, req SettlementCloseoutRequest, source string, progressionReason string) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionApprovedForSettlement {
		return TransactionReceipt{}, fmt.Errorf("%w: settlement must be approved-for-settlement before closeout", ErrInvalidSettlementProgressionState)
	}

	transaction.SettlementProgressionStatus = SettlementProgressionSettled
	transaction.CanonicalSettlementStatus = SettlementSettled
	transaction.SettlementProgressionReasonCode = SettlementProgressionReasonCodeApprove
	transaction.SettlementProgressionReason = progressionReason
	transaction.DisputeReady = false
	s.transactions[req.TransactionReceiptID] = transaction

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              source,
		Subtype:             "settled",
		Reason:              req.RuntimeReference,
		Type:                EventSettlementUpdated,
	})

	return cloneTransactionReceipt(transaction), nil
}

func (s *Store) RecordSettlementFailure(ctx context.Context, req SettlementFailureRequest) error {
	return s.recordTransactionFailure(ctx, req, "settlement_execution")
}

func (s *Store) RecordEscrowReleaseFailure(ctx context.Context, req SettlementFailureRequest) error {
	return s.recordTransactionFailure(ctx, req, "escrow_release")
}

func (s *Store) RecordEscrowRefundSuccess(_ context.Context, req EscrowRefundEvidenceRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowExecutionStatus != EscrowExecutionStatusFunded {
		return fmt.Errorf("%w: escrow must remain funded before recording escrow refund success", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionReviewNeeded {
		return fmt.Errorf("%w: settlement must remain review-needed before recording escrow refund success", ErrInvalidSettlementProgressionState)
	}

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "escrow_refund",
		Subtype:             "refunded",
		Reason:              req.RuntimeReference,
		Type:                EventSettlementUpdated,
	})

	return nil
}

func (s *Store) RecordEscrowDisputeHoldSuccess(_ context.Context, req EscrowDisputeHoldEvidenceRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowExecutionStatus != EscrowExecutionStatusFunded {
		return fmt.Errorf("%w: escrow must remain funded before recording dispute hold success", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionDisputeReady {
		return fmt.Errorf("%w: settlement must remain dispute-ready before recording dispute hold success", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowReference != req.EscrowReference {
		return fmt.Errorf("%w: escrow reference does not match transaction", ErrInvalidSettlementProgressionState)
	}

	reason := req.RuntimeReference
	if strings.TrimSpace(reason) == "" {
		reason = req.EscrowReference
	}
	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "dispute_hold",
		Subtype:             "held",
		Reason:              reason,
		Type:                EventSettlementUpdated,
	})

	return nil
}

func (s *Store) RecordEscrowRefundFailure(ctx context.Context, req SettlementFailureRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionReviewNeeded {
		return fmt.Errorf("%w: settlement must remain review-needed before recording escrow refund failure", ErrInvalidSettlementProgressionState)
	}

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "escrow_refund",
		Subtype:             "failed",
		Reason:              req.Reason,
		Type:                EventSettlementExecutionFailed,
	})

	return nil
}

func (s *Store) RecordEscrowDisputeHoldFailure(_ context.Context, req EscrowDisputeHoldFailureRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowExecutionStatus != EscrowExecutionStatusFunded {
		return fmt.Errorf("%w: escrow must remain funded before recording dispute hold failure", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionDisputeReady {
		return fmt.Errorf("%w: settlement must remain dispute-ready before recording dispute hold failure", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowReference != req.EscrowReference {
		return fmt.Errorf("%w: escrow reference does not match transaction", ErrInvalidSettlementProgressionState)
	}

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "dispute_hold",
		Subtype:             "failed",
		Reason:              req.Reason,
		Type:                EventSettlementExecutionFailed,
	})

	return nil
}

func (s *Store) ApplyEscrowAdjudication(_ context.Context, req EscrowAdjudicationRequest) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowExecutionStatus != EscrowExecutionStatusFunded {
		return TransactionReceipt{}, fmt.Errorf("%w: escrow must remain funded before adjudication", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionDisputeReady {
		return TransactionReceipt{}, fmt.Errorf("%w: settlement must remain dispute-ready before adjudication", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowReference != req.EscrowReference {
		return TransactionReceipt{}, fmt.Errorf("%w: escrow reference does not match transaction", ErrInvalidSettlementProgressionState)
	}
	if req.Outcome != EscrowAdjudicationRelease && req.Outcome != EscrowAdjudicationRefund {
		return TransactionReceipt{}, fmt.Errorf("%w: invalid escrow adjudication outcome", ErrInvalidSettlementProgressionState)
	}

	held := false
	for _, event := range s.events[req.SubmissionReceiptID] {
		if event.Source == "dispute_hold" && event.Subtype == "held" {
			held = true
			break
		}
	}
	if !held {
		return TransactionReceipt{}, fmt.Errorf("%w: dispute hold evidence is required before adjudication", ErrInvalidSettlementProgressionState)
	}

	transaction.EscrowAdjudication = req.Outcome
	s.transactions[req.TransactionReceiptID] = transaction

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = string(req.Outcome)
	}
	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "escrow_adjudication",
		Subtype:             string(req.Outcome),
		Reason:              reason,
		Type:                EventSettlementUpdated,
	})

	return cloneTransactionReceipt(transaction), nil
}

func (s *Store) RecordEscrowAdjudicationFailure(_ context.Context, req EscrowAdjudicationFailureRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.EscrowReference != "" && req.EscrowReference != "" && transaction.EscrowReference != req.EscrowReference {
		return fmt.Errorf("%w: escrow reference does not match transaction", ErrInvalidSettlementProgressionState)
	}

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "escrow_adjudication",
		Subtype:             "failed",
		Reason:              req.Reason,
		Type:                EventSettlementExecutionFailed,
	})

	return nil
}

func (s *Store) recordTransactionFailure(_ context.Context, req SettlementFailureRequest, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionApprovedForSettlement {
		return fmt.Errorf("%w: settlement must remain approved-for-settlement before recording failure", ErrInvalidSettlementProgressionState)
	}

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              source,
		Subtype:             "failed",
		Reason:              req.Reason,
		Type:                EventSettlementExecutionFailed,
	})

	return nil
}

func (s *Store) MarkPartialSettlementSettled(_ context.Context, req PartialSettlementCloseoutRequest) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionApprovedForSettlement {
		return TransactionReceipt{}, fmt.Errorf("%w: settlement must be approved-for-settlement before partial closeout", ErrInvalidSettlementProgressionState)
	}

	remainingHint, err := canonicalizePartialSettlementHint(req.RemainingAmount)
	if err != nil {
		return TransactionReceipt{}, err
	}

	transaction.SettlementProgressionStatus = SettlementProgressionPartiallySettled
	transaction.CanonicalSettlementStatus = SettlementPartiallySettled
	transaction.SettlementProgressionReasonCode = SettlementProgressionReasonCodeApprove
	transaction.SettlementProgressionReason = "partial settlement executed"
	transaction.PartialSettlementHint = remainingHint
	transaction.DisputeReady = false
	s.transactions[req.TransactionReceiptID] = transaction

	return cloneTransactionReceipt(transaction), nil
}

func (s *Store) MarkSettlementPartiallySettled(ctx context.Context, req SettlementPartialCloseoutRequest) (TransactionReceipt, error) {
	return s.MarkPartialSettlementSettled(ctx, req)
}

func (s *Store) RecordPartialSettlementSuccess(_ context.Context, req PartialSettlementExecutionEvidenceRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionPartiallySettled {
		return fmt.Errorf("%w: settlement must be partially-settled before recording success", ErrInvalidSettlementProgressionState)
	}

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "partial_settlement_execution",
		Subtype:             "partially-settled",
		Reason:              req.RuntimeReference,
		Type:                EventSettlementUpdated,
	})

	return nil
}

func (s *Store) RecordSettlementPartialSuccess(ctx context.Context, req SettlementPartialExecutionEvidenceRequest) error {
	return s.RecordPartialSettlementSuccess(ctx, req)
}

func (s *Store) RecordPartialSettlementFailure(_ context.Context, req PartialSettlementFailureRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[req.TransactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}
	submission, ok := s.submissions[req.SubmissionReceiptID]
	if !ok {
		return ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != req.TransactionReceiptID {
		return fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != req.SubmissionReceiptID {
		return fmt.Errorf("%w: submission is not current for transaction", ErrInvalidSettlementProgressionState)
	}
	if transaction.SettlementProgressionStatus != SettlementProgressionApprovedForSettlement {
		return fmt.Errorf("%w: settlement must remain approved-for-settlement before recording partial failure", ErrInvalidSettlementProgressionState)
	}

	s.events[req.SubmissionReceiptID] = append(s.events[req.SubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: req.SubmissionReceiptID,
		Source:              "partial_settlement_execution",
		Subtype:             "failed",
		Reason:              req.Reason,
		Type:                EventSettlementExecutionFailed,
	})

	return nil
}

func (s *Store) ApplyKnowledgeExchangeRuntimeProgression(_ context.Context, transactionReceiptID string, next KnowledgeExchangeRuntimeStatus, submissionReceiptID string) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}
	if submissionReceiptID != "" {
		submission, ok := s.submissions[submissionReceiptID]
		if !ok {
			return TransactionReceipt{}, ErrSubmissionReceiptNotFound
		}
		if submission.TransactionReceiptID != transactionReceiptID {
			return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
		}
	}
	if err := validateKnowledgeExchangeRuntimeTransition(tx.KnowledgeExchangeRuntimeStatus, next); err != nil {
		return TransactionReceipt{}, err
	}
	if submissionReceiptID != "" {
		tx.CurrentSubmissionReceiptID = submissionReceiptID
	}
	tx.KnowledgeExchangeRuntimeStatus = next
	s.transactions[transactionReceiptID] = tx

	return cloneTransactionReceipt(tx), nil
}

func (s *Store) ApplyUpfrontPaymentApproval(_ context.Context, transactionReceiptID, submissionReceiptID string, outcome paymentapproval.Outcome) (TransactionReceipt, error) {
	status, err := paymentApprovalStatusFromDecision(outcome.Decision)
	if err != nil {
		return TransactionReceipt{}, err
	}
	if err := validatePaymentApprovalStatus(status); err != nil {
		return TransactionReceipt{}, err
	}
	if submissionReceiptID == "" {
		return TransactionReceipt{}, fmt.Errorf("%w: submission_receipt_id is required", ErrSubmissionReceiptNotFound)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}

	submission, ok := s.submissions[submissionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != transactionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}

	transaction.CurrentPaymentApprovalStatus = status
	transaction.CanonicalDecision = string(outcome.Decision)
	transaction.CanonicalSettlementHint = string(outcome.SuggestedMode)
	s.transactions[transactionReceiptID] = transaction

	s.events[submissionReceiptID] = append(s.events[submissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: submissionReceiptID,
		Source:              "approval",
		Subtype:             "approval.upfront_payment",
		Type:                EventPaymentApproval,
	})

	return cloneTransactionReceipt(transaction), nil
}

func (s *Store) BindEscrowExecutionInput(_ context.Context, transactionReceiptID, submissionReceiptID string, input EscrowExecutionInput) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}

	submission, ok := s.submissions[submissionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != transactionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != submissionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission is not current for transaction", ErrInvalidEscrowExecutionState)
	}

	inputCopy := cloneEscrowExecutionInput(&input)
	transaction.EscrowExecutionStatus = EscrowExecutionStatusPending
	transaction.EscrowExecutionInput = inputCopy
	transaction.EscrowReference = ""
	s.transactions[transactionReceiptID] = transaction

	return cloneTransactionReceipt(transaction), nil
}

func (s *Store) ApplyEscrowExecutionProgress(_ context.Context, transactionReceiptID, submissionReceiptID string, status EscrowExecutionStatus, escrowReference string, eventType EventType, reason string) (TransactionReceipt, error) {
	if err := validateEventType(eventType); err != nil {
		return TransactionReceipt{}, err
	}
	if err := validateEscrowExecutionStatus(status); err != nil {
		return TransactionReceipt{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}

	submission, ok := s.submissions[submissionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if submission.TransactionReceiptID != transactionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission does not belong to transaction", ErrSubmissionReceiptNotFound)
	}
	if transaction.CurrentSubmissionReceiptID != submissionReceiptID {
		return TransactionReceipt{}, fmt.Errorf("%w: submission is not current for transaction", ErrInvalidEscrowExecutionState)
	}
	if transaction.EscrowExecutionInput == nil {
		return TransactionReceipt{}, fmt.Errorf("%w: escrow execution input is required", ErrInvalidEscrowExecutionState)
	}
	if err := validateEscrowExecutionTransition(transaction.EscrowExecutionStatus, status); err != nil {
		return TransactionReceipt{}, err
	}
	if status == EscrowExecutionStatusFunded {
		if escrowReference == "" {
			return TransactionReceipt{}, fmt.Errorf("%w: escrow reference is required for funded progress", ErrInvalidEscrowExecutionState)
		}
		if transaction.EscrowExecutionStatus != EscrowExecutionStatusCreated {
			return TransactionReceipt{}, fmt.Errorf("%w: funded progress requires created state", ErrInvalidEscrowExecutionState)
		}
	}

	expectedEventType, err := escrowExecutionEventTypeForStatus(status)
	if err != nil {
		return TransactionReceipt{}, err
	}
	if eventType != expectedEventType {
		return TransactionReceipt{}, fmt.Errorf("%w: status %q does not match event type %q", ErrInvalidEscrowExecutionState, status, eventType)
	}
	transaction.EscrowExecutionStatus = status
	if escrowReference != "" {
		transaction.EscrowReference = escrowReference
	}
	s.transactions[transactionReceiptID] = transaction

	s.events[submissionReceiptID] = append(s.events[submissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: submissionReceiptID,
		Source:              "escrow_execution",
		Subtype:             string(status),
		Reason:              reason,
		Type:                eventType,
	})

	return cloneTransactionReceipt(transaction), nil
}

func (s *Store) AppendReceiptEvent(_ context.Context, submissionReceiptID string, eventType EventType) error {
	if err := validateEventType(eventType); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.submissions[submissionReceiptID]; !ok {
		return ErrSubmissionReceiptNotFound
	}

	s.events[submissionReceiptID] = append(s.events[submissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: submissionReceiptID,
		Source:              "manual",
		Subtype:             string(eventType),
		Type:                eventType,
	})

	return nil
}

func (s *Store) AppendPaymentExecutionAuthorized(ctx context.Context, submissionReceiptID string) error {
	return s.appendPaymentExecutionEvent(ctx, submissionReceiptID, EventPaymentExecutionAuthorized, "", "authorized")
}

func (s *Store) AppendPaymentExecutionDenied(ctx context.Context, submissionReceiptID, reason string) error {
	return s.appendPaymentExecutionEvent(ctx, submissionReceiptID, EventPaymentExecutionDenied, reason, "denied")
}

func (s *Store) GetSubmissionReceipt(_ context.Context, submissionReceiptID string) (SubmissionReceipt, []ReceiptEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	submission, ok := s.submissions[submissionReceiptID]
	if !ok {
		return SubmissionReceipt{}, nil, ErrSubmissionReceiptNotFound
	}

	events := append([]ReceiptEvent(nil), s.events[submissionReceiptID]...)

	return submission, events, nil
}

func (s *Store) GetTransactionReceipt(_ context.Context, transactionReceiptID string) (TransactionReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}

	return cloneTransactionReceipt(transaction), nil
}

func (s *Store) appendPaymentExecutionEvent(_ context.Context, submissionReceiptID string, eventType EventType, reason string, subtype string) error {
	if err := validateEventType(eventType); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.submissions[submissionReceiptID]; !ok {
		return ErrSubmissionReceiptNotFound
	}

	s.events[submissionReceiptID] = append(s.events[submissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: submissionReceiptID,
		Source:              "payment_execution",
		Subtype:             subtype,
		Reason:              reason,
		Type:                eventType,
	})

	return nil
}

func validateCreateSubmissionInput(in CreateSubmissionInput) error {
	switch {
	case in.TransactionID == "":
		return fmt.Errorf("%w: transaction_id is required", ErrInvalidSubmissionInput)
	case in.ArtifactLabel == "":
		return fmt.Errorf("%w: artifact_label is required", ErrInvalidSubmissionInput)
	case in.PayloadHash == "":
		return fmt.Errorf("%w: payload_hash is required", ErrInvalidSubmissionInput)
	case in.SourceLineageDigest == "":
		return fmt.Errorf("%w: source_lineage_digest is required", ErrInvalidSubmissionInput)
	default:
		return nil
	}
}

func validateEventType(eventType EventType) error {
	switch eventType {
	case EventDraftExportability,
		EventFinalExportability,
		EventApprovalRequested,
		EventApprovalResolved,
		EventPaymentApproval,
		EventPaymentExecutionAuthorized,
		EventPaymentExecutionDenied,
		EventEscrowExecutionStarted,
		EventEscrowExecutionCreated,
		EventEscrowExecutionFunded,
		EventEscrowExecutionFailed,
		EventSettlementUpdated,
		EventSettlementExecutionFailed,
		EventEscalated,
		EventDisputed:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidReceiptEventType, eventType)
	}
}

func validatePaymentApprovalStatus(status PaymentApprovalStatus) error {
	switch status {
	case PaymentApprovalPending, PaymentApprovalApproved, PaymentApprovalRejected, PaymentApprovalEscalated:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidPaymentApprovalStatus, status)
	}
}

func validateEscrowExecutionStatus(status EscrowExecutionStatus) error {
	switch status {
	case EscrowExecutionStatusPending, EscrowExecutionStatusCreated, EscrowExecutionStatusFunded, EscrowExecutionStatusFailed:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidEscrowExecutionStatus, status)
	}
}

func escrowExecutionEventTypeForStatus(status EscrowExecutionStatus) (EventType, error) {
	switch status {
	case EscrowExecutionStatusPending:
		return EventEscrowExecutionStarted, nil
	case EscrowExecutionStatusCreated:
		return EventEscrowExecutionCreated, nil
	case EscrowExecutionStatusFunded:
		return EventEscrowExecutionFunded, nil
	case EscrowExecutionStatusFailed:
		return EventEscrowExecutionFailed, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidEscrowExecutionStatus, status)
	}
}

func validateEscrowExecutionTransition(current, next EscrowExecutionStatus) error {
	switch current {
	case "":
		if next == EscrowExecutionStatusCreated || next == EscrowExecutionStatusFunded || next == EscrowExecutionStatusFailed || next == EscrowExecutionStatusPending {
			return nil
		}
	case EscrowExecutionStatusPending:
		if next == EscrowExecutionStatusPending || next == EscrowExecutionStatusCreated || next == EscrowExecutionStatusFailed {
			return nil
		}
	case EscrowExecutionStatusCreated:
		if next == EscrowExecutionStatusFunded || next == EscrowExecutionStatusFailed {
			return nil
		}
	case EscrowExecutionStatusFunded:
		if next == EscrowExecutionStatusFailed {
			return nil
		}
	case EscrowExecutionStatusFailed:
	}

	return fmt.Errorf("%w: illegal transition from %q to %q", ErrInvalidEscrowExecutionState, current, next)
}

func validateKnowledgeExchangeRuntimeTransition(current, next KnowledgeExchangeRuntimeStatus) error {
	switch current {
	case "":
		if next == RuntimeStatusOpened {
			return nil
		}
	case RuntimeStatusOpened:
		if next == RuntimeStatusExportabilityAdvisory || next == RuntimeStatusPaymentApproved {
			return nil
		}
	case RuntimeStatusExportabilityAdvisory:
		if next == RuntimeStatusPaymentApproved {
			return nil
		}
	case RuntimeStatusPaymentApproved:
		if next == RuntimeStatusPaymentAuthorized || next == RuntimeStatusEscrowFunded {
			return nil
		}
	case RuntimeStatusPaymentAuthorized:
		if next == RuntimeStatusWorkStarted {
			return nil
		}
	case RuntimeStatusEscrowFunded:
		if next == RuntimeStatusWorkStarted {
			return nil
		}
	case RuntimeStatusWorkStarted:
		if next == RuntimeStatusSubmissionReceived {
			return nil
		}
	case RuntimeStatusSubmissionReceived:
		if next == RuntimeStatusReleaseApproved || next == RuntimeStatusRevisionRequested || next == RuntimeStatusEscalated || next == RuntimeStatusDisputeReady {
			return nil
		}
	case RuntimeStatusRevisionRequested:
		if next == RuntimeStatusSubmissionReceived {
			return nil
		}
	}

	return fmt.Errorf("%w: %q -> %q", ErrInvalidKnowledgeExchangeRuntimeState, current, next)
}

func validateSettlementProgressionTransition(current, next SettlementProgressionStatus) error {
	switch current {
	case "":
		if next == SettlementProgressionPending || next == SettlementProgressionApprovedForSettlement || next == SettlementProgressionReviewNeeded {
			return nil
		}
	case SettlementProgressionPending:
		if next == SettlementProgressionApprovedForSettlement || next == SettlementProgressionReviewNeeded {
			return nil
		}
	case SettlementProgressionApprovedForSettlement:
		if next == SettlementProgressionInProgress || next == SettlementProgressionSettled || next == SettlementProgressionPartiallySettled {
			return nil
		}
	case SettlementProgressionReviewNeeded:
		if next == SettlementProgressionReviewNeeded || next == SettlementProgressionApprovedForSettlement || next == SettlementProgressionDisputeReady {
			return nil
		}
	case SettlementProgressionInProgress:
		if next == SettlementProgressionSettled || next == SettlementProgressionPartiallySettled || next == SettlementProgressionReviewNeeded {
			return nil
		}
	case SettlementProgressionPartiallySettled:
		if next == SettlementProgressionSettled || next == SettlementProgressionReviewNeeded || next == SettlementProgressionDisputeReady {
			return nil
		}
	}

	return fmt.Errorf("%w: %q -> %q", ErrInvalidSettlementProgressionState, current, next)
}

func validateSettlementProgressionReasonCode(next SettlementProgressionStatus, reasonCode SettlementProgressionReasonCode) error {
	switch next {
	case SettlementProgressionApprovedForSettlement,
		SettlementProgressionInProgress,
		SettlementProgressionPartiallySettled,
		SettlementProgressionSettled:
		if reasonCode == SettlementProgressionReasonCodeApprove {
			return nil
		}
		return fmt.Errorf("%w: %q requires approve reason code, got %q", ErrInvalidSettlementProgressionState, next, reasonCode)
	case SettlementProgressionReviewNeeded:
		fallthrough
	case SettlementProgressionDisputeReady:
		switch reasonCode {
		case SettlementProgressionReasonCodeReject,
			SettlementProgressionReasonCodeRequestRevision,
			SettlementProgressionReasonCodeEscalate:
			return nil
		default:
			return fmt.Errorf("%w: %q requires reject, request-revision, or escalate reason code, got %q", ErrInvalidSettlementProgressionState, next, reasonCode)
		}
	default:
		return nil
	}
}

func canonicalSettlementStatusForProgression(progress SettlementProgressionStatus) SettlementStatus {
	switch progress {
	case SettlementProgressionPartiallySettled:
		return SettlementPartiallySettled
	case SettlementProgressionSettled:
		return SettlementSettled
	case SettlementProgressionDisputeReady:
		return SettlementDisputed
	default:
		return SettlementPending
	}
}

func validateCanonicalOpenInputConflict(existing TransactionReceipt, in OpenTransactionInput) error {
	switch {
	case existing.Counterparty != "" && existing.Counterparty != in.Counterparty:
		return fmt.Errorf("%w: counterparty conflicts with existing transaction baseline", ErrInvalidSubmissionInput)
	case existing.RequestedScope != "" && existing.RequestedScope != in.RequestedScope:
		return fmt.Errorf("%w: requested_scope conflicts with existing transaction baseline", ErrInvalidSubmissionInput)
	case existing.PriceContext != "" && existing.PriceContext != in.PriceContext:
		return fmt.Errorf("%w: price_context conflicts with existing transaction baseline", ErrInvalidSubmissionInput)
	case existing.TrustContext != "" && existing.TrustContext != in.TrustContext:
		return fmt.Errorf("%w: trust_context conflicts with existing transaction baseline", ErrInvalidSubmissionInput)
	default:
		return nil
	}
}

func cloneEscrowExecutionInput(input *EscrowExecutionInput) *EscrowExecutionInput {
	if input == nil {
		return nil
	}

	cloned := *input
	if len(input.Milestones) > 0 {
		cloned.Milestones = append([]EscrowMilestoneInput(nil), input.Milestones...)
	}

	return &cloned
}

func cloneTransactionReceipt(transaction TransactionReceipt) TransactionReceipt {
	transaction.EscrowExecutionInput = cloneEscrowExecutionInput(transaction.EscrowExecutionInput)
	return transaction
}

func canonicalizePartialSettlementHint(remainingAmount string) (string, error) {
	parsed, err := finance.ParseUSDC(strings.TrimSpace(remainingAmount))
	if err != nil {
		return "", err
	}
	if parsed.Sign() <= 0 {
		return "", fmt.Errorf("%w: remaining amount must be positive", ErrInvalidSettlementProgressionState)
	}

	return "settle:" + finance.FormatUSDC(parsed) + "-usdc", nil
}

func paymentApprovalStatusFromDecision(decision paymentapproval.Decision) (PaymentApprovalStatus, error) {
	switch decision {
	case paymentapproval.DecisionApprove:
		return PaymentApprovalApproved, nil
	case paymentapproval.DecisionReject:
		return PaymentApprovalRejected, nil
	case paymentapproval.DecisionEscalate:
		return PaymentApprovalEscalated, nil
	default:
		return PaymentApprovalPending, fmt.Errorf("%w: %q", ErrInvalidPaymentApprovalStatus, decision)
	}
}
