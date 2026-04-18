package librarian

import (
	"context"
	"fmt"
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

type inquiryPayload struct {
	Question string `json:"question"`
	Context  string `json:"context,omitempty"`
	Answer   string `json:"answer,omitempty"`
}

type inquiryProjection struct {
	Question string
	Context  string
	Answer   string
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
		inq, err := s.entToInquiry(e)
		if err != nil {
			return nil, fmt.Errorf("decode inquiry %q: %w", e.ID, err)
		}
		result = append(result, inq)
	}
	return result, nil
}

// ResolveInquiry marks an inquiry as resolved with the user's answer and optional knowledge key.
func (s *InquiryStore) ResolveInquiry(ctx context.Context, id uuid.UUID, answer, knowledgeKey string) error {
	builder := s.client.Inquiry.UpdateOneID(id).
		SetStatus(inquiry.StatusResolved).
		SetResolvedAt(time.Now())
	if answer != "" {
		payload, err := s.loadInquiryPayload(ctx, id)
		if err != nil {
			return fmt.Errorf("load inquiry payload: %w", err)
		}
		payload.Answer = answer
		projection, ciphertext, nonce, keyVersion, err := s.prepareInquiryPayload(payload)
		if err != nil {
			return fmt.Errorf("prepare inquiry answer payload: %w", err)
		}
		builder.SetQuestion(projection.Question)
		if projection.Context != "" {
			builder.SetContext(projection.Context)
		} else {
			builder.ClearContext()
		}
		builder.SetAnswer(projection.Answer)
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
func (s *InquiryStore) entToInquiry(e *ent.Inquiry) (Inquiry, error) {
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
	if s.payloads == nil || e.PayloadCiphertext == nil || e.PayloadNonce == nil || e.PayloadKeyVersion == nil {
		return inq, nil
	}
	payload, err := security.UnprotectJSONBundle[inquiryPayload](s.payloads, *e.PayloadCiphertext, *e.PayloadNonce, *e.PayloadKeyVersion)
	if err != nil {
		return Inquiry{}, fmt.Errorf("decrypt inquiry payload: %w", err)
	}
	inq.Question = payload.Question
	inq.Context = payload.Context
	inq.Answer = payload.Answer
	return inq, nil
}

func (s *InquiryStore) prepareInquiryWrite(question, contextText string) (string, string, []byte, []byte, int, error) {
	if s.payloads == nil {
		return question, contextText, nil, nil, 0, nil
	}
	projection, ciphertext, nonce, keyVersion, err := s.prepareInquiryPayload(inquiryPayload{Question: question, Context: contextText})
	if err != nil {
		return "", "", nil, nil, 0, err
	}
	return projection.Question, projection.Context, ciphertext, nonce, keyVersion, nil
}

func (s *InquiryStore) prepareInquiryPayload(payload inquiryPayload) (inquiryProjection, []byte, []byte, int, error) {
	projection := inquiryProjection{
		Question: security.RedactedProjection(payload.Question, 512),
		Context:  security.RedactedProjection(payload.Context, 512),
		Answer:   security.RedactedProjection(payload.Answer, 512),
	}
	if s.payloads == nil {
		return projection, nil, nil, 0, nil
	}
	ciphertext, nonce, keyVersion, err := security.ProtectJSONBundle(s.payloads, payload)
	if err != nil {
		return inquiryProjection{}, nil, nil, 0, err
	}
	return projection, ciphertext, nonce, keyVersion, nil
}

func (s *InquiryStore) loadInquiryPayload(ctx context.Context, id uuid.UUID) (inquiryPayload, error) {
	e, err := s.client.Inquiry.Get(ctx, id)
	if err != nil {
		return inquiryPayload{}, err
	}
	if s.payloads == nil || e.PayloadCiphertext == nil || e.PayloadNonce == nil || e.PayloadKeyVersion == nil {
		payload := inquiryPayload{
			Question: e.Question,
		}
		if e.Context != nil {
			payload.Context = *e.Context
		}
		if e.Answer != nil {
			payload.Answer = *e.Answer
		}
		return payload, nil
	}
	return security.UnprotectJSONBundle[inquiryPayload](s.payloads, *e.PayloadCiphertext, *e.PayloadNonce, *e.PayloadKeyVersion)
}
