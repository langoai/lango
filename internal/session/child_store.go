package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/langoai/lango/internal/types"
)

// SessionLifecycleEvent describes a session lifecycle transition.
type SessionLifecycleEvent struct {
	Type       string // "fork", "merge", "discard"
	ChildKey   string
	ParentKey  string
	AgentName  string
}

// ChildStoreOption configures an InMemoryChildStore.
type ChildStoreOption func(*InMemoryChildStore)

// WithLifecycleHook registers a callback invoked after fork/merge/discard operations.
func WithLifecycleHook(h func(SessionLifecycleEvent)) ChildStoreOption {
	return func(s *InMemoryChildStore) {
		s.lifecycleHook = h
	}
}

// InMemoryChildStore implements ChildSessionStore using an in-memory map.
// It wraps an existing Store for parent session access.
type InMemoryChildStore struct {
	parent        Store
	mu            sync.RWMutex
	children      map[string]*ChildSession // keyed by child session key
	parentIndex   map[string][]string      // parent key -> child keys
	lifecycleHook func(SessionLifecycleEvent)
}

// NewInMemoryChildStore creates a new in-memory child session store.
func NewInMemoryChildStore(parent Store, opts ...ChildStoreOption) *InMemoryChildStore {
	s := &InMemoryChildStore{
		parent:      parent,
		children:    make(map[string]*ChildSession),
		parentIndex: make(map[string][]string),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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
	s.parentIndex[child.ParentKey] = append(s.parentIndex[child.ParentKey], child.Key)
	s.mu.Unlock()

	if s.lifecycleHook != nil {
		s.lifecycleHook(SessionLifecycleEvent{
			Type:      "fork",
			ChildKey:  child.Key,
			ParentKey: parentKey,
			AgentName: agentName,
		})
	}

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
		if err := s.parent.AppendMessage(child.ParentKey, Message{
			Role:      types.RoleAssistant,
			Content:   summary,
			Timestamp: time.Now(),
			Author:    child.AgentName,
		}); err != nil {
			return err
		}
	} else {
		// Append all child messages to parent.
		for _, msg := range child.History {
			if err := s.parent.AppendMessage(child.ParentKey, msg); err != nil {
				return fmt.Errorf("append child message to parent: %w", err)
			}
		}
	}

	if s.lifecycleHook != nil {
		s.lifecycleHook(SessionLifecycleEvent{
			Type:      "merge",
			ChildKey:  childKey,
			ParentKey: child.ParentKey,
			AgentName: child.AgentName,
		})
	}
	return nil
}

// DiscardChild removes a child session without merging.
func (s *InMemoryChildStore) DiscardChild(childKey string) error {
	s.mu.Lock()

	child, ok := s.children[childKey]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("child session %q not found", childKey)
	}

	// Remove from parent index.
	parentKey := child.ParentKey
	kids := s.parentIndex[parentKey]
	for i, k := range kids {
		if k == childKey {
			s.parentIndex[parentKey] = append(kids[:i], kids[i+1:]...)
			break
		}
	}
	if len(s.parentIndex[parentKey]) == 0 {
		delete(s.parentIndex, parentKey)
	}

	agentName := child.AgentName
	delete(s.children, childKey)
	s.mu.Unlock()

	if s.lifecycleHook != nil {
		s.lifecycleHook(SessionLifecycleEvent{
			Type:      "discard",
			ChildKey:  childKey,
			ParentKey: parentKey,
			AgentName: agentName,
		})
	}
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

// ChildrenOf returns all child sessions for a parent using the parent index.
func (s *InMemoryChildStore) ChildrenOf(parentKey string) ([]*ChildSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := s.parentIndex[parentKey]
	result := make([]*ChildSession, 0, len(keys))
	for _, k := range keys {
		if child, ok := s.children[k]; ok {
			result = append(result, child)
		}
	}
	return result, nil
}
