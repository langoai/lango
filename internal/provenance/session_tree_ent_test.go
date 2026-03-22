package provenance

import (
	"context"
	"testing"
	"time"

	"github.com/langoai/lango/internal/ent/enttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntSessionTreeStore_CRUD(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	defer client.Close()

	store := NewEntSessionTreeStore(client)
	ctx := context.Background()

	root := SessionNode{
		SessionKey: "root",
		AgentName:  "root",
		Depth:      0,
		Status:     SessionStatusActive,
		CreatedAt:  time.Now().Add(-time.Minute),
	}
	child := SessionNode{
		SessionKey: "child",
		ParentKey:  "root",
		AgentName:  "worker",
		Depth:      1,
		Status:     SessionStatusActive,
		CreatedAt:  time.Now(),
	}

	require.NoError(t, store.SaveNode(ctx, root))
	require.NoError(t, store.SaveNode(ctx, child))

	gotRoot, err := store.GetNode(ctx, "root")
	require.NoError(t, err)
	assert.Equal(t, "root", gotRoot.SessionKey)

	children, err := store.GetChildren(ctx, "root")
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, "child", children[0].SessionKey)

	list, err := store.ListAll(ctx, 10)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "child", list[0].SessionKey)

	require.NoError(t, store.UpdateStatus(ctx, "child", SessionStatusMerged, &child.CreatedAt))
	updated, err := store.GetNode(ctx, "child")
	require.NoError(t, err)
	assert.Equal(t, SessionStatusMerged, updated.Status)
	assert.NotNil(t, updated.ClosedAt)
}
