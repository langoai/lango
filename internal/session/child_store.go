package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/langoai/lango/internal/types"
)

// InMemoryChildStore implements ChildSessionStore using an in-memory map.
// It wraps an existing Store for parent session access.
type InMemoryChildStore struct {
	parent   Store
	mu       sync.RWMutex
	children map[string]*ChildSession // keyed by child session key
}

// NewInMemoryChildStore creates a new in-memory child session store.
func NewInMemoryChildStore(parent Store) *InMemoryChildStore {
	return &InMemoryChildStore{
		parent:   parent,
		children: make(map[string]*ChildSession),
	}
}

// Compile-time interface check.
var _ ChildSessionStore = (*InMemoryChildStore)(nil)

// ForkChild creates a new child session from a parent session.
func (s *InMemoryChildStore) ForkChild(parentKey, agentName string, cfg ChildSessionConfig) (*ChildSession, error) {
	child := NewChildSession(parentKey, agentName, cfg)

	// Copy inherited history from parent if requested.
	if cfg.InheritHistory > 0 {
		parentSession, err := s.parent.Get(parentKey)
		if err != nil {
			return nil, fmt.Errorf("get parent session %q: %w", parentKey, err)
		}

		history := parentSession.History
		if len(history) > cfg.InheritHistory {
			history = history[len(history)-cfg.InheritHistory:]
		}

		// Deep copy messages to avoid shared slice mutations.
		child.History = make([]Message, len(history))
		copy(child.History, history)
	}

	s.mu.Lock()
	s.children[child.Key] = child
	s.mu.Unlock()

	return child, nil
}

// MergeChild merges a child session's messages back into the parent.
func (s *InMemoryChildStore) MergeChild(childKey string, summary string) error {
	s.mu.Lock()
	child, ok := s.children[childKey]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("child session %q not found", childKey)
	}
	if child.IsMerged() {
		s.mu.Unlock()
		return fmt.Errorf("child session %q already merged", childKey)
	}
	child.MergedAt = time.Now()
	s.mu.Unlock()

	// Determine what to append to parent.
	if summary != "" {
		// Append a single summary message instead of full history.
		return s.parent.AppendMessage(child.ParentKey, Message{
			Role:      types.RoleAssistant,
			Content:   summary,
			Timestamp: time.Now(),
			Author:    child.AgentName,
		})
	}

	// Append all child messages to parent.
	for _, msg := range child.History {
		if err := s.parent.AppendMessage(child.ParentKey, msg); err != nil {
			return fmt.Errorf("append child message to parent: %w", err)
		}
	}
	return nil
}

// DiscardChild removes a child session without merging.
func (s *InMemoryChildStore) DiscardChild(childKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.children[childKey]; !ok {
		return fmt.Errorf("child session %q not found", childKey)
	}
	delete(s.children, childKey)
	return nil
}

// GetChild retrieves a child session by key.
func (s *InMemoryChildStore) GetChild(childKey string) (*ChildSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	child, ok := s.children[childKey]
	if !ok {
		return nil, fmt.Errorf("child session %q not found", childKey)
	}
	return child, nil
}

// ChildrenOf returns all child sessions for a parent.
func (s *InMemoryChildStore) ChildrenOf(parentKey string) ([]*ChildSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ChildSession
	for _, child := range s.children {
		if child.ParentKey == parentKey {
			result = append(result, child)
		}
	}
	return result, nil
}
