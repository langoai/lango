package provenance

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionTree_RegisterRoot(t *testing.T) {
	store := NewMemoryTreeStore()
	tree := NewSessionTree(store)
	ctx := context.Background()

	node, err := tree.RegisterSession(ctx, "root-1", "", "orchestrator", "main task")
	require.NoError(t, err)
	assert.Equal(t, "root-1", node.SessionKey)
	assert.Empty(t, node.ParentKey)
	assert.Equal(t, 0, node.Depth)
	assert.Equal(t, SessionStatusActive, node.Status)
}

func TestSessionTree_RegisterChild(t *testing.T) {
	store := NewMemoryTreeStore()
	tree := NewSessionTree(store)
	ctx := context.Background()

	_, err := tree.RegisterSession(ctx, "root-1", "", "orchestrator", "main task")
	require.NoError(t, err)

	child, err := tree.RegisterSession(ctx, "child-1", "root-1", "researcher", "research subtask")
	require.NoError(t, err)
	assert.Equal(t, "root-1", child.ParentKey)
	assert.Equal(t, 1, child.Depth)
}

func TestSessionTree_RegisterEmptyKey(t *testing.T) {
	store := NewMemoryTreeStore()
	tree := NewSessionTree(store)

	_, err := tree.RegisterSession(context.Background(), "", "", "agent", "goal")
	assert.ErrorIs(t, err, ErrInvalidSessionKey)
}

func TestSessionTree_CloseSession(t *testing.T) {
	store := NewMemoryTreeStore()
	tree := NewSessionTree(store)
	ctx := context.Background()

	_, err := tree.RegisterSession(ctx, "sess-1", "", "agent", "goal")
	require.NoError(t, err)

	require.NoError(t, tree.CloseSession(ctx, "sess-1", SessionStatusCompleted))

	node, err := store.GetNode(ctx, "sess-1")
	require.NoError(t, err)
	assert.Equal(t, SessionStatusCompleted, node.Status)
	assert.NotNil(t, node.ClosedAt)
}

func TestSessionTree_GetTree(t *testing.T) {
	store := NewMemoryTreeStore()
	tree := NewSessionTree(store)
	ctx := context.Background()

	_, err := tree.RegisterSession(ctx, "root", "", "orchestrator", "main")
	require.NoError(t, err)
	_, err = tree.RegisterSession(ctx, "child-1", "root", "researcher", "research")
	require.NoError(t, err)
	_, err = tree.RegisterSession(ctx, "child-2", "root", "executor", "execute")
	require.NoError(t, err)
	_, err = tree.RegisterSession(ctx, "grandchild-1", "child-1", "worker", "work")
	require.NoError(t, err)

	// Full tree.
	nodes, err := tree.GetTree(ctx, "root", 10)
	require.NoError(t, err)
	assert.Len(t, nodes, 4)

	// Limited depth.
	nodes, err = tree.GetTree(ctx, "root", 1)
	require.NoError(t, err)
	assert.Len(t, nodes, 3) // root + 2 direct children
}

func TestSessionTree_GetTree_NotFound(t *testing.T) {
	store := NewMemoryTreeStore()
	tree := NewSessionTree(store)

	_, err := tree.GetTree(context.Background(), "nonexistent", 10)
	assert.ErrorIs(t, err, ErrSessionNotFound)
}
