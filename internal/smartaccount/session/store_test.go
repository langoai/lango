package session

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sa "github.com/langoai/lango/internal/smartaccount"
)

func makeSessionKey(id, parentID string, revoked bool, expiresAt time.Time) *sa.SessionKey {
	return &sa.SessionKey{
		ID:        id,
		PublicKey: []byte{0x01, 0x02},
		Address:   common.HexToAddress("0x1234"),
		ParentID:  parentID,
		Policy: sa.SessionPolicy{
			AllowedTargets:   []common.Address{common.HexToAddress("0xaaaa")},
			AllowedFunctions: []string{"0x12345678"},
			SpendLimit:       big.NewInt(1000),
			ValidAfter:       time.Now().Add(-1 * time.Hour),
			ValidUntil:       expiresAt,
		},
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		Revoked:   revoked,
	}
}

func TestMemoryStore_Save(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	key := makeSessionKey("k1", "", false, time.Now().Add(time.Hour))
	err := store.Save(ctx, key)
	require.NoError(t, err)

	got, err := store.Get(ctx, "k1")
	require.NoError(t, err)
	assert.Equal(t, "k1", got.ID)
}

func TestMemoryStore_Save_Overwrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	key := makeSessionKey("k1", "", false, time.Now().Add(time.Hour))
	require.NoError(t, store.Save(ctx, key))

	key.Revoked = true
	require.NoError(t, store.Save(ctx, key))

	got, err := store.Get(ctx, "k1")
	require.NoError(t, err)
	assert.True(t, got.Revoked)
}

func TestMemoryStore_Save_IsolatesCopy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	key := makeSessionKey("k1", "", false, time.Now().Add(time.Hour))
	require.NoError(t, store.Save(ctx, key))

	// Mutate original should not affect stored copy.
	key.Revoked = true
	got, err := store.Get(ctx, "k1")
	require.NoError(t, err)
	assert.False(t, got.Revoked)
}

func TestMemoryStore_Get_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	_, err := store.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, sa.ErrSessionNotFound)
}

func TestMemoryStore_Get_ReturnsCopy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	key := makeSessionKey("k1", "", false, time.Now().Add(time.Hour))
	require.NoError(t, store.Save(ctx, key))

	got, err := store.Get(ctx, "k1")
	require.NoError(t, err)

	// Mutating the returned copy should not affect the store.
	got.Revoked = true
	got2, err := store.Get(ctx, "k1")
	require.NoError(t, err)
	assert.False(t, got2.Revoked)
}

func TestMemoryStore_List_SortedByCreatedAt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	now := time.Now()
	k1 := makeSessionKey("k1", "", false, now.Add(time.Hour))
	k1.CreatedAt = now.Add(2 * time.Hour)
	k2 := makeSessionKey("k2", "", false, now.Add(time.Hour))
	k2.CreatedAt = now.Add(1 * time.Hour)
	k3 := makeSessionKey("k3", "", false, now.Add(time.Hour))
	k3.CreatedAt = now

	require.NoError(t, store.Save(ctx, k1))
	require.NoError(t, store.Save(ctx, k2))
	require.NoError(t, store.Save(ctx, k3))

	list, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, "k3", list[0].ID)
	assert.Equal(t, "k2", list[1].ID)
	assert.Equal(t, "k1", list[2].ID)
}

func TestMemoryStore_List_Empty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	list, err := store.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestMemoryStore_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	key := makeSessionKey("k1", "", false, time.Now().Add(time.Hour))
	require.NoError(t, store.Save(ctx, key))

	err := store.Delete(ctx, "k1")
	require.NoError(t, err)

	_, err = store.Get(ctx, "k1")
	assert.ErrorIs(t, err, sa.ErrSessionNotFound)
}

func TestMemoryStore_Delete_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	err := store.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, sa.ErrSessionNotFound)
}

func TestMemoryStore_ListByParent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	master := makeSessionKey("master", "", false, time.Now().Add(time.Hour))
	child1 := makeSessionKey("child1", "master", false, time.Now().Add(time.Hour))
	child2 := makeSessionKey("child2", "master", false, time.Now().Add(time.Hour))
	other := makeSessionKey("other", "other-parent", false, time.Now().Add(time.Hour))

	for _, k := range []*sa.SessionKey{master, child1, child2, other} {
		require.NoError(t, store.Save(ctx, k))
	}

	children, err := store.ListByParent(ctx, "master")
	require.NoError(t, err)
	require.Len(t, children, 2)

	ids := []string{children[0].ID, children[1].ID}
	assert.Contains(t, ids, "child1")
	assert.Contains(t, ids, "child2")
}

func TestMemoryStore_ListByParent_NoResults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	result, err := store.ListByParent(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestMemoryStore_ListActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	active := makeSessionKey("active", "", false, time.Now().Add(time.Hour))
	expired := makeSessionKey("expired", "", false, time.Now().Add(-time.Hour))
	revoked := makeSessionKey("revoked", "", true, time.Now().Add(time.Hour))

	for _, k := range []*sa.SessionKey{active, expired, revoked} {
		require.NoError(t, store.Save(ctx, k))
	}

	result, err := store.ListActive(ctx)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "active", result[0].ID)
}
