package memory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/observation"
	"github.com/langowarny/lango/internal/ent/reflection"
)

// Store provides CRUD operations for observations and reflections.
type Store struct {
	client *ent.Client
	logger *zap.SugaredLogger
}

// NewStore creates a new observational memory store.
func NewStore(client *ent.Client, logger *zap.SugaredLogger) *Store {
	return &Store{
		client: client,
		logger: logger,
	}
}

// SaveObservation persists an observation to the database.
func (s *Store) SaveObservation(ctx context.Context, obs Observation) error {
	builder := s.client.Observation.Create().
		SetSessionKey(obs.SessionKey).
		SetContent(obs.Content).
		SetTokenCount(obs.TokenCount).
		SetSourceStartIndex(obs.SourceStartIndex).
		SetSourceEndIndex(obs.SourceEndIndex)

	if obs.ID != uuid.Nil {
		builder.SetID(obs.ID)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("save observation: %w", err)
	}
	return nil
}

// ListObservations returns observations for a session ordered by created_at ascending.
func (s *Store) ListObservations(ctx context.Context, sessionKey string) ([]Observation, error) {
	entries, err := s.client.Observation.Query().
		Where(observation.SessionKey(sessionKey)).
		Order(observation.ByCreatedAt()).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("list observations: %w", err)
	}

	result := make([]Observation, 0, len(entries))
	for _, e := range entries {
		result = append(result, Observation{
			ID:               e.ID,
			SessionKey:       e.SessionKey,
			Content:          e.Content,
			TokenCount:       e.TokenCount,
			SourceStartIndex: e.SourceStartIndex,
			SourceEndIndex:   e.SourceEndIndex,
			CreatedAt:        e.CreatedAt,
		})
	}
	return result, nil
}

// DeleteObservations deletes observations by their IDs.
func (s *Store) DeleteObservations(ctx context.Context, ids []uuid.UUID) error {
	_, err := s.client.Observation.Delete().
		Where(observation.IDIn(ids...)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete observations: %w", err)
	}
	return nil
}

// DeleteObservationsBySession deletes all observations for a session.
func (s *Store) DeleteObservationsBySession(ctx context.Context, sessionKey string) error {
	_, err := s.client.Observation.Delete().
		Where(observation.SessionKey(sessionKey)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete observations by session: %w", err)
	}
	return nil
}

// SaveReflection persists a reflection to the database.
func (s *Store) SaveReflection(ctx context.Context, ref Reflection) error {
	builder := s.client.Reflection.Create().
		SetSessionKey(ref.SessionKey).
		SetContent(ref.Content).
		SetTokenCount(ref.TokenCount).
		SetGeneration(ref.Generation)

	if ref.ID != uuid.Nil {
		builder.SetID(ref.ID)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("save reflection: %w", err)
	}
	return nil
}

// DeleteReflections deletes reflections by their IDs.
func (s *Store) DeleteReflections(ctx context.Context, ids []uuid.UUID) error {
	_, err := s.client.Reflection.Delete().
		Where(reflection.IDIn(ids...)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete reflections: %w", err)
	}
	return nil
}

// ListReflections returns reflections for a session ordered by created_at ascending.
func (s *Store) ListReflections(ctx context.Context, sessionKey string) ([]Reflection, error) {
	entries, err := s.client.Reflection.Query().
		Where(reflection.SessionKey(sessionKey)).
		Order(reflection.ByCreatedAt()).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("list reflections: %w", err)
	}

	result := make([]Reflection, 0, len(entries))
	for _, e := range entries {
		result = append(result, Reflection{
			ID:         e.ID,
			SessionKey: e.SessionKey,
			Content:    e.Content,
			TokenCount: e.TokenCount,
			Generation: e.Generation,
			CreatedAt:  e.CreatedAt,
		})
	}
	return result, nil
}
