package approval

import (
	"sync"
	"time"
)

// HistoryEntry records a single approval decision.
type HistoryEntry struct {
	Timestamp   time.Time
	RequestID   string
	ToolName    string
	SessionKey  string
	Summary     string
	SafetyLevel string
	Outcome     string // open set: "bypass", "granted", "denied", "timeout", "replay_blocked", "unavailable", etc.
	Provider    string
}

// HistoryStore is an append-only in-memory store for approval decisions.
// It uses a ring buffer capped at maxSize entries. Safe for concurrent use.
type HistoryStore struct {
	mu      sync.RWMutex
	entries []HistoryEntry
	maxSize int
}

// NewHistoryStore creates a history store with the given capacity.
// When the capacity is reached, the oldest entry is evicted.
func NewHistoryStore(maxSize int) *HistoryStore {
	if maxSize <= 0 {
		maxSize = 500
	}
	return &HistoryStore{
		entries: make([]HistoryEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Append adds an entry. If at capacity, the oldest entry is evicted.
func (s *HistoryStore) Append(entry HistoryEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) >= s.maxSize {
		// Shift left by 1 to evict oldest
		copy(s.entries, s.entries[1:])
		s.entries = s.entries[:len(s.entries)-1]
	}
	s.entries = append(s.entries, entry)
}

// List returns all entries in newest-first order.
// The returned slice is a copy safe for concurrent use.
func (s *HistoryStore) List() []HistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.entries) == 0 {
		return nil
	}
	result := make([]HistoryEntry, len(s.entries))
	// Reverse copy: newest first
	for i, e := range s.entries {
		result[len(s.entries)-1-i] = e
	}
	return result
}

// Count returns the number of stored entries.
func (s *HistoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// CountByOutcome returns a map of outcome to count.
func (s *HistoryStore) CountByOutcome() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]int)
	for _, e := range s.entries {
		result[e.Outcome]++
	}
	return result
}
