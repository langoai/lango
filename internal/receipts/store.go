package receipts

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
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
	s.transactions[txReceiptID] = transaction

	return submission, transaction, nil
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

	return transaction, nil
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

func (s *Store) AppendPaymentExecutionAuthorized(ctx context.Context, transactionReceiptID string) error {
	return s.appendPaymentExecutionEvent(ctx, transactionReceiptID, EventPaymentExecutionAuthorized, "", "authorized")
}

func (s *Store) AppendPaymentExecutionDenied(ctx context.Context, transactionReceiptID, reason string) error {
	return s.appendPaymentExecutionEvent(ctx, transactionReceiptID, EventPaymentExecutionDenied, reason, "denied")
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

	return transaction, nil
}

func (s *Store) appendPaymentExecutionEvent(_ context.Context, transactionReceiptID string, eventType EventType, reason string, subtype string) error {
	if err := validateEventType(eventType); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return ErrTransactionReceiptNotFound
	}

	submissionReceiptID := transaction.CurrentSubmissionReceiptID
	if submissionReceiptID == "" {
		return ErrSubmissionReceiptNotFound
	}
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
		EventSettlementUpdated,
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
