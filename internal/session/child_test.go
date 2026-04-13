package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/types"
)

// mockStore implements Store for testing child sessions.
type mockStore struct {
	sessions map[string]*Session
}

func newMockStore() *mockStore {
	return &mockStore{sessions: make(map[string]*Session)}
}

func (m *mockStore) Create(s *Session) error          { m.sessions[s.Key] = s; return nil }
func (m *mockStore) Get(key string) (*Session, error) { return m.sessions[key], nil }
func (m *mockStore) Update(s *Session) error          { m.sessions[s.Key] = s; return nil }
func (m *mockStore) Delete(key string) error          { delete(m.sessions, key); return nil }
func (m *mockStore) Close() error                     { return nil }
func (m *mockStore) GetSalt(_ string) ([]byte, error) { return nil, nil }
func (m *mockStore) SetSalt(_ string, _ []byte) error                          { return nil }
func (m *mockStore) ListSessions(_ context.Context) ([]SessionSummary, error)  { return nil, nil }

func (m *mockStore) AppendMessage(key string, msg Message) error {
	s := m.sessions[key]
	if s == nil {
		return nil
	}
	s.History = append(s.History, msg)
	return nil
}

func (m *mockStore) AnnotateTimeout(_ string, _ string) error { return nil }

func TestNewChildSession(t *testing.T) {
	t.Parallel()
	cs := NewChildSession("parent-1", "operator", ChildSessionConfig{
		MaxMessages: 100,
	})

	assert.Contains(t, cs.Key, "parent-1:child:operator:")
	assert.Equal(t, "parent-1", cs.ParentKey)
	assert.Equal(t, "operator", cs.AgentName)
	assert.False(t, cs.IsMerged())
	assert.Empty(t, cs.History)
}

func TestChildSession_AppendMessage(t *testing.T) {
	t.Parallel()
	cs := NewChildSession("p1", "agent", ChildSessionConfig{MaxMessages: 3})

	for i := 0; i < 5; i++ {
		cs.AppendMessage(Message{
			Role:    types.RoleUser,
			Content: "msg",
		})
	}

	assert.Len(t, cs.History, 3, "should enforce MaxMessages limit")
}

func TestChildSession_AppendMessage_Unlimited(t *testing.T) {
	t.Parallel()
	cs := NewChildSession("p1", "agent", ChildSessionConfig{})

	for i := 0; i < 10; i++ {
		cs.AppendMessage(Message{Role: types.RoleUser, Content: "msg"})
	}

	assert.Len(t, cs.History, 10, "no limit means all messages kept")
}

func TestInMemoryChildStore_ForkChild(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	_ = store.Create(&Session{
		Key: "parent-1",
		History: []Message{
			{Role: types.RoleUser, Content: "hello"},
			{Role: types.RoleAssistant, Content: "hi"},
			{Role: types.RoleUser, Content: "how are you"},
		},
	})

	cs := NewInMemoryChildStore(store)

	tests := []struct {
		name           string
		giveInherit    int
		wantHistoryLen int
	}{
		{
			name:           "no inheritance",
			giveInherit:    0,
			wantHistoryLen: 0,
		},
		{
			name:           "inherit last 2",
			giveInherit:    2,
			wantHistoryLen: 2,
		},
		{
			name:           "inherit more than available",
			giveInherit:    10,
			wantHistoryLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			child, err := cs.ForkChild("parent-1", "test-agent", ChildSessionConfig{
				InheritHistory: tt.giveInherit,
			})
			require.NoError(t, err)
			assert.Len(t, child.History, tt.wantHistoryLen)
		})
	}
}

func TestInMemoryChildStore_MergeChild_Summary(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	_ = store.Create(&Session{Key: "parent-1"})

	cs := NewInMemoryChildStore(store)

	child, err := cs.ForkChild("parent-1", "operator", ChildSessionConfig{})
	require.NoError(t, err)

	child.AppendMessage(Message{Role: types.RoleUser, Content: "do something"})
	child.AppendMessage(Message{Role: types.RoleAssistant, Content: "done"})

	err = cs.MergeChild(child.Key, "Operator completed the task successfully.")
	require.NoError(t, err)

	parent := store.sessions["parent-1"]
	require.Len(t, parent.History, 1)
	assert.Equal(t, "Operator completed the task successfully.", parent.History[0].Content)
	assert.Equal(t, "operator", parent.History[0].Author)
}

func TestInMemoryChildStore_MergeChild_FullHistory(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	_ = store.Create(&Session{Key: "parent-1"})

	cs := NewInMemoryChildStore(store)

	child, err := cs.ForkChild("parent-1", "operator", ChildSessionConfig{})
	require.NoError(t, err)

	child.AppendMessage(Message{Role: types.RoleUser, Content: "msg1"})
	child.AppendMessage(Message{Role: types.RoleAssistant, Content: "msg2"})

	err = cs.MergeChild(child.Key, "")
	require.NoError(t, err)

	parent := store.sessions["parent-1"]
	assert.Len(t, parent.History, 2)
}

func TestInMemoryChildStore_MergeChild_AlreadyMerged(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	_ = store.Create(&Session{Key: "parent-1"})

	cs := NewInMemoryChildStore(store)

	child, err := cs.ForkChild("parent-1", "operator", ChildSessionConfig{})
	require.NoError(t, err)

	err = cs.MergeChild(child.Key, "summary")
	require.NoError(t, err)

	err = cs.MergeChild(child.Key, "summary again")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already merged")
}

func TestInMemoryChildStore_DiscardChild(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	_ = store.Create(&Session{Key: "parent-1"})

	cs := NewInMemoryChildStore(store)

	child, err := cs.ForkChild("parent-1", "operator", ChildSessionConfig{})
	require.NoError(t, err)

	err = cs.DiscardChild(child.Key)
	require.NoError(t, err)

	_, err = cs.GetChild(child.Key)
	require.Error(t, err)
}

func TestInMemoryChildStore_ChildrenOf(t *testing.T) {
	t.Parallel()
	store := newMockStore()
	_ = store.Create(&Session{Key: "parent-1"})
	_ = store.Create(&Session{Key: "parent-2"})

	cs := NewInMemoryChildStore(store)

	_, _ = cs.ForkChild("parent-1", "agent-a", ChildSessionConfig{})
	_, _ = cs.ForkChild("parent-1", "agent-b", ChildSessionConfig{})
	_, _ = cs.ForkChild("parent-2", "agent-c", ChildSessionConfig{})

	children, err := cs.ChildrenOf("parent-1")
	require.NoError(t, err)
	assert.Len(t, children, 2)

	children2, err := cs.ChildrenOf("parent-2")
	require.NoError(t, err)
	assert.Len(t, children2, 1)
}

func TestChildSession_IsMerged(t *testing.T) {
	t.Parallel()
	cs := NewChildSession("p1", "agent", ChildSessionConfig{})
	assert.False(t, cs.IsMerged())

	cs.MergedAt = time.Now()
	assert.True(t, cs.IsMerged())
}
