package librarian

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/inquiry"
	"github.com/langoai/lango/internal/security"
)

// InquiryStore provides CRUD operations for knowledge inquiries.
type InquiryStore struct {
	client   *ent.Client
	logger   *zap.SugaredLogger
	payloads security.PayloadProtector
}

// NewInquiryStore creates a new inquiry store.
func NewInquiryStore(client *ent.Client, logger *zap.SugaredLogger) *InquiryStore {
	return &InquiryStore{
		client: client,
		logger: logger,
	}
}

func (s *InquiryStore) SetPayloadProtector(protector security.PayloadProtector) {
	s.payloads = protector
}

// SaveInquiry persists a new inquiry to the database.
func (s *InquiryStore) SaveInquiry(ctx context.Context, inq Inquiry) error {
	question, contextText, ciphertext, nonce, keyVersion, err := s.prepareInquiryWrite(inq.Question, inq.Context)
	if err != nil {
		return fmt.Errorf("prepare inquiry payload: %w", err)
	}
	builder := s.client.Inquiry.Create().
		SetSessionKey(inq.SessionKey).
		SetTopic(inq.Topic).
		SetQuestion(question).
		SetPriority(inquiry.Priority(inq.Priority)).
		SetStatus(inquiry.StatusPending)
	if ciphertext != nil {
		builder.SetPayloadCiphertext(ciphertext)
		builder.SetPayloadNonce(nonce)
		builder.SetPayloadKeyVersion(keyVersion)
	}

	if contextText != "" {
		builder.SetContext(contextText)
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
		SetResolvedAt(time.Now())
	if answer != "" {
		projection, ciphertext, nonce, keyVersion, err := s.prepareProtectedText(answer)
		if err != nil {
			return fmt.Errorf("prepare inquiry answer payload: %w", err)
		}
		builder.SetAnswer(projection)
		if ciphertext != nil {
			builder.SetPayloadCiphertext(ciphertext)
			builder.SetPayloadNonce(nonce)
			builder.SetPayloadKeyVersion(keyVersion)
		}
	}

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

func (s *InquiryStore) prepareInquiryWrite(question, contextText string) (string, string, []byte, []byte, int, error) {
	if s.payloads == nil {
		return question, contextText, nil, nil, 0, nil
	}
	plaintext := question
	if contextText != "" {
		plaintext += "\n" + contextText
	}
	ciphertext, nonce, keyVersion, err := s.payloads.EncryptPayload([]byte(plaintext))
	if err != nil {
		return "", "", nil, nil, 0, err
	}
	return redactInquiryProjection(question), redactInquiryProjection(contextText), ciphertext, nonce, keyVersion, nil
}

func (s *InquiryStore) prepareProtectedText(text string) (string, []byte, []byte, int, error) {
	if s.payloads == nil || text == "" {
		return text, nil, nil, 0, nil
	}
	ciphertext, nonce, keyVersion, err := s.payloads.EncryptPayload([]byte(text))
	if err != nil {
		return "", nil, nil, 0, err
	}
	return redactInquiryProjection(text), ciphertext, nonce, keyVersion, nil
}

func redactInquiryProjection(text string) string {
	return strings.TrimSpace(text)
}
