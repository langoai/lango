package receipts

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type CreateSubmissionInput struct {
	TransactionID       string
	ArtifactLabel       string
	PayloadHash         string
	SourceLineageDigest string
}

type ReceiptEvent struct {
	SubmissionReceiptID string
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

func (s *Store) ApplyUpfrontPaymentApproval(_ context.Context, transactionReceiptID string, status PaymentApprovalStatus, canonicalDecision string, canonicalSettlementHint string) (TransactionReceipt, error) {
	if err := validatePaymentApprovalStatus(status); err != nil {
		return TransactionReceipt{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	transaction, ok := s.transactions[transactionReceiptID]
	if !ok {
		return TransactionReceipt{}, ErrTransactionReceiptNotFound
	}

	if transaction.CurrentSubmissionReceiptID == "" {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}
	if _, ok := s.submissions[transaction.CurrentSubmissionReceiptID]; !ok {
		return TransactionReceipt{}, ErrSubmissionReceiptNotFound
	}

	transaction.CurrentPaymentApprovalStatus = status
	transaction.CanonicalPaymentApprovalDecision = canonicalDecision
	transaction.CanonicalPaymentSettlementHint = canonicalSettlementHint
	s.transactions[transactionReceiptID] = transaction

	s.events[transaction.CurrentSubmissionReceiptID] = append(s.events[transaction.CurrentSubmissionReceiptID], ReceiptEvent{
		SubmissionReceiptID: transaction.CurrentSubmissionReceiptID,
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
		Type:                eventType,
	})

	return nil
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
