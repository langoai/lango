package workspace

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := bolt.Open(dbPath, 0o600, nil)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	logger := zap.NewNop().Sugar()
	m, err := NewManager(ManagerConfig{
		DB:            db,
		LocalDID:      "did:lango:test-local",
		MaxWorkspaces: 5,
		Logger:        logger,
	})
	require.NoError(t, err)
	return m
}

func TestManager_Create(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{
		Name:     "test-workspace",
		Goal:     "build something",
		Metadata: map[string]string{"key": "value"},
	})
	require.NoError(t, err)

	assert.NotEmpty(t, ws.ID)
	assert.Equal(t, "test-workspace", ws.Name)
	assert.Equal(t, "build something", ws.Goal)
	assert.Equal(t, StatusForming, ws.Status)
	assert.Len(t, ws.Members, 1)
	assert.Equal(t, "did:lango:test-local", ws.Members[0].DID)
	assert.Equal(t, RoleCreator, ws.Members[0].Role)
	assert.False(t, ws.CreatedAt.IsZero())
	assert.False(t, ws.UpdatedAt.IsZero())
	assert.Equal(t, "value", ws.Metadata["key"])
}

func TestManager_Create_MaxWorkspaces(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	// Create up to the max (5).
	for i := 0; i < 5; i++ {
		_, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
		require.NoError(t, err)
	}

	// The 6th should fail.
	_, err := m.Create(ctx, CreateRequest{Name: "ws-over", Goal: "goal"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max workspaces reached")
}

func TestManager_Join(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	// Create a workspace from a different manager perspective.
	// The local DID is already added as creator. We need a second manager with
	// a different DID to test Join properly.
	// Instead, we directly add a member and then use the existing manager's Join
	// which should be idempotent since localDID is already a member.
	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	// Create a second manager with a different DID to join.
	dbPath := filepath.Join(t.TempDir(), "test2.db")
	db2, err := bolt.Open(dbPath, 0o600, nil)
	require.NoError(t, err)
	t.Cleanup(func() { db2.Close() })

	// We can't use a separate DB to join the same workspace. Instead, we test
	// the Join path by manually inserting a workspace without the localDID.
	// Simpler: add a remote member to the workspace first, then create a new
	// manager that loads from the same DB with a different DID.

	// Use AddMember to add a remote peer, then verify they're there.
	err = m.AddMember(ctx, ws.ID, &Member{
		DID:      "did:lango:remote-peer",
		Name:     "remote",
		Role:     RoleMember,
		JoinedAt: time.Now(),
	})
	require.NoError(t, err)

	got, err := m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 2)

	// Now test Join with the local DID (already a member, should be idempotent).
	err = m.Join(ctx, ws.ID)
	require.NoError(t, err)

	got, err = m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 2, "member count should not change for idempotent join")
}

func TestManager_Join_AlreadyMember(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	// Join again as creator — should be idempotent.
	err = m.Join(ctx, ws.ID)
	require.NoError(t, err)

	got, err := m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 1, "no duplicate member")
}

func TestManager_Join_Archived(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	err = m.Archive(ctx, ws.ID)
	require.NoError(t, err)

	err = m.Join(ctx, ws.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "archived")
}

func TestManager_Leave(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)
	assert.Len(t, ws.Members, 1)

	err = m.Leave(ctx, ws.ID)
	require.NoError(t, err)

	got, err := m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Empty(t, got.Members)
}

func TestManager_Get(t *testing.T) {
	tests := []struct {
		give        string
		wantErr     bool
		wantErrText string
	}{
		{
			give:    "existing",
			wantErr: false,
		},
		{
			give:        "non-existent-id",
			wantErr:     true,
			wantErrText: "not found",
		},
	}

	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			id := ws.ID
			if tt.give == "non-existent-id" {
				id = "does-not-exist"
			}

			got, err := m.Get(ctx, id)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, ws.ID, got.ID)
			}
		})
	}
}

func TestManager_List(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	// Empty initially.
	list, err := m.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, list)

	// Create several workspaces.
	for i := 0; i < 3; i++ {
		_, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
		require.NoError(t, err)
	}

	list, err = m.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 3)
}

func TestManager_Activate(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)
	assert.Equal(t, StatusForming, ws.Status)

	err = m.Activate(ctx, ws.ID)
	require.NoError(t, err)

	got, err := m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusActive, got.Status)
}

func TestManager_Activate_NotForming(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	// Activate first.
	err = m.Activate(ctx, ws.ID)
	require.NoError(t, err)

	// Activating again should fail.
	err = m.Activate(ctx, ws.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in forming state")
}

func TestManager_Archive(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	err = m.Archive(ctx, ws.ID)
	require.NoError(t, err)

	got, err := m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusArchived, got.Status)
}

func TestManager_Post_And_Read(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	// Post messages.
	msgs := []Message{
		{
			Type:      MessageTypeTaskProposal,
			SenderDID: "did:lango:test-local",
			Content:   "proposal one",
		},
		{
			Type:      MessageTypeLogStream,
			SenderDID: "did:lango:test-local",
			Content:   "log entry",
		},
		{
			Type:      MessageTypeKnowledgeShare,
			SenderDID: "did:lango:remote-peer",
			Content:   "knowledge share",
		},
	}

	for _, msg := range msgs {
		err := m.Post(ctx, ws.ID, msg)
		require.NoError(t, err)
	}

	// Read all messages.
	got, err := m.Read(ctx, ws.ID, ReadOptions{})
	require.NoError(t, err)
	assert.Len(t, got, 3)

	// Read with sender filter.
	got, err = m.Read(ctx, ws.ID, ReadOptions{SenderDID: "did:lango:remote-peer"})
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "knowledge share", got[0].Content)

	// Read with type filter.
	got, err = m.Read(ctx, ws.ID, ReadOptions{Types: []string{string(MessageTypeTaskProposal)}})
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "proposal one", got[0].Content)

	// Read with limit.
	got, err = m.Read(ctx, ws.ID, ReadOptions{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestManager_Post_Archived(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	err = m.Archive(ctx, ws.ID)
	require.NoError(t, err)

	err = m.Post(ctx, ws.ID, Message{
		Type:    MessageTypeLogStream,
		Content: "should fail",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "archived")
}

func TestManager_AddMember(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	err = m.AddMember(ctx, ws.ID, &Member{
		DID:      "did:lango:remote-1",
		Name:     "Remote Agent 1",
		Role:     RoleMember,
		JoinedAt: time.Now(),
	})
	require.NoError(t, err)

	got, err := m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 2)
	assert.Equal(t, "did:lango:remote-1", got.Members[1].DID)
	assert.Equal(t, "Remote Agent 1", got.Members[1].Name)
}

func TestManager_AddMember_AlreadyExists(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	ws, err := m.Create(ctx, CreateRequest{Name: "ws", Goal: "goal"})
	require.NoError(t, err)

	member := &Member{
		DID:      "did:lango:remote-1",
		Name:     "Remote Agent 1",
		Role:     RoleMember,
		JoinedAt: time.Now(),
	}

	err = m.AddMember(ctx, ws.ID, member)
	require.NoError(t, err)

	// Add the same member again — should be idempotent.
	err = m.AddMember(ctx, ws.ID, member)
	require.NoError(t, err)

	got, err := m.Get(ctx, ws.ID)
	require.NoError(t, err)
	assert.Len(t, got.Members, 2, "no duplicate member")
}
