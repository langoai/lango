package session

import (
	"context"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"

	sa "github.com/langoai/lango/internal/smartaccount"
)

// Store persists session keys.
type Store interface {
	Save(ctx context.Context, key *sa.SessionKey) error
	Get(ctx context.Context, id string) (*sa.SessionKey, error)
	List(ctx context.Context) ([]*sa.SessionKey, error)
	Delete(ctx context.Context, id string) error
	ListByParent(ctx context.Context, parentID string) ([]*sa.SessionKey, error)
	ListActive(ctx context.Context) ([]*sa.SessionKey, error)
}

// MemoryStore is an in-memory Store implementation.
type MemoryStore struct {
	mu   sync.RWMutex
	keys map[string]*sa.SessionKey
}

// NewMemoryStore creates a new in-memory session key store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]*sa.SessionKey)}
}

// Save stores a copy of the session key.
func (s *MemoryStore) Save(_ context.Context, key *sa.SessionKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := copySessionKey(key)
	s.keys[cp.ID] = cp
	return nil
}

// Get returns a copy of the session key with the given ID.
func (s *MemoryStore) Get(_ context.Context, id string) (*sa.SessionKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.keys[id]
	if !ok {
		return nil, sa.ErrSessionNotFound
	}
	return copySessionKey(key), nil
}

// List returns all session keys sorted by CreatedAt ascending.
func (s *MemoryStore) List(_ context.Context) ([]*sa.SessionKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*sa.SessionKey, 0, len(s.keys))
	for _, key := range s.keys {
		result = append(result, copySessionKey(key))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

// Delete removes the session key with the given ID.
func (s *MemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.keys[id]; !ok {
		return sa.ErrSessionNotFound
	}
	delete(s.keys, id)
	return nil
}

// ListByParent returns all session keys with the given parent ID.
func (s *MemoryStore) ListByParent(
	_ context.Context, parentID string,
) ([]*sa.SessionKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*sa.SessionKey
	for _, key := range s.keys {
		if key.ParentID == parentID {
			result = append(result, copySessionKey(key))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

// ListActive returns all session keys that are currently active.
func (s *MemoryStore) ListActive(_ context.Context) ([]*sa.SessionKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*sa.SessionKey
	for _, key := range s.keys {
		if key.IsActive() {
			result = append(result, copySessionKey(key))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

// copySessionKey returns a deep copy of a session key.
func copySessionKey(src *sa.SessionKey) *sa.SessionKey {
	cp := *src

	if src.PublicKey != nil {
		cp.PublicKey = make([]byte, len(src.PublicKey))
		copy(cp.PublicKey, src.PublicKey)
	}

	cp.Policy = copyPolicy(src.Policy)
	return &cp
}

// copyPolicy returns a deep copy of a session policy.
func copyPolicy(src sa.SessionPolicy) sa.SessionPolicy {
	cp := src

	if src.AllowedTargets != nil {
		cp.AllowedTargets = make([]common.Address, len(src.AllowedTargets))
		copy(cp.AllowedTargets, src.AllowedTargets)
	}

	if src.AllowedFunctions != nil {
		cp.AllowedFunctions = make([]string, len(src.AllowedFunctions))
		copy(cp.AllowedFunctions, src.AllowedFunctions)
	}

	if src.SpendLimit != nil {
		cp.SpendLimit = new(big.Int).Set(src.SpendLimit)
	}
	return cp
}
