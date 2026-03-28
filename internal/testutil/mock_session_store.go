package testutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/langoai/lango/internal/session"
)

// Compile-time interface check.
var _ session.Store = (*MockSessionStore)(nil)

// MockSessionStore is a thread-safe in-memory implementation of session.Store
// for use in tests. All error fields can be set to inject failures.
type MockSessionStore struct {
	mu       sync.Mutex
	sessions map[string]*session.Session
	salts    map[string][]byte

	// Configurable error injection
	CreateErr           error
	GetErr              error
	UpdateErr           error
	DeleteErr           error
	AppendMessageErr    error
	AnnotateTimeoutErr  error
	CloseErr            error
	GetSaltErr          error
	SetSaltErr          error

	// Call counters
	createCalls           int
	getCalls              int
	updateCalls           int
	deleteCalls           int
	appendMessageCalls    int
	annotateTimeoutCalls  int
	closeCalls            int
}

// NewMockSessionStore creates a new MockSessionStore.
func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: make(map[string]*session.Session),
		salts:    make(map[string][]byte),
	}
}

func (m *MockSessionStore) Create(s *session.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalls++
	if m.CreateErr != nil {
		return m.CreateErr
	}
	cp := *s
	m.sessions[s.Key] = &cp
	return nil
}

func (m *MockSessionStore) Get(key string) (*session.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalls++
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	s, ok := m.sessions[key]
	if !ok {
		return nil, fmt.Errorf("session %q not found", key)
	}
	cp := *s
	return &cp, nil
}

func (m *MockSessionStore) Update(s *session.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCalls++
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	cp := *s
	m.sessions[s.Key] = &cp
	return nil
}

func (m *MockSessionStore) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalls++
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.sessions, key)
	return nil
}

func (m *MockSessionStore) AppendMessage(key string, msg session.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.appendMessageCalls++
	if m.AppendMessageErr != nil {
		return m.AppendMessageErr
	}
	s, ok := m.sessions[key]
	if !ok {
		return fmt.Errorf("session %q not found", key)
	}
	s.History = append(s.History, msg)
	return nil
}

func (m *MockSessionStore) AnnotateTimeout(key string, partial string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.annotateTimeoutCalls++
	if m.AnnotateTimeoutErr != nil {
		return m.AnnotateTimeoutErr
	}
	s, ok := m.sessions[key]
	if !ok {
		return fmt.Errorf("session %q not found", key)
	}
	content := "[This response was interrupted due to a timeout]"
	if partial != "" {
		content = partial + "\n\n---\n" + content
	}
	s.History = append(s.History, session.Message{
		Role:    "assistant",
		Content: content,
	})
	return nil
}

func (m *MockSessionStore) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalls++
	return m.CloseErr
}

func (m *MockSessionStore) GetSalt(name string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetSaltErr != nil {
		return nil, m.GetSaltErr
	}
	salt, ok := m.salts[name]
	if !ok {
		return nil, fmt.Errorf("salt %q not found", name)
	}
	return salt, nil
}

func (m *MockSessionStore) SetSalt(name string, salt []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.SetSaltErr != nil {
		return m.SetSaltErr
	}
	m.salts[name] = salt
	return nil
}

func (m *MockSessionStore) ListSessions(_ context.Context) ([]session.SessionSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	summaries := make([]session.SessionSummary, 0, len(m.sessions))
	for _, s := range m.sessions {
		summaries = append(summaries, session.SessionSummary{
			Key:       s.Key,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}
	return summaries, nil
}

// Inspection methods

// CreateCalls returns the number of Create calls.
func (m *MockSessionStore) CreateCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createCalls
}

// GetCalls returns the number of Get calls.
func (m *MockSessionStore) GetCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getCalls
}

// UpdateCalls returns the number of Update calls.
func (m *MockSessionStore) UpdateCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updateCalls
}

// DeleteCalls returns the number of Delete calls.
func (m *MockSessionStore) DeleteCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deleteCalls
}

// AppendMessageCalls returns the number of AppendMessage calls.
func (m *MockSessionStore) AppendMessageCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.appendMessageCalls
}

// AnnotateTimeoutCalls returns the number of AnnotateTimeout calls.
func (m *MockSessionStore) AnnotateTimeoutCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.annotateTimeoutCalls
}

// CloseCalls returns the number of Close calls.
func (m *MockSessionStore) CloseCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closeCalls
}

// SessionCount returns the number of stored sessions.
func (m *MockSessionStore) SessionCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}

// HasSession returns true if a session with the given key exists.
func (m *MockSessionStore) HasSession(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.sessions[key]
	return ok
}
