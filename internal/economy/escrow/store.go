package escrow

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrEscrowExists   = errors.New("escrow already exists")
	ErrEscrowNotFound = errors.New("escrow not found")
)

// Store defines the interface for escrow persistence.
type Store interface {
	Create(entry *EscrowEntry) error
	Get(id string) (*EscrowEntry, error)
	List() []*EscrowEntry
	ListByPeer(peerDID string) []*EscrowEntry
	ListByStatus(status EscrowStatus) []*EscrowEntry
	ListByStatusBefore(status EscrowStatus, before time.Time) []*EscrowEntry
	Update(entry *EscrowEntry) error
	Delete(id string) error
}

// memoryStore implements Store in-memory.
type memoryStore struct {
	mu      sync.RWMutex
	escrows map[string]*EscrowEntry
}

// NewMemoryStore creates a new in-memory escrow store.
func NewMemoryStore() Store {
	return &memoryStore{
		escrows: make(map[string]*EscrowEntry),
	}
}

func (s *memoryStore) Create(entry *EscrowEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.escrows[entry.ID]; exists {
		return fmt.Errorf("create %q: %w", entry.ID, ErrEscrowExists)
	}

	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now
	s.escrows[entry.ID] = entry
	return nil
}

func (s *memoryStore) Get(id string) (*EscrowEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.escrows[id]
	if !exists {
		return nil, fmt.Errorf("get %q: %w", id, ErrEscrowNotFound)
	}
	return entry, nil
}

func (s *memoryStore) List() []*EscrowEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*EscrowEntry, 0, len(s.escrows))
	for _, e := range s.escrows {
		result = append(result, e)
	}
	return result
}

func (s *memoryStore) ListByStatus(status EscrowStatus) []*EscrowEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*EscrowEntry
	for _, e := range s.escrows {
		if e.Status == status {
			result = append(result, e)
		}
	}
	return result
}

func (s *memoryStore) ListByStatusBefore(status EscrowStatus, before time.Time) []*EscrowEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*EscrowEntry
	for _, e := range s.escrows {
		if e.Status == status && e.CreatedAt.Before(before) {
			result = append(result, e)
		}
	}
	return result
}

func (s *memoryStore) ListByPeer(peerDID string) []*EscrowEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*EscrowEntry, 0, len(s.escrows))
	for _, e := range s.escrows {
		if e.BuyerDID == peerDID || e.SellerDID == peerDID {
			result = append(result, e)
		}
	}
	return result
}

func (s *memoryStore) Update(entry *EscrowEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.escrows[entry.ID]; !exists {
		return fmt.Errorf("update %q: %w", entry.ID, ErrEscrowNotFound)
	}

	entry.UpdatedAt = time.Now()
	s.escrows[entry.ID] = entry
	return nil
}

func (s *memoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.escrows[id]; !exists {
		return fmt.Errorf("delete %q: %w", id, ErrEscrowNotFound)
	}

	delete(s.escrows, id)
	return nil
}
