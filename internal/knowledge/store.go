package knowledge

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/auditlog"
	"github.com/langowarny/lango/internal/ent/externalref"
	entknowledge "github.com/langowarny/lango/internal/ent/knowledge"
	entlearning "github.com/langowarny/lango/internal/ent/learning"
	"github.com/langowarny/lango/internal/ent/predicate"
)

// EmbedCallback is an optional hook called when content is saved, enabling
// asynchronous embedding without importing the embedding package.
type EmbedCallback func(id, collection, content string, metadata map[string]string)

// GraphCallback is an optional hook called when content is saved, enabling
// asynchronous graph relationship updates without importing the graph package.
type GraphCallback func(id, collection, content string, metadata map[string]string)

// Store provides CRUD operations for knowledge, learning, skill, audit, and external ref entities.
type Store struct {
	client *ent.Client
	logger *zap.SugaredLogger

	// Optional embedding hook (nil = disabled).
	onEmbed EmbedCallback
	// Optional graph relationship hook (nil = disabled).
	onGraph GraphCallback

	// Rate limiting per session
	mu              sync.Mutex
	knowledgeCounts map[string]int
	learningCounts  map[string]int
	maxKnowledge    int
	maxLearnings    int
}

// NewStore creates a new knowledge store.
func NewStore(client *ent.Client, logger *zap.SugaredLogger, maxKnowledge, maxLearnings int) *Store {
	if maxKnowledge <= 0 {
		maxKnowledge = 20
	}
	if maxLearnings <= 0 {
		maxLearnings = 10
	}
	return &Store{
		client:          client,
		logger:          logger,
		knowledgeCounts: make(map[string]int),
		learningCounts:  make(map[string]int),
		maxKnowledge:    maxKnowledge,
		maxLearnings:    maxLearnings,
	}
}

// SetEmbedCallback sets the optional embedding hook.
func (s *Store) SetEmbedCallback(cb EmbedCallback) {
	s.onEmbed = cb
}

// SetGraphCallback sets the optional graph relationship hook.
func (s *Store) SetGraphCallback(cb GraphCallback) {
	s.onGraph = cb
}

// SaveKnowledge creates or updates a knowledge entry by key.
func (s *Store) SaveKnowledge(ctx context.Context, sessionKey string, entry KnowledgeEntry) error {
	if err := s.reserveKnowledgeSlot(sessionKey); err != nil {
		return err
	}

	existing, err := s.client.Knowledge.Query().
		Where(entknowledge.Key(entry.Key)).
		Only(ctx)

	if ent.IsNotFound(err) {
		builder := s.client.Knowledge.Create().
			SetKey(entry.Key).
			SetCategory(entknowledge.Category(entry.Category)).
			SetContent(entry.Content)

		if len(entry.Tags) > 0 {
			builder.SetTags(entry.Tags)
		}
		if entry.Source != "" {
			builder.SetSource(entry.Source)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("create knowledge: %w", err)
		}

		meta := map[string]string{"category": entry.Category}
		if s.onEmbed != nil {
			s.onEmbed(entry.Key, "knowledge", entry.Content, meta)
		}
		if s.onGraph != nil {
			s.onGraph(entry.Key, "knowledge", entry.Content, meta)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("query knowledge: %w", err)
	}

	updater := existing.Update().
		SetCategory(entknowledge.Category(entry.Category)).
		SetContent(entry.Content)

	if len(entry.Tags) > 0 {
		updater.SetTags(entry.Tags)
	}
	if entry.Source != "" {
		updater.SetSource(entry.Source)
	}

	_, err = updater.Save(ctx)
	if err != nil {
		return fmt.Errorf("update knowledge: %w", err)
	}

	if s.onEmbed != nil {
		s.onEmbed(entry.Key, "knowledge", entry.Content, map[string]string{
			"category": entry.Category,
		})
	}
	return nil
}

// GetKnowledge retrieves a knowledge entry by key.
func (s *Store) GetKnowledge(ctx context.Context, key string) (*KnowledgeEntry, error) {
	k, err := s.client.Knowledge.Query().
		Where(entknowledge.Key(key)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil, fmt.Errorf("knowledge not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("query knowledge: %w", err)
	}

	return &KnowledgeEntry{
		Key:      k.Key,
		Category: string(k.Category),
		Content:  k.Content,
		Tags:     k.Tags,
		Source:   k.Source,
	}, nil
}

// SearchKnowledge searches knowledge entries by content/key keyword matching.
func (s *Store) SearchKnowledge(ctx context.Context, query string, category string, limit int) ([]KnowledgeEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	var predicates []predicate.Knowledge
	if query != "" {
		predicates = append(predicates, entknowledge.Or(
			entknowledge.ContentContains(query),
			entknowledge.KeyContains(query),
		))
	}
	if category != "" {
		predicates = append(predicates, entknowledge.CategoryEQ(entknowledge.Category(category)))
	}

	entries, err := s.client.Knowledge.Query().
		Where(predicates...).
		Order(entknowledge.ByRelevanceScore(sql.OrderDesc())).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("search knowledge: %w", err)
	}

	result := make([]KnowledgeEntry, 0, len(entries))
	for _, k := range entries {
		result = append(result, KnowledgeEntry{
			Key:      k.Key,
			Category: string(k.Category),
			Content:  k.Content,
			Tags:     k.Tags,
			Source:   k.Source,
		})
	}
	return result, nil
}

// IncrementKnowledgeUseCount increments the use count for a knowledge entry.
func (s *Store) IncrementKnowledgeUseCount(ctx context.Context, key string) error {
	n, err := s.client.Knowledge.Update().
		Where(entknowledge.Key(key)).
		AddUseCount(1).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("increment knowledge use count: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("knowledge not found: %s", key)
	}
	return nil
}

// DeleteKnowledge deletes a knowledge entry by key.
func (s *Store) DeleteKnowledge(ctx context.Context, key string) error {
	n, err := s.client.Knowledge.Delete().
		Where(entknowledge.Key(key)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete knowledge: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("knowledge not found: %s", key)
	}
	return nil
}

// SaveLearning creates a new learning entry.
func (s *Store) SaveLearning(ctx context.Context, sessionKey string, entry LearningEntry) error {
	if err := s.reserveLearningSlot(sessionKey); err != nil {
		return err
	}

	builder := s.client.Learning.Create().
		SetTrigger(entry.Trigger).
		SetCategory(entlearning.Category(entry.Category))

	if entry.ErrorPattern != "" {
		builder.SetErrorPattern(entry.ErrorPattern)
	}
	if entry.Diagnosis != "" {
		builder.SetDiagnosis(entry.Diagnosis)
	}
	if entry.Fix != "" {
		builder.SetFix(entry.Fix)
	}
	if len(entry.Tags) > 0 {
		builder.SetTags(entry.Tags)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create learning: %w", err)
	}

	if s.onEmbed != nil {
		content := entry.Trigger
		if entry.Fix != "" {
			content += "\n" + entry.Fix
		}
		s.onEmbed(created.ID.String(), "learning", content, map[string]string{
			"category": entry.Category,
		})
	}

	return nil
}

// GetLearning retrieves a learning entry by its UUID.
func (s *Store) GetLearning(ctx context.Context, id uuid.UUID) (*LearningEntry, error) {
	l, err := s.client.Learning.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get learning: %w", err)
	}
	return &LearningEntry{
		Trigger:      l.Trigger,
		ErrorPattern: l.ErrorPattern,
		Diagnosis:    l.Diagnosis,
		Fix:          l.Fix,
		Category:     string(l.Category),
		Tags:         l.Tags,
	}, nil
}

// SearchLearnings searches learnings by error pattern or trigger substring match.
func (s *Store) SearchLearnings(ctx context.Context, errorPattern string, category string, limit int) ([]LearningEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	var predicates []predicate.Learning
	if errorPattern != "" {
		predicates = append(predicates, entlearning.Or(
			entlearning.ErrorPatternContains(errorPattern),
			entlearning.TriggerContains(errorPattern),
		))
	}
	if category != "" {
		predicates = append(predicates, entlearning.CategoryEQ(entlearning.Category(category)))
	}

	entries, err := s.client.Learning.Query().
		Where(predicates...).
		Order(entlearning.ByConfidence(sql.OrderDesc())).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("search learnings: %w", err)
	}

	result := make([]LearningEntry, 0, len(entries))
	for _, l := range entries {
		result = append(result, LearningEntry{
			Trigger:      l.Trigger,
			ErrorPattern: l.ErrorPattern,
			Diagnosis:    l.Diagnosis,
			Fix:          l.Fix,
			Category:     string(l.Category),
			Tags:         l.Tags,
		})
	}
	return result, nil
}

// SearchLearningEntities searches learnings and returns raw Ent entities for confidence boosting.
func (s *Store) SearchLearningEntities(ctx context.Context, errorPattern string, limit int) ([]*ent.Learning, error) {
	if limit <= 0 {
		limit = 5
	}

	var predicates []predicate.Learning
	if errorPattern != "" {
		predicates = append(predicates, entlearning.Or(
			entlearning.ErrorPatternContains(errorPattern),
			entlearning.TriggerContains(errorPattern),
		))
	}

	return s.client.Learning.Query().
		Where(predicates...).
		Order(entlearning.ByConfidence(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

// BoostLearningConfidence increments success count and recalculates confidence.
// When confidenceBoost > 0, it is added directly to the current confidence (for
// fractional graph propagation). When 0, the existing success/occurrence ratio is used.
// Confidence is always clamped to [0.1, 1.0].
func (s *Store) BoostLearningConfidence(ctx context.Context, id uuid.UUID, successDelta int, confidenceBoost float64) error {
	l, err := s.client.Learning.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get learning: %w", err)
	}

	newSuccess := l.SuccessCount + successDelta
	newOccurrence := l.OccurrenceCount + 1

	var newConfidence float64
	if confidenceBoost > 0 {
		newConfidence = l.Confidence + confidenceBoost
	} else {
		newConfidence = float64(newSuccess) / float64(newSuccess+newOccurrence)
	}

	if newConfidence < 0.1 {
		newConfidence = 0.1
	}
	if newConfidence > 1.0 {
		newConfidence = 1.0
	}

	_, err = l.Update().
		SetSuccessCount(newSuccess).
		SetOccurrenceCount(newOccurrence).
		SetConfidence(newConfidence).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("boost learning confidence: %w", err)
	}
	return nil
}

// SaveAuditLog creates a new audit log entry.
func (s *Store) SaveAuditLog(ctx context.Context, entry AuditEntry) error {
	builder := s.client.AuditLog.Create().
		SetAction(auditlog.Action(entry.Action)).
		SetActor(entry.Actor)

	if entry.SessionKey != "" {
		builder.SetSessionKey(entry.SessionKey)
	}
	if entry.Target != "" {
		builder.SetTarget(entry.Target)
	}
	if entry.Details != nil {
		builder.SetDetails(entry.Details)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

// SaveExternalRef creates or updates an external reference.
func (s *Store) SaveExternalRef(ctx context.Context, name, refType, location, summary string) error {
	existing, err := s.client.ExternalRef.Query().
		Where(externalref.Name(name)).
		Only(ctx)

	if ent.IsNotFound(err) {
		builder := s.client.ExternalRef.Create().
			SetName(name).
			SetRefType(externalref.RefType(refType)).
			SetLocation(location)

		if summary != "" {
			builder.SetSummary(summary)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("create external ref: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("query external ref: %w", err)
	}

	updater := existing.Update().
		SetRefType(externalref.RefType(refType)).
		SetLocation(location)

	if summary != "" {
		updater.SetSummary(summary)
	}

	_, err = updater.Save(ctx)
	if err != nil {
		return fmt.Errorf("update external ref: %w", err)
	}
	return nil
}

// SearchExternalRefs searches external references by name or summary.
func (s *Store) SearchExternalRefs(ctx context.Context, query string) ([]ExternalRefEntry, error) {
	var predicates []predicate.ExternalRef
	if query != "" {
		predicates = append(predicates, externalref.Or(
			externalref.NameContains(query),
			externalref.SummaryContains(query),
		))
	}

	refs, err := s.client.ExternalRef.Query().
		Where(predicates...).
		Limit(10).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("search external refs: %w", err)
	}

	result := make([]ExternalRefEntry, 0, len(refs))
	for _, r := range refs {
		result = append(result, ExternalRefEntry{
			Name:     r.Name,
			RefType:  string(r.RefType),
			Location: r.Location,
			Summary:  r.Summary,
			Metadata: r.Metadata,
		})
	}
	return result, nil
}

// Rate limiting helpers

func (s *Store) reserveKnowledgeSlot(sessionKey string) error {
	if sessionKey == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.knowledgeCounts[sessionKey] >= s.maxKnowledge {
		return fmt.Errorf("knowledge save limit reached for session (%d/%d)", s.maxKnowledge, s.maxKnowledge)
	}
	s.knowledgeCounts[sessionKey]++
	return nil
}

func (s *Store) reserveLearningSlot(sessionKey string) error {
	if sessionKey == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.learningCounts[sessionKey] >= s.maxLearnings {
		return fmt.Errorf("learning save limit reached for session (%d/%d)", s.maxLearnings, s.maxLearnings)
	}
	s.learningCounts[sessionKey]++
	return nil
}

