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

func defaultPolicy(d time.Duration) sa.SessionPolicy {
	now := time.Now()
	return sa.SessionPolicy{
		AllowedTargets:   []common.Address{common.HexToAddress("0xaaaa")},
		AllowedFunctions: []string{"0x12345678"},
		SpendLimit:       big.NewInt(1000),
		ValidAfter:       now,
		ValidUntil:       now.Add(d),
	}
}

func TestManager_Create_MasterSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	policy := defaultPolicy(1 * time.Hour)
	sk, err := mgr.Create(ctx, policy, "")
	require.NoError(t, err)

	assert.NotEmpty(t, sk.ID)
	assert.True(t, sk.IsMaster())
	assert.True(t, sk.IsActive())
	assert.NotEmpty(t, sk.PublicKey)
}

func TestManager_Create_TaskSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	// Create parent.
	parentPolicy := defaultPolicy(2 * time.Hour)
	parent, err := mgr.Create(ctx, parentPolicy, "")
	require.NoError(t, err)

	// Create child with wider bounds — should be tightened.
	childPolicy := sa.SessionPolicy{
		AllowedTargets: []common.Address{
			common.HexToAddress("0xaaaa"),
			common.HexToAddress("0xbbbb"),
		},
		AllowedFunctions: []string{"0x12345678", "0xabcdef00"},
		SpendLimit:       big.NewInt(5000),
		ValidAfter:       time.Now().Add(-2 * time.Hour),
		ValidUntil:       time.Now().Add(4 * time.Hour),
	}

	child, err := mgr.Create(ctx, childPolicy, parent.ID)
	require.NoError(t, err)

	assert.Equal(t, parent.ID, child.ParentID)
	assert.False(t, child.IsMaster())

	// SpendLimit should be tightened to parent's.
	assert.Equal(t, 0, child.Policy.SpendLimit.Cmp(big.NewInt(1000)))

	// AllowedTargets should be intersected.
	assert.Len(t, child.Policy.AllowedTargets, 1)
	assert.Equal(t, common.HexToAddress("0xaaaa"), child.Policy.AllowedTargets[0])

	// AllowedFunctions should be intersected.
	assert.Len(t, child.Policy.AllowedFunctions, 1)
	assert.Equal(t, "0x12345678", child.Policy.AllowedFunctions[0])
}

func TestManager_Create_ParentNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	_, err := mgr.Create(ctx, defaultPolicy(time.Hour), "nonexistent")
	assert.ErrorIs(t, err, sa.ErrSessionNotFound)
}

func TestManager_Create_ParentExpired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store, WithMaxDuration(24*time.Hour))

	// Create an already expired parent directly in store.
	parent := makeSessionKey("parent", "", false, time.Now().Add(-time.Minute))
	require.NoError(t, store.Save(ctx, parent))

	_, err := mgr.Create(ctx, defaultPolicy(time.Hour), "parent")
	assert.ErrorIs(t, err, sa.ErrSessionExpired)
}

func TestManager_Create_ParentRevoked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	parent := makeSessionKey("parent", "", true, time.Now().Add(time.Hour))
	require.NoError(t, store.Save(ctx, parent))

	_, err := mgr.Create(ctx, defaultPolicy(time.Hour), "parent")
	assert.ErrorIs(t, err, sa.ErrSessionRevoked)
}

func TestManager_Create_ExceedMaxDuration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store, WithMaxDuration(1*time.Hour))

	policy := defaultPolicy(2 * time.Hour) // exceeds 1h max
	_, err := mgr.Create(ctx, policy, "")
	assert.ErrorIs(t, err, sa.ErrPolicyViolation)
}

func TestManager_Create_ExceedMaxKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store, WithMaxKeys(2), WithMaxDuration(24*time.Hour))

	policy := defaultPolicy(1 * time.Hour)
	_, err := mgr.Create(ctx, policy, "")
	require.NoError(t, err)
	_, err = mgr.Create(ctx, policy, "")
	require.NoError(t, err)

	// Third should fail.
	_, err = mgr.Create(ctx, policy, "")
	assert.ErrorIs(t, err, sa.ErrPolicyViolation)
}

func TestManager_Create_WithOnChainRegistration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	registered := false
	regFn := func(_ context.Context, _ common.Address, _ sa.SessionPolicy) (string, error) {
		registered = true
		return "0xtxhash", nil
	}

	mgr := NewManager(store, WithOnChainRegistration(regFn))
	_, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
	require.NoError(t, err)
	assert.True(t, registered)
}

func TestManager_Revoke(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	sk, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
	require.NoError(t, err)

	err = mgr.Revoke(ctx, sk.ID)
	require.NoError(t, err)

	got, err := mgr.Get(ctx, sk.ID)
	require.NoError(t, err)
	assert.True(t, got.Revoked)
}

func TestManager_Revoke_CascadesToChildren(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store, WithMaxDuration(24*time.Hour))

	parent, err := mgr.Create(ctx, defaultPolicy(2*time.Hour), "")
	require.NoError(t, err)

	child, err := mgr.Create(ctx, defaultPolicy(1*time.Hour), parent.ID)
	require.NoError(t, err)

	// Revoke parent should cascade to child.
	err = mgr.Revoke(ctx, parent.ID)
	require.NoError(t, err)

	gotChild, err := mgr.Get(ctx, child.ID)
	require.NoError(t, err)
	assert.True(t, gotChild.Revoked)
}

func TestManager_Revoke_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	err := mgr.Revoke(ctx, "nonexistent")
	assert.ErrorIs(t, err, sa.ErrSessionNotFound)
}

func TestManager_Revoke_WithOnChainCallback(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()

	revokedAddrs := make([]common.Address, 0)
	revFn := func(_ context.Context, addr common.Address) (string, error) {
		revokedAddrs = append(revokedAddrs, addr)
		return "0xtxhash", nil
	}

	mgr := NewManager(store, WithOnChainRevocation(revFn))
	sk, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
	require.NoError(t, err)

	err = mgr.Revoke(ctx, sk.ID)
	require.NoError(t, err)
	assert.Len(t, revokedAddrs, 1)
}

func TestManager_RevokeAll(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store, WithMaxKeys(5))

	for range 3 {
		_, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
		require.NoError(t, err)
	}

	err := mgr.RevokeAll(ctx)
	require.NoError(t, err)

	active, err := store.ListActive(ctx)
	require.NoError(t, err)
	assert.Empty(t, active)
}

func TestManager_CleanupExpired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	// Insert expired and active keys directly.
	expired := makeSessionKey("exp1", "", false, time.Now().Add(-time.Hour))
	active := makeSessionKey("act1", "", false, time.Now().Add(time.Hour))
	require.NoError(t, store.Save(ctx, expired))
	require.NoError(t, store.Save(ctx, active))

	removed, err := mgr.CleanupExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	// Only active key remains.
	all, err := store.List(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, "act1", all[0].ID)
}

func TestManager_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store)

	sk, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
	require.NoError(t, err)

	got, err := mgr.Get(ctx, sk.ID)
	require.NoError(t, err)
	assert.Equal(t, sk.ID, got.ID)
}

func TestManager_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := NewMemoryStore()
	mgr := NewManager(store, WithMaxKeys(5))

	for range 3 {
		_, err := mgr.Create(ctx, defaultPolicy(time.Hour), "")
		require.NoError(t, err)
	}

	list, err := mgr.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 3)
}
