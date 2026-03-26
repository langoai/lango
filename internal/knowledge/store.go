package knowledge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/auditlog"
	"github.com/langoai/lango/internal/ent/externalref"
	entknowledge "github.com/langoai/lango/internal/ent/knowledge"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	"github.com/langoai/lango/internal/ent/predicate"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/search"
)

// Store provides CRUD operations for knowledge, learning, skill, audit, and external ref entities.
type Store struct {
	client          *ent.Client
	logger          *zap.SugaredLogger
	bus             *eventbus.Bus      // Optional event bus for cross-domain notifications.
	fts5Index       *search.FTS5Index  // Optional FTS5 index for knowledge search.
	learningFTS5Idx *search.FTS5Index  // Optional FTS5 index for learning search.
}

// NewStore creates a new knowledge store.
func NewStore(client *ent.Client, logger *zap.SugaredLogger) *Store {
	return &Store{
		client: client,
		logger: logger,
	}
}

// SetEventBus sets the optional event bus for publishing content events.
func (s *Store) SetEventBus(bus *eventbus.Bus) {
	s.bus = bus
}

// SetFTS5Index sets the optional FTS5 index for knowledge search.
// When set, SearchKnowledge uses FTS5 with BM25 ranking instead of LIKE.
func (s *Store) SetFTS5Index(idx *search.FTS5Index) {
	s.fts5Index = idx
}

// SetLearningFTS5Index sets the optional FTS5 index for learning search.
// When set, SearchLearnings uses FTS5 with BM25 ranking instead of LIKE.
func (s *Store) SetLearningFTS5Index(idx *search.FTS5Index) {
	s.learningFTS5Idx = idx
}

// publishContentSaved publishes a ContentSavedEvent if the bus is configured.
// needsGraph mirrors the original SetGraphCallback behavior: only new knowledge
// creation triggers graph processing; updates and learning saves are embed-only.
func (s *Store) publishContentSaved(id, collection, content string, metadata map[string]string, isNew, needsGraph bool) {
	if s.bus == nil {
		return
	}
	s.bus.Publish(eventbus.ContentSavedEvent{
		ID:         id,
		Collection: collection,
		Content:    content,
		Metadata:   metadata,
		Source:     "knowledge",
		IsNew:      isNew,
		NeedsGraph: needsGraph,
	})
}

// SaveKnowledge creates or updates a knowledge entry by key.
func (s *Store) SaveKnowledge(ctx context.Context, sessionKey string, entry KnowledgeEntry) error {
	existing, err := s.client.Knowledge.Query().
		Where(entknowledge.Key(entry.Key)).
		Only(ctx)

	if ent.IsNotFound(err) {
		builder := s.client.Knowledge.Create().
			SetKey(entry.Key).
			SetCategory(entry.Category).
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

		meta := map[string]string{"category": string(entry.Category)}
		s.publishContentSaved(entry.Key, "knowledge", entry.Content, meta, true, true)
		s.syncKnowledgeFTS5(ctx, entry.Key, entry.Content, false)
		return nil
	}
	if err != nil {
		return fmt.Errorf("query knowledge: %w", err)
	}

	updater := existing.Update().
		SetCategory(entry.Category).
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

	s.publishContentSaved(entry.Key, "knowledge", entry.Content, map[string]string{
		"category": string(entry.Category),
	}, false, false)
	s.syncKnowledgeFTS5(ctx, entry.Key, entry.Content, true)
	return nil
}

// GetKnowledge retrieves a knowledge entry by key.
func (s *Store) GetKnowledge(ctx context.Context, key string) (*KnowledgeEntry, error) {
	k, err := s.client.Knowledge.Query().
		Where(entknowledge.Key(key)).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil, fmt.Errorf("get knowledge %q: %w", key, ErrKnowledgeNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("query knowledge: %w", err)
	}

	return &KnowledgeEntry{
		Key:      k.Key,
		Category: k.Category,
		Content:  k.Content,
		Tags:     k.Tags,
		Source:   k.Source,
	}, nil
}

// SearchKnowledge searches knowledge entries by content/key keyword matching.
// When an FTS5 index is available, it uses FTS5 MATCH with BM25 ranking.
// Falls back to per-keyword OR LIKE predicates when FTS5 is unavailable or errors.
func (s *Store) SearchKnowledge(ctx context.Context, query string, category string, limit int) ([]KnowledgeEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	// FTS5 path: search index first, then resolve from Ent.
	if s.fts5Index != nil && query != "" {
		results, err := s.fts5Index.Search(ctx, query, limit)
		if err != nil {
			s.logger.Warnw("FTS5 knowledge search error, falling back to LIKE", "error", err)
		} else if len(results) > 0 {
			return s.resolveKnowledgeByKeys(ctx, results, category)
		} else {
			return nil, nil
		}
	}

	return s.searchKnowledgeLIKE(ctx, query, category, limit)
}

// searchKnowledgeLIKE is the original LIKE-based search path.
func (s *Store) searchKnowledgeLIKE(ctx context.Context, query string, category string, limit int) ([]KnowledgeEntry, error) {
	var predicates []predicate.Knowledge
	if query != "" {
		if kwPreds := knowledgeKeywordPredicates(query); len(kwPreds) > 0 {
			predicates = append(predicates, entknowledge.Or(kwPreds...))
		}
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
			Category: k.Category,
			Content:  k.Content,
			Tags:     k.Tags,
			Source:   k.Source,
		})
	}
	return result, nil
}

// resolveKnowledgeByKeys loads knowledge entries from Ent by FTS5 result keys,
// preserving FTS5 BM25 ordering. Category filter is applied via Ent.
func (s *Store) resolveKnowledgeByKeys(ctx context.Context, ftsResults []search.SearchResult, category string) ([]KnowledgeEntry, error) {
	keys := make([]string, 0, len(ftsResults))
	for _, r := range ftsResults {
		keys = append(keys, r.RowID)
	}

	preds := []predicate.Knowledge{entknowledge.KeyIn(keys...)}
	if category != "" {
		preds = append(preds, entknowledge.CategoryEQ(entknowledge.Category(category)))
	}

	entries, err := s.client.Knowledge.Query().
		Where(preds...).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve knowledge by keys: %w", err)
	}

	// Build map for ordering by FTS5 rank.
	byKey := make(map[string]*ent.Knowledge, len(entries))
	for _, k := range entries {
		byKey[k.Key] = k
	}

	result := make([]KnowledgeEntry, 0, len(ftsResults))
	for _, r := range ftsResults {
		k, ok := byKey[r.RowID]
		if !ok {
			continue // Filtered out by category or missing.
		}
		result = append(result, KnowledgeEntry{
			Key:      k.Key,
			Category: k.Category,
			Content:  k.Content,
			Tags:     k.Tags,
			Source:   k.Source,
		})
	}
	return result, nil
}

// knowledgeKeywordPredicates splits a query into keywords and creates
// individual ContentContains/KeyContains predicates for each.
func knowledgeKeywordPredicates(query string) []predicate.Knowledge {
	keywords := strings.Fields(query)
	preds := make([]predicate.Knowledge, 0, len(keywords)*2)
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		preds = append(preds,
			entknowledge.ContentContains(kw),
			entknowledge.KeyContains(kw),
		)
	}
	return preds
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
		return fmt.Errorf("increment use count %q: %w", key, ErrKnowledgeNotFound)
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
		return fmt.Errorf("delete knowledge %q: %w", key, ErrKnowledgeNotFound)
	}
	s.deleteKnowledgeFTS5(ctx, key)
	return nil
}

// SaveLearning creates a new learning entry.
func (s *Store) SaveLearning(ctx context.Context, sessionKey string, entry LearningEntry) error {
	builder := s.client.Learning.Create().
		SetTrigger(entry.Trigger).
		SetCategory(entry.Category)

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

	content := entry.Trigger
	if entry.Fix != "" {
		content += "\n" + entry.Fix
	}
	s.publishContentSaved(created.ID.String(), "learning", content, map[string]string{
		"category": string(entry.Category),
	}, true, false)
	s.syncLearningFTS5(ctx, created.ID.String(), entry.Trigger, entry.ErrorPattern, entry.Fix)

	return nil
}

// GetLearning retrieves a learning entry by its UUID.
func (s *Store) GetLearning(ctx context.Context, id uuid.UUID) (*LearningEntry, error) {
	l, err := s.client.Learning.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("get learning %q: %w", id, ErrLearningNotFound)
		}
		return nil, fmt.Errorf("get learning: %w", err)
	}
	return &LearningEntry{
		Trigger:      l.Trigger,
		ErrorPattern: l.ErrorPattern,
		Diagnosis:    l.Diagnosis,
		Fix:          l.Fix,
		Category:     l.Category,
		Tags:         l.Tags,
	}, nil
}

// SearchLearnings searches learnings by error pattern or trigger substring match.
// When an FTS5 index is available, it uses FTS5 MATCH with BM25 ranking.
// Falls back to per-keyword OR LIKE predicates when FTS5 is unavailable or errors.
func (s *Store) SearchLearnings(ctx context.Context, errorPattern string, category string, limit int) ([]LearningEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	// FTS5 path.
	if s.learningFTS5Idx != nil && errorPattern != "" {
		results, err := s.learningFTS5Idx.Search(ctx, errorPattern, limit)
		if err != nil {
			s.logger.Warnw("FTS5 learning search error, falling back to LIKE", "error", err)
		} else if len(results) > 0 {
			return s.resolveLearningsByIDs(ctx, results, category)
		} else {
			return nil, nil
		}
	}

	return s.searchLearningsLIKE(ctx, errorPattern, category, limit)
}

// searchLearningsLIKE is the original LIKE-based search path.
func (s *Store) searchLearningsLIKE(ctx context.Context, errorPattern string, category string, limit int) ([]LearningEntry, error) {
	var predicates []predicate.Learning
	if errorPattern != "" {
		if kwPreds := learningKeywordPredicates(errorPattern); len(kwPreds) > 0 {
			predicates = append(predicates, entlearning.Or(kwPreds...))
		}
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
			Category:     l.Category,
			Tags:         l.Tags,
		})
	}
	return result, nil
}

// resolveLearningsByIDs loads learning entries from Ent by FTS5 result IDs,
// preserving FTS5 BM25 ordering.
func (s *Store) resolveLearningsByIDs(ctx context.Context, ftsResults []search.SearchResult, category string) ([]LearningEntry, error) {
	ids := make([]uuid.UUID, 0, len(ftsResults))
	for _, r := range ftsResults {
		id, err := uuid.Parse(r.RowID)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	preds := []predicate.Learning{entlearning.IDIn(ids...)}
	if category != "" {
		preds = append(preds, entlearning.CategoryEQ(entlearning.Category(category)))
	}

	entries, err := s.client.Learning.Query().
		Where(preds...).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve learnings by IDs: %w", err)
	}

	byID := make(map[uuid.UUID]*ent.Learning, len(entries))
	for _, l := range entries {
		byID[l.ID] = l
	}

	result := make([]LearningEntry, 0, len(ftsResults))
	for _, r := range ftsResults {
		id, err := uuid.Parse(r.RowID)
		if err != nil {
			continue
		}
		l, ok := byID[id]
		if !ok {
			continue
		}
		result = append(result, LearningEntry{
			Trigger:      l.Trigger,
			ErrorPattern: l.ErrorPattern,
			Diagnosis:    l.Diagnosis,
			Fix:          l.Fix,
			Category:     l.Category,
			Tags:         l.Tags,
		})
	}
	return result, nil
}

// learningKeywordPredicates splits a query into keywords and creates
// individual ErrorPatternContains/TriggerContains predicates for each.
func learningKeywordPredicates(query string) []predicate.Learning {
	keywords := strings.Fields(query)
	preds := make([]predicate.Learning, 0, len(keywords)*2)
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		preds = append(preds,
			entlearning.ErrorPatternContains(kw),
			entlearning.TriggerContains(kw),
		)
	}
	return preds
}

// SearchLearningEntities searches learnings and returns raw Ent entities for confidence boosting.
// Uses per-keyword OR predicates to avoid SQLite LIKE pattern complexity limits.
func (s *Store) SearchLearningEntities(ctx context.Context, errorPattern string, limit int) ([]*ent.Learning, error) {
	if limit <= 0 {
		limit = 5
	}

	var predicates []predicate.Learning
	if errorPattern != "" {
		if kwPreds := learningKeywordPredicates(errorPattern); len(kwPreds) > 0 {
			predicates = append(predicates, entlearning.Or(kwPreds...))
		}
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
// Uses per-keyword OR predicates to avoid SQLite LIKE pattern complexity limits.
func (s *Store) SearchExternalRefs(ctx context.Context, query string) ([]ExternalRefEntry, error) {
	var predicates []predicate.ExternalRef
	if query != "" {
		if kwPreds := externalRefKeywordPredicates(query); len(kwPreds) > 0 {
			predicates = append(predicates, externalref.Or(kwPreds...))
		}
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

// externalRefKeywordPredicates splits a query into keywords and creates
// individual NameContains/SummaryContains predicates for each.
func externalRefKeywordPredicates(query string) []predicate.ExternalRef {
	keywords := strings.Fields(query)
	preds := make([]predicate.ExternalRef, 0, len(keywords)*2)
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		preds = append(preds,
			externalref.NameContains(kw),
			externalref.SummaryContains(kw),
		)
	}
	return preds
}

// syncKnowledgeFTS5 updates the FTS5 index after a knowledge write.
// Logs warning on failure; never blocks the Ent write.
func (s *Store) syncKnowledgeFTS5(ctx context.Context, key, content string, isUpdate bool) {
	if s.fts5Index == nil {
		return
	}
	var err error
	if isUpdate {
		err = s.fts5Index.Update(ctx, key, []string{key, content})
	} else {
		err = s.fts5Index.Insert(ctx, key, []string{key, content})
	}
	if err != nil {
		s.logger.Warnw("FTS5 knowledge sync failed", "key", key, "error", err)
	}
}

// deleteKnowledgeFTS5 removes a knowledge entry from the FTS5 index.
func (s *Store) deleteKnowledgeFTS5(ctx context.Context, key string) {
	if s.fts5Index == nil {
		return
	}
	if err := s.fts5Index.Delete(ctx, key); err != nil {
		s.logger.Warnw("FTS5 knowledge delete failed", "key", key, "error", err)
	}
}

// syncLearningFTS5 inserts a learning entry into the FTS5 index.
func (s *Store) syncLearningFTS5(ctx context.Context, id, trigger, errorPattern, fix string) {
	if s.learningFTS5Idx == nil {
		return
	}
	if err := s.learningFTS5Idx.Insert(ctx, id, []string{trigger, errorPattern, fix}); err != nil {
		s.logger.Warnw("FTS5 learning sync failed", "id", id, "error", err)
	}
}

// deleteLearningFTS5 removes a learning entry from the FTS5 index.
func (s *Store) deleteLearningFTS5(ctx context.Context, id string) {
	if s.learningFTS5Idx == nil {
		return
	}
	if err := s.learningFTS5Idx.Delete(ctx, id); err != nil {
		s.logger.Warnw("FTS5 learning delete failed", "id", id, "error", err)
	}
}

// LearningStats holds aggregate statistics about learning entries.
type LearningStats struct {
	TotalCount       int                          `json:"total_count"`
	ByCategory       map[entlearning.Category]int `json:"by_category"`
	AvgConfidence    float64                      `json:"avg_confidence"`
	OldestEntry      time.Time                    `json:"oldest_entry,omitempty"`
	NewestEntry      time.Time                    `json:"newest_entry,omitempty"`
	TotalOccurrences int                          `json:"total_occurrences"`
	TotalSuccesses   int                          `json:"total_successes"`
}

// GetLearningStats returns aggregate statistics about stored learning entries.
func (s *Store) GetLearningStats(ctx context.Context) (*LearningStats, error) {
	entries, err := s.client.Learning.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query learnings: %w", err)
	}

	stats := &LearningStats{
		ByCategory: make(map[entlearning.Category]int),
	}
	stats.TotalCount = len(entries)
	if stats.TotalCount == 0 {
		return stats, nil
	}

	var totalConf float64
	for _, e := range entries {
		stats.ByCategory[e.Category]++
		totalConf += e.Confidence
		stats.TotalOccurrences += e.OccurrenceCount
		stats.TotalSuccesses += e.SuccessCount
		if stats.OldestEntry.IsZero() || e.CreatedAt.Before(stats.OldestEntry) {
			stats.OldestEntry = e.CreatedAt
		}
		if e.CreatedAt.After(stats.NewestEntry) {
			stats.NewestEntry = e.CreatedAt
		}
	}
	stats.AvgConfidence = totalConf / float64(stats.TotalCount)
	return stats, nil
}

// ListLearnings returns learning entries with optional filtering and pagination.
// Pass zero-value for parameters to skip a filter.
func (s *Store) ListLearnings(ctx context.Context, category string, minConfidence float64, olderThan time.Time, limit, offset int) ([]*ent.Learning, int, error) {
	q := s.client.Learning.Query()

	if category != "" {
		q = q.Where(entlearning.CategoryEQ(entlearning.Category(category)))
	}
	if minConfidence > 0 {
		q = q.Where(entlearning.ConfidenceGTE(minConfidence))
	}
	if !olderThan.IsZero() {
		q = q.Where(entlearning.CreatedAtLTE(olderThan))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count learnings: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}
	entries, err := q.
		Order(entlearning.ByCreatedAt(sql.OrderDesc())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list learnings: %w", err)
	}
	return entries, total, nil
}

// DeleteLearning deletes a single learning entry by UUID.
func (s *Store) DeleteLearning(ctx context.Context, id uuid.UUID) error {
	err := s.client.Learning.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("delete learning %q: %w", id, ErrLearningNotFound)
		}
		return fmt.Errorf("delete learning: %w", err)
	}
	s.deleteLearningFTS5(ctx, id.String())
	return nil
}

// DeleteLearningsWhere deletes learning entries matching the given criteria
// and returns the number of deleted entries.
func (s *Store) DeleteLearningsWhere(ctx context.Context, category string, maxConfidence float64, olderThan time.Time) (int, error) {
	q := s.client.Learning.Delete()

	var preds []predicate.Learning
	if category != "" {
		preds = append(preds, entlearning.CategoryEQ(entlearning.Category(category)))
	}
	if maxConfidence > 0 {
		preds = append(preds, entlearning.ConfidenceLTE(maxConfidence))
	}
	if !olderThan.IsZero() {
		preds = append(preds, entlearning.CreatedAtLTE(olderThan))
	}
	if len(preds) == 0 {
		return 0, fmt.Errorf("at least one filter criterion is required for bulk delete")
	}

	n, err := q.Where(preds...).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete learnings: %w", err)
	}
	return n, nil
}
