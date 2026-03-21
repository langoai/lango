package provenance

import (
	"context"
	"sort"
	"sync"
)

// MemoryAttributionStore is an in-memory AttributionStore for tests.
type MemoryAttributionStore struct {
	mu    sync.RWMutex
	items map[string]Attribution
}

// NewMemoryAttributionStore creates a new in-memory attribution store.
func NewMemoryAttributionStore() *MemoryAttributionStore {
	return &MemoryAttributionStore{
		items: make(map[string]Attribution),
	}
}

func (s *MemoryAttributionStore) SaveAttribution(_ context.Context, attr Attribution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[attr.ID] = attr
	return nil
}

func (s *MemoryAttributionStore) ListBySession(_ context.Context, sessionKey string, limit int) ([]Attribution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Attribution, 0, len(s.items))
	for _, item := range s.items {
		if item.SessionKey == sessionKey {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
