package provenance

import (
	"context"
	"sort"
	"sync"
)

var _ CheckpointStore = (*MemoryStore)(nil)

// MemoryStore is an in-memory CheckpointStore for testing.
type MemoryStore struct {
	mu          sync.RWMutex
	checkpoints map[string]Checkpoint // id -> checkpoint
}

// NewMemoryStore creates a new in-memory CheckpointStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		checkpoints: make(map[string]Checkpoint),
	}
}

func (s *MemoryStore) SaveCheckpoint(_ context.Context, cp Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkpoints[cp.ID] = cp
	return nil
}

func (s *MemoryStore) GetCheckpoint(_ context.Context, id string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cp, ok := s.checkpoints[id]
	if !ok {
		return nil, ErrCheckpointNotFound
	}
	result := cp
	return &result, nil
}

func (s *MemoryStore) ListByRun(_ context.Context, runID string) ([]Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Checkpoint
	for _, cp := range s.checkpoints {
		if cp.RunID == runID {
			result = append(result, cp)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].JournalSeq < result[j].JournalSeq
	})
	return result, nil
}

func (s *MemoryStore) ListBySession(_ context.Context, sessionKey string, limit int) ([]Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Checkpoint
	for _, cp := range s.checkpoints {
		if cp.SessionKey == sessionKey {
			result = append(result, cp)
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

func (s *MemoryStore) CountBySession(_ context.Context, sessionKey string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, cp := range s.checkpoints {
		if cp.SessionKey == sessionKey {
			count++
		}
	}
	return count, nil
}

func (s *MemoryStore) DeleteCheckpoint(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.checkpoints[id]; !ok {
		return ErrCheckpointNotFound
	}
	delete(s.checkpoints, id)
	return nil
}
