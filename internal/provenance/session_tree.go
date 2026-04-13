package provenance

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SessionTreeStore is the persistence interface for session tree nodes.
type SessionTreeStore interface {
	// SaveNode persists a session node.
	SaveNode(ctx context.Context, node SessionNode) error

	// GetNode returns a node by session key.
	GetNode(ctx context.Context, sessionKey string) (*SessionNode, error)

	// GetChildren returns direct children of a session.
	GetChildren(ctx context.Context, parentKey string) ([]SessionNode, error)

	// ListAll returns all session nodes, ordered by created_at desc.
	ListAll(ctx context.Context, limit int) ([]SessionNode, error)

	// UpdateStatus updates the status and optional closed_at of a node.
	UpdateStatus(ctx context.Context, sessionKey string, status SessionStatus, closedAt *time.Time) error
}

// MemoryTreeStore is an in-memory SessionTreeStore for testing.
type MemoryTreeStore struct {
	mu    sync.RWMutex
	nodes map[string]SessionNode
}

var _ SessionTreeStore = (*MemoryTreeStore)(nil)

// NewMemoryTreeStore creates a new in-memory session tree store.
func NewMemoryTreeStore() *MemoryTreeStore {
	return &MemoryTreeStore{
		nodes: make(map[string]SessionNode),
	}
}

func (s *MemoryTreeStore) SaveNode(_ context.Context, node SessionNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nodes[node.SessionKey] = node
	return nil
}

func (s *MemoryTreeStore) GetNode(_ context.Context, sessionKey string) (*SessionNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, ok := s.nodes[sessionKey]
	if !ok {
		return nil, ErrSessionNotFound
	}
	result := node
	return &result, nil
}

func (s *MemoryTreeStore) GetChildren(_ context.Context, parentKey string) ([]SessionNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var children []SessionNode
	for _, node := range s.nodes {
		if node.ParentKey == parentKey {
			children = append(children, node)
		}
	}
	return children, nil
}

func (s *MemoryTreeStore) ListAll(_ context.Context, limit int) ([]SessionNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]SessionNode, 0, len(s.nodes))
	for _, node := range s.nodes {
		result = append(result, node)
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *MemoryTreeStore) UpdateStatus(_ context.Context, sessionKey string, status SessionStatus, closedAt *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[sessionKey]
	if !ok {
		return ErrSessionNotFound
	}
	node.Status = status
	node.ClosedAt = closedAt
	s.nodes[sessionKey] = node
	return nil
}

// SessionTree manages session hierarchy registration and queries.
type SessionTree struct {
	store SessionTreeStore
}

// NewSessionTree creates a new session tree manager.
func NewSessionTree(store SessionTreeStore) *SessionTree {
	return &SessionTree{store: store}
}

// GetNode retrieves a session node by key.
func (t *SessionTree) GetNode(ctx context.Context, sessionKey string) (*SessionNode, error) {
	return t.store.GetNode(ctx, sessionKey)
}

// RegisterSession creates a new node in the session tree.
func (t *SessionTree) RegisterSession(ctx context.Context, sessionKey, parentKey, agentName, goal string) (*SessionNode, error) {
	if sessionKey == "" {
		return nil, ErrInvalidSessionKey
	}

	depth := 0
	if parentKey != "" {
		parent, err := t.store.GetNode(ctx, parentKey)
		if err != nil {
			return nil, fmt.Errorf("get parent session: %w", err)
		}
		depth = parent.Depth + 1
	}

	node := SessionNode{
		SessionKey: sessionKey,
		ParentKey:  parentKey,
		AgentName:  agentName,
		Goal:       goal,
		Depth:      depth,
		Status:     SessionStatusActive,
		CreatedAt:  time.Now(),
	}

	if err := t.store.SaveNode(ctx, node); err != nil {
		return nil, fmt.Errorf("save session node: %w", err)
	}
	return &node, nil
}

// CloseSession marks a session as completed, merged, or discarded.
func (t *SessionTree) CloseSession(ctx context.Context, sessionKey string, status SessionStatus) error {
	now := time.Now()
	return t.store.UpdateStatus(ctx, sessionKey, status, &now)
}

// GetTree returns the subtree rooted at the given session, up to maxDepth levels.
func (t *SessionTree) GetTree(ctx context.Context, rootKey string, maxDepth int) ([]SessionNode, error) {
	root, err := t.store.GetNode(ctx, rootKey)
	if err != nil {
		return nil, err
	}

	var result []SessionNode
	result = append(result, *root)

	if maxDepth <= 0 {
		return result, nil
	}

	return t.collectChildren(ctx, rootKey, maxDepth, result)
}

func (t *SessionTree) collectChildren(ctx context.Context, parentKey string, remainingDepth int, acc []SessionNode) ([]SessionNode, error) {
	if remainingDepth <= 0 {
		return acc, nil
	}

	children, err := t.store.GetChildren(ctx, parentKey)
	if err != nil {
		return acc, err
	}

	for _, child := range children {
		acc = append(acc, child)
		acc, err = t.collectChildren(ctx, child.SessionKey, remainingDepth-1, acc)
		if err != nil {
			return acc, err
		}
	}
	return acc, nil
}
