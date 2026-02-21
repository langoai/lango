package memory

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/observation"
	"github.com/langowarny/lango/internal/ent/reflection"
	"github.com/langowarny/lango/internal/types"
)

// Store provides CRUD operations for observations and reflections.
type Store struct {
	client     *ent.Client
	logger     *zap.SugaredLogger
	onEmbed    types.EmbedCallback
	onGraph    types.ContentCallback
	graphHooks *GraphHooks

	// lastObsMu protects lastObsIDs for concurrent SaveObservation calls.
	lastObsMu  sync.Mutex
	lastObsIDs map[string]string // session_key → last observation ID
}

// NewStore creates a new observational memory store.
func NewStore(client *ent.Client, logger *zap.SugaredLogger) *Store {
	return &Store{
		client:     client,
		logger:     logger,
		lastObsIDs: make(map[string]string),
	}
}

// SetEmbedCallback sets the optional embedding hook.
func (s *Store) SetEmbedCallback(cb types.EmbedCallback) {
	s.onEmbed = cb
}

// SetGraphCallback sets the optional graph relationship hook.
func (s *Store) SetGraphCallback(cb types.ContentCallback) {
	s.onGraph = cb
}

// SetGraphHooks sets the graph hooks for temporal/session triple generation.
func (s *Store) SetGraphHooks(hooks *GraphHooks) {
	s.graphHooks = hooks
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

	saved, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("save observation: %w", err)
	}

	savedID := saved.ID.String()

	meta := map[string]string{"session_key": obs.SessionKey}
	if s.onEmbed != nil {
		s.onEmbed(savedID, "observation", obs.Content, meta)
	}
	if s.onGraph != nil {
		s.onGraph(savedID, "observation", obs.Content, meta)
	}

	// Graph hooks: temporal ordering and session membership.
	if s.graphHooks != nil {
		s.lastObsMu.Lock()
		previousID := s.lastObsIDs[obs.SessionKey]
		s.lastObsIDs[obs.SessionKey] = savedID
		s.lastObsMu.Unlock()

		obs.ID = saved.ID
		s.graphHooks.OnObservation(obs, previousID)
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

// GetObservation retrieves a single observation by its ID.
func (s *Store) GetObservation(ctx context.Context, id uuid.UUID) (*Observation, error) {
	e, err := s.client.Observation.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get observation: %w", err)
	}
	return &Observation{
		ID:               e.ID,
		SessionKey:       e.SessionKey,
		Content:          e.Content,
		TokenCount:       e.TokenCount,
		SourceStartIndex: e.SourceStartIndex,
		SourceEndIndex:   e.SourceEndIndex,
		CreatedAt:        e.CreatedAt,
	}, nil
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

	saved, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("save reflection: %w", err)
	}

	savedID := saved.ID.String()

	meta := map[string]string{"session_key": ref.SessionKey}
	if s.onEmbed != nil {
		s.onEmbed(savedID, "reflection", ref.Content, meta)
	}
	if s.onGraph != nil {
		s.onGraph(savedID, "reflection", ref.Content, meta)
	}

	// Graph hooks: reflection-observation links.
	if s.graphHooks != nil {
		ref.ID = saved.ID
		// Collect recent observation IDs for this session.
		var obsIDs []string
		observations, listErr := s.ListObservations(ctx, ref.SessionKey)
		if listErr == nil {
			for _, obs := range observations {
				obsIDs = append(obsIDs, obs.ID.String())
			}
		}
		s.graphHooks.OnReflection(ref, obsIDs)
	}

	return nil
}

// GetReflection retrieves a single reflection by its ID.
func (s *Store) GetReflection(ctx context.Context, id uuid.UUID) (*Reflection, error) {
	e, err := s.client.Reflection.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get reflection: %w", err)
	}
	return &Reflection{
		ID:         e.ID,
		SessionKey: e.SessionKey,
		Content:    e.Content,
		TokenCount: e.TokenCount,
		Generation: e.Generation,
		CreatedAt:  e.CreatedAt,
	}, nil
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

// DeleteReflectionsBySession deletes all reflections for a session.
func (s *Store) DeleteReflectionsBySession(ctx context.Context, sessionKey string) error {
	_, err := s.client.Reflection.Delete().
		Where(reflection.SessionKey(sessionKey)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete reflections by session: %w", err)
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

// ListRecentReflections returns the N most recent reflections for a session.
// Results are ordered by created_at ascending (oldest first) for chronological display.
func (s *Store) ListRecentReflections(ctx context.Context, sessionKey string, limit int) ([]Reflection, error) {
	entries, err := s.client.Reflection.Query().
		Where(reflection.SessionKey(sessionKey)).
		Order(reflection.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recent reflections: %w", err)
	}

	// Reverse to ascending order for chronological display.
	result := make([]Reflection, len(entries))
	for i, e := range entries {
		result[len(entries)-1-i] = Reflection{
			ID:         e.ID,
			SessionKey: e.SessionKey,
			Content:    e.Content,
			TokenCount: e.TokenCount,
			Generation: e.Generation,
			CreatedAt:  e.CreatedAt,
		}
	}
	return result, nil
}

// ListRecentObservations returns the N most recent observations for a session.
// Results are ordered by created_at ascending (oldest first) for chronological display.
func (s *Store) ListRecentObservations(ctx context.Context, sessionKey string, limit int) ([]Observation, error) {
	entries, err := s.client.Observation.Query().
		Where(observation.SessionKey(sessionKey)).
		Order(observation.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recent observations: %w", err)
	}

	// Reverse to ascending order for chronological display.
	result := make([]Observation, len(entries))
	for i, e := range entries {
		result[len(entries)-1-i] = Observation{
			ID:               e.ID,
			SessionKey:       e.SessionKey,
			Content:          e.Content,
			TokenCount:       e.TokenCount,
			SourceStartIndex: e.SourceStartIndex,
			SourceEndIndex:   e.SourceEndIndex,
			CreatedAt:        e.CreatedAt,
		}
	}
	return result, nil
}
