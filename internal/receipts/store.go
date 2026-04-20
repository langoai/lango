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
	s.mu.Lock()
	defer s.mu.Unlock()

	txReceiptID, ok := s.txByExternalID[in.TransactionID]
	if !ok {
		txReceiptID = uuid.NewString()
		s.txByExternalID[in.TransactionID] = txReceiptID
		s.transactions[txReceiptID] = TransactionReceipt{
			TransactionReceiptID:      txReceiptID,
			TransactionID:             in.TransactionID,
			CanonicalApprovalStatus:   ApprovalPending,
			CanonicalSettlementStatus: SettlementPending,
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

func (s *Store) AppendReceiptEvent(_ context.Context, submissionReceiptID string, eventType EventType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	submission, ok := s.submissions[submissionReceiptID]
	if !ok {
		return fmt.Errorf("submission receipt not found")
	}

	s.submissions[submissionReceiptID] = submission
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
		return SubmissionReceipt{}, nil, fmt.Errorf("submission receipt not found")
	}

	events := append([]ReceiptEvent(nil), s.events[submissionReceiptID]...)

	return submission, events, nil
}
