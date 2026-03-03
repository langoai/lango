package agentmemory

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var _ Store = (*InMemoryStore)(nil)

// InMemoryStore is a thread-safe in-memory implementation of Store.
type InMemoryStore struct {
	mu      sync.RWMutex
	entries map[string]map[string]*Entry // agentName -> key -> Entry
}

// NewInMemoryStore creates a new in-memory agent memory store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		entries: make(map[string]map[string]*Entry),
	}
}

func (s *InMemoryStore) Save(entry *Entry) error {
	if entry.AgentName == "" {
		return fmt.Errorf("save: agent_name is required")
	}
	if entry.Key == "" {
		return fmt.Errorf("save: key is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	agentMap, ok := s.entries[entry.AgentName]
	if !ok {
		agentMap = make(map[string]*Entry)
		s.entries[entry.AgentName] = agentMap
	}

	now := time.Now()
	if existing, ok := agentMap[entry.Key]; ok {
		// Upsert: update mutable fields, preserve ID and CreatedAt.
		existing.Content = entry.Content
		existing.Scope = entry.Scope
		existing.Kind = entry.Kind
		existing.Confidence = entry.Confidence
		existing.Tags = entry.Tags
		existing.UpdatedAt = now
	} else {
		clone := *entry
		if clone.ID == "" {
			clone.ID = uuid.New().String()
		}
		clone.CreatedAt = now
		clone.UpdatedAt = now
		agentMap[entry.Key] = &clone
	}

	return nil
}

func (s *InMemoryStore) Get(agentName, key string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agentMap, ok := s.entries[agentName]
	if !ok {
		return nil, nil
	}
	e, ok := agentMap[key]
	if !ok {
		return nil, nil
	}
	clone := *e
	return &clone, nil
}

func (s *InMemoryStore) Search(agentName string, opts SearchOptions) ([]*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agentMap := s.entries[agentName]
	if agentMap == nil {
		return nil, nil
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	var results []*Entry
	for _, e := range agentMap {
		if !matchesSearch(e, opts) {
			continue
		}
		clone := *e
		results = append(results, &clone)
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

func (s *InMemoryStore) SearchWithContext(agentName string, query string, limit int) ([]*Entry, error) {
	if limit <= 0 {
		limit = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Entry
	queryLower := strings.ToLower(query)

	// Phase 1: instance-scoped entries for this agent.
	if agentMap := s.entries[agentName]; agentMap != nil {
		for _, e := range agentMap {
			if matchesQuery(e, queryLower) {
				clone := *e
				results = append(results, &clone)
			}
		}
	}

	// Phase 2: global entries from all agents.
	for name, agentMap := range s.entries {
		if name == agentName {
			continue
		}
		for _, e := range agentMap {
			if e.Scope != ScopeGlobal {
				continue
			}
			if matchesQuery(e, queryLower) {
				clone := *e
				results = append(results, &clone)
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func (s *InMemoryStore) Delete(agentName, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agentMap, ok := s.entries[agentName]
	if !ok {
		return nil
	}
	delete(agentMap, key)
	return nil
}

func (s *InMemoryStore) IncrementUseCount(agentName, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agentMap, ok := s.entries[agentName]
	if !ok {
		return nil
	}
	if e, ok := agentMap[key]; ok {
		e.UseCount++
		e.UpdatedAt = time.Now()
	}
	return nil
}

func (s *InMemoryStore) Prune(agentName string, minConfidence float64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agentMap, ok := s.entries[agentName]
	if !ok {
		return 0, nil
	}

	var pruned int
	for key, e := range agentMap {
		if e.Confidence < minConfidence {
			delete(agentMap, key)
			pruned++
		}
	}
	return pruned, nil
}

// matchesSearch returns true if the entry matches the given search options.
func matchesSearch(e *Entry, opts SearchOptions) bool {
	if opts.Scope != "" && e.Scope != opts.Scope {
		return false
	}
	if opts.Kind != "" && e.Kind != opts.Kind {
		return false
	}
	if opts.MinConfidence > 0 && e.Confidence < opts.MinConfidence {
		return false
	}
	if len(opts.Tags) > 0 && !hasAnyTag(e.Tags, opts.Tags) {
		return false
	}
	if opts.Query != "" {
		return matchesQuery(e, strings.ToLower(opts.Query))
	}
	return true
}

// matchesQuery returns true if the entry's key or content contains the query.
func matchesQuery(e *Entry, queryLower string) bool {
	return strings.Contains(strings.ToLower(e.Key), queryLower) ||
		strings.Contains(strings.ToLower(e.Content), queryLower)
}

// hasAnyTag returns true if entryTags contains any of the filter tags.
func hasAnyTag(entryTags, filterTags []string) bool {
	set := make(map[string]struct{}, len(entryTags))
	for _, t := range entryTags {
		set[t] = struct{}{}
	}
	for _, t := range filterTags {
		if _, ok := set[t]; ok {
			return true
		}
	}
	return false
}
