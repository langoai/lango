package agentmemory

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"entgo.io/ent/dialect/sql"

	"github.com/langoai/lango/internal/ent"
	entam "github.com/langoai/lango/internal/ent/agentmemory"
	"github.com/langoai/lango/internal/ent/predicate"
	"github.com/langoai/lango/internal/security"
)

var _ Store = (*EntStore)(nil)

// EntStore is a persistent Ent-backed implementation of Store.
type EntStore struct {
	client   *ent.Client
	payloads security.PayloadProtector
}

// NewEntStore creates a new Ent-backed agent memory store.
func NewEntStore(client *ent.Client) *EntStore {
	return &EntStore{client: client}
}

func (s *EntStore) SetPayloadProtector(protector security.PayloadProtector) {
	s.payloads = protector
}

func (s *EntStore) Save(entry *Entry) error {
	if entry.AgentName == "" {
		return fmt.Errorf("save: agent_name is required")
	}
	if entry.Key == "" {
		return fmt.Errorf("save: key is required")
	}
	if entry.Kind != "" && !entry.Kind.Valid() {
		return fmt.Errorf("save: invalid memory kind %q", entry.Kind)
	}

	ctx := context.Background()

	existing, err := s.client.AgentMemory.Query().
		Where(
			entam.AgentName(entry.AgentName),
			entam.Key(entry.Key),
		).
		Only(ctx)

	if ent.IsNotFound(err) {
		content, ciphertext, nonce, keyVersion, err := s.prepareContent(entry.Content)
		if err != nil {
			return fmt.Errorf("prepare agent memory payload: %w", err)
		}
		builder := s.client.AgentMemory.Create().
			SetAgentName(entry.AgentName).
			SetKey(entry.Key).
			SetContent(content).
			SetScope(entam.Scope(entry.Scope)).
			SetKind(entam.Kind(entry.Kind)).
			SetConfidence(entry.Confidence)
		if ciphertext != nil {
			builder.SetContentCiphertext(ciphertext)
			builder.SetContentNonce(nonce)
			builder.SetContentKeyVersion(keyVersion)
		}

		if len(entry.Tags) > 0 {
			builder.SetTags(entry.Tags)
		}

		_, err = builder.Save(ctx)
		if err != nil {
			return fmt.Errorf("create agent memory: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("query agent memory: %w", err)
	}

	// Upsert: update mutable fields, preserve ID and CreatedAt.
	content, ciphertext, nonce, keyVersion, err := s.prepareContent(entry.Content)
	if err != nil {
		return fmt.Errorf("prepare agent memory payload: %w", err)
	}
	updater := existing.Update().
		SetContent(content).
		SetScope(entam.Scope(entry.Scope)).
		SetKind(entam.Kind(entry.Kind)).
		SetConfidence(entry.Confidence)
	if ciphertext != nil {
		updater.SetContentCiphertext(ciphertext)
		updater.SetContentNonce(nonce)
		updater.SetContentKeyVersion(keyVersion)
	}

	if len(entry.Tags) > 0 {
		updater.SetTags(entry.Tags)
	} else {
		updater.ClearTags()
	}

	_, err = updater.Save(ctx)
	if err != nil {
		return fmt.Errorf("update agent memory: %w", err)
	}
	return nil
}

func (s *EntStore) Get(agentName, key string) (*Entry, error) {
	ctx := context.Background()

	e, err := s.client.AgentMemory.Query().
		Where(
			entam.AgentName(agentName),
			entam.Key(key),
		).
		Only(ctx)

	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get agent memory: %w", err)
	}

	return s.entToEntry(e)
}

func (s *EntStore) Search(agentName string, opts SearchOptions) ([]*Entry, error) {
	ctx := context.Background()

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	var preds []predicate.AgentMemory
	preds = append(preds, entam.AgentName(agentName))

	if opts.Scope != "" {
		preds = append(preds, entam.ScopeEQ(entam.Scope(opts.Scope)))
	}
	if opts.Kind != "" {
		preds = append(preds, entam.KindEQ(entam.Kind(opts.Kind)))
	}
	if opts.MinConfidence > 0 {
		preds = append(preds, entam.ConfidenceGTE(opts.MinConfidence))
	}
	if opts.Query != "" {
		kwPreds := keywordPredicates(opts.Query)
		if len(kwPreds) > 0 {
			preds = append(preds, entam.Or(kwPreds...))
		}
	}
	if len(opts.Tags) > 0 {
		preds = append(preds, tagPredicates(opts.Tags)...)
	}

	entries, err := s.client.AgentMemory.Query().
		Where(preds...).
		Order(
			entam.ByConfidence(sql.OrderDesc()),
			entam.ByUseCount(sql.OrderDesc()),
			entam.ByUpdatedAt(sql.OrderDesc()),
		).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("search agent memory: %w", err)
	}

	return s.entsToEntries(entries)
}

// SearchWithContext resolves entries with scope fallback:
// Phase 1 = agent's all entries (any scope), Phase 2 = other agents' global only.
// Phase 1 sorted first, Phase 2 after, then global limit.
//
// NOTE: type scope not yet implemented — only instance and global scopes are
// considered. Type scope entries are skipped during context resolution.
func (s *EntStore) SearchWithContext(agentName string, query string, limit int) ([]*Entry, error) {
	return s.SearchWithContextOptions(agentName, SearchOptions{Query: query, Limit: limit})
}

// SearchWithContextOptions is like SearchWithContext but accepts full
// SearchOptions (Kind, Tags, MinConfidence). Filters are applied during
// collection, before limit truncation, so no results are lost.
//
// NOTE: type scope not yet implemented — only instance and global scopes are
// considered. Type scope entries are skipped during context resolution.
func (s *EntStore) SearchWithContextOptions(agentName string, opts SearchOptions) ([]*Entry, error) {
	ctx := context.Background()

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	// Phase 1: all entries for this agent (any scope).
	phase1Preds := []predicate.AgentMemory{entam.AgentName(agentName)}
	if opts.Query != "" {
		if kwPreds := keywordPredicates(opts.Query); len(kwPreds) > 0 {
			phase1Preds = append(phase1Preds, entam.Or(kwPreds...))
		}
	}
	phase1Preds = append(phase1Preds, contextFilterPredicates(opts)...)

	phase1, err := s.client.AgentMemory.Query().
		Where(phase1Preds...).
		Order(
			entam.ByConfidence(sql.OrderDesc()),
			entam.ByUseCount(sql.OrderDesc()),
			entam.ByUpdatedAt(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("search agent memory phase1: %w", err)
	}

	// Phase 2: global entries from other agents.
	phase2Preds := []predicate.AgentMemory{
		entam.AgentNameNEQ(agentName),
		entam.ScopeEQ(entam.ScopeGlobal),
	}
	if opts.Query != "" {
		if kwPreds := keywordPredicates(opts.Query); len(kwPreds) > 0 {
			phase2Preds = append(phase2Preds, entam.Or(kwPreds...))
		}
	}
	phase2Preds = append(phase2Preds, contextFilterPredicates(opts)...)

	phase2, err := s.client.AgentMemory.Query().
		Where(phase2Preds...).
		Order(
			entam.ByConfidence(sql.OrderDesc()),
			entam.ByUseCount(sql.OrderDesc()),
			entam.ByUpdatedAt(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("search agent memory phase2: %w", err)
	}

	// Combine: phase 1 first, then phase 2, then truncate to limit.
	combined := make([]*Entry, 0, len(phase1)+len(phase2))
	for _, e := range phase1 {
		entry, err := s.entToEntry(e)
		if err != nil {
			return nil, err
		}
		combined = append(combined, entry)
	}
	for _, e := range phase2 {
		entry, err := s.entToEntry(e)
		if err != nil {
			return nil, err
		}
		combined = append(combined, entry)
	}

	if len(combined) > limit {
		combined = combined[:limit]
	}
	return combined, nil
}

func (s *EntStore) Delete(agentName, key string) error {
	ctx := context.Background()

	_, err := s.client.AgentMemory.Delete().
		Where(
			entam.AgentName(agentName),
			entam.Key(key),
		).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("delete agent memory: %w", err)
	}
	return nil
}

func (s *EntStore) IncrementUseCount(agentName, key string) error {
	ctx := context.Background()

	n, err := s.client.AgentMemory.Update().
		Where(
			entam.AgentName(agentName),
			entam.Key(key),
		).
		AddUseCount(1).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("increment use count: %w", err)
	}
	// No error if entry not found (matches InMemoryStore behavior).
	_ = n
	return nil
}

func (s *EntStore) Prune(agentName string, minConfidence float64) (int, error) {
	ctx := context.Background()

	n, err := s.client.AgentMemory.Delete().
		Where(
			entam.AgentName(agentName),
			entam.ConfidenceLT(minConfidence),
		).
		Exec(ctx)

	if err != nil {
		return 0, fmt.Errorf("prune agent memory: %w", err)
	}
	return n, nil
}

func (s *EntStore) ListAgentNames() ([]string, error) {
	ctx := context.Background()

	// Use GroupBy to get distinct agent names.
	var results []struct {
		AgentName string `json:"agent_name"`
	}
	err := s.client.AgentMemory.Query().
		GroupBy(entam.FieldAgentName).
		Scan(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("list agent names: %w", err)
	}

	names := make([]string, 0, len(results))
	for _, r := range results {
		names = append(names, r.AgentName)
	}
	sort.Strings(names)
	return names, nil
}

// ListAll returns all entries for a given agent, sorted by updated_at DESC.
func (s *EntStore) ListAll(agentName string) ([]*Entry, error) {
	ctx := context.Background()

	entries, err := s.client.AgentMemory.Query().
		Where(entam.AgentName(agentName)).
		Order(entam.ByUpdatedAt(sql.OrderDesc())).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("list all agent memory: %w", err)
	}

	return s.entsToEntries(entries)
}

// ─── helpers ──────────────────────────────────────────────────────

// entToEntry converts an Ent AgentMemory entity to a domain Entry.
func (s *EntStore) entToEntry(e *ent.AgentMemory) (*Entry, error) {
	content := e.Content
	if s.payloads != nil && e.ContentCiphertext != nil && e.ContentNonce != nil && e.ContentKeyVersion != nil {
		plaintext, err := s.payloads.DecryptPayload(*e.ContentCiphertext, *e.ContentNonce, *e.ContentKeyVersion)
		if err != nil {
			return nil, fmt.Errorf("decrypt agent memory %q: %w", e.ID, err)
		}
		content = string(plaintext)
	}
	return &Entry{
		ID:         e.ID.String(),
		AgentName:  e.AgentName,
		Scope:      MemoryScope(e.Scope),
		Kind:       MemoryKind(e.Kind),
		Key:        e.Key,
		Content:    content,
		Confidence: e.Confidence,
		UseCount:   e.UseCount,
		Tags:       e.Tags,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}, nil
}

// entsToEntries converts a slice of Ent entities to domain entries.
func (s *EntStore) entsToEntries(ents []*ent.AgentMemory) ([]*Entry, error) {
	result := make([]*Entry, 0, len(ents))
	for _, e := range ents {
		entry, err := s.entToEntry(e)
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, nil
}

func (s *EntStore) prepareContent(content string) (string, []byte, []byte, int, error) {
	if s.payloads == nil {
		return strings.TrimSpace(content), nil, nil, 0, nil
	}
	projection, ciphertext, nonce, keyVersion, err := security.ProtectText(s.payloads, content, 256)
	if err != nil {
		return "", nil, nil, 0, err
	}
	return projection, ciphertext, nonce, keyVersion, nil
}

// keywordPredicates splits a query into keywords and creates
// individual ContentContains/KeyContains predicates for each.
func keywordPredicates(query string) []predicate.AgentMemory {
	keywords := strings.Fields(query)
	preds := make([]predicate.AgentMemory, 0, len(keywords)*2)
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		preds = append(preds,
			entam.ContentContains(kw),
			entam.KeyContains(kw),
		)
	}
	return preds
}

// contextFilterPredicates builds predicates for Kind, MinConfidence, and Tags
// used in SearchWithContext* methods.
func contextFilterPredicates(opts SearchOptions) []predicate.AgentMemory {
	var preds []predicate.AgentMemory
	if opts.Kind != "" {
		preds = append(preds, entam.KindEQ(entam.Kind(opts.Kind)))
	}
	if opts.MinConfidence > 0 {
		preds = append(preds, entam.ConfidenceGTE(opts.MinConfidence))
	}
	if len(opts.Tags) > 0 {
		preds = append(preds, tagPredicates(opts.Tags)...)
	}
	return preds
}

// tagPredicates creates OR predicates for JSON array tag matching.
// Since Ent JSON fields don't have native array-contains predicates for
// all backends, we use ContentContains on the tags field as an approximation.
// This works correctly for SQLite's JSON storage.
func tagPredicates(tags []string) []predicate.AgentMemory {
	preds := make([]predicate.AgentMemory, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		// Match against the raw JSON string stored in the tags column.
		preds = append(preds, predicate.AgentMemory(sql.FieldContains(entam.FieldTags, tag)))
	}
	if len(preds) == 0 {
		return nil
	}
	return []predicate.AgentMemory{entam.Or(preds...)}
}
