package librarian

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/inquiry"
)

// InquiryStore provides CRUD operations for knowledge inquiries.
type InquiryStore struct {
	client *ent.Client
	logger *zap.SugaredLogger
}

// NewInquiryStore creates a new inquiry store.
func NewInquiryStore(client *ent.Client, logger *zap.SugaredLogger) *InquiryStore {
	return &InquiryStore{
		client: client,
		logger: logger,
	}
}

// SaveInquiry persists a new inquiry to the database.
func (s *InquiryStore) SaveInquiry(ctx context.Context, inq Inquiry) error {
	builder := s.client.Inquiry.Create().
		SetSessionKey(inq.SessionKey).
		SetTopic(inq.Topic).
		SetQuestion(inq.Question).
		SetPriority(inquiry.Priority(inq.Priority)).
		SetStatus(inquiry.StatusPending)

	if inq.Context != "" {
		builder.SetContext(inq.Context)
	}
	if inq.SourceObservationID != "" {
		builder.SetSourceObservationID(inq.SourceObservationID)
	}
	if inq.ID != uuid.Nil {
		builder.SetID(inq.ID)
	}

	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("save inquiry: %w", err)
	}
	return nil
}

// ListPendingInquiries returns pending inquiries for a session, ordered by priority and creation time.
func (s *InquiryStore) ListPendingInquiries(ctx context.Context, sessionKey string, limit int) ([]Inquiry, error) {
	entries, err := s.client.Inquiry.Query().
		Where(
			inquiry.SessionKey(sessionKey),
			inquiry.StatusEQ(inquiry.StatusPending),
		).
		Order(inquiry.ByCreatedAt()).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending inquiries: %w", err)
	}

	result := make([]Inquiry, 0, len(entries))
	for _, e := range entries {
		result = append(result, entToInquiry(e))
	}
	return result, nil
}

// ResolveInquiry marks an inquiry as resolved with the user's answer and optional knowledge key.
func (s *InquiryStore) ResolveInquiry(ctx context.Context, id uuid.UUID, answer, knowledgeKey string) error {
	builder := s.client.Inquiry.UpdateOneID(id).
		SetStatus(inquiry.StatusResolved).
		SetAnswer(answer).
		SetResolvedAt(time.Now())

	if knowledgeKey != "" {
		builder.SetKnowledgeKey(knowledgeKey)
	}

	if _, err := builder.Save(ctx); err != nil {
		return fmt.Errorf("resolve inquiry: %w", err)
	}
	return nil
}

// DismissInquiry marks an inquiry as dismissed.
func (s *InquiryStore) DismissInquiry(ctx context.Context, id uuid.UUID) error {
	if _, err := s.client.Inquiry.UpdateOneID(id).
		SetStatus(inquiry.StatusDismissed).
		SetResolvedAt(time.Now()).
		Save(ctx); err != nil {
		return fmt.Errorf("dismiss inquiry: %w", err)
	}
	return nil
}

// CountPendingBySession returns the number of pending inquiries for a session.
func (s *InquiryStore) CountPendingBySession(ctx context.Context, sessionKey string) (int, error) {
	count, err := s.client.Inquiry.Query().
		Where(
			inquiry.SessionKey(sessionKey),
			inquiry.StatusEQ(inquiry.StatusPending),
		).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count pending inquiries: %w", err)
	}
	return count, nil
}

// entToInquiry converts an ent inquiry entity to the domain type.
func entToInquiry(e *ent.Inquiry) Inquiry {
	inq := Inquiry{
		ID:         e.ID,
		SessionKey: e.SessionKey,
		Topic:      e.Topic,
		Question:   e.Question,
		Priority:   string(e.Priority),
		Status:     string(e.Status),
		CreatedAt:  e.CreatedAt,
		ResolvedAt: e.ResolvedAt,
	}
	if e.Context != nil {
		inq.Context = *e.Context
	}
	if e.Answer != nil {
		inq.Answer = *e.Answer
	}
	if e.KnowledgeKey != nil {
		inq.KnowledgeKey = *e.KnowledgeKey
	}
	if e.SourceObservationID != nil {
		inq.SourceObservationID = *e.SourceObservationID
	}
	return inq
}
