package escrow

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/langoai/lango/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEntTestEntry(id string) *EscrowEntry {
	return &EscrowEntry{
		ID:          id,
		BuyerDID:    "did:example:buyer",
		SellerDID:   "did:example:seller",
		TotalAmount: big.NewInt(1000),
		Status:      StatusPending,
		Milestones: []Milestone{
			{ID: "m1", Description: "Design", Amount: big.NewInt(400), Status: MilestonePending},
			{ID: "m2", Description: "Build", Amount: big.NewInt(600), Status: MilestonePending},
		},
		TaskID:    "task-1",
		Reason:    "test deal",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func TestEntStore_CreateAndGet(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	store := NewEntStore(client)

	entry := newEntTestEntry("escrow-1")
	require.NoError(t, store.Create(entry))

	got, err := store.Get("escrow-1")
	require.NoError(t, err)

	assert.Equal(t, "escrow-1", got.ID)
	assert.Equal(t, "did:example:buyer", got.BuyerDID)
	assert.Equal(t, "did:example:seller", got.SellerDID)
	assert.Equal(t, big.NewInt(1000), got.TotalAmount)
	assert.Equal(t, StatusPending, got.Status)
	assert.Equal(t, "task-1", got.TaskID)
	assert.Equal(t, "test deal", got.Reason)
	require.Len(t, got.Milestones, 2)
	assert.Equal(t, "m1", got.Milestones[0].ID)
	assert.Equal(t, big.NewInt(400), got.Milestones[0].Amount)
	assert.Equal(t, "m2", got.Milestones[1].ID)
	assert.Equal(t, big.NewInt(600), got.Milestones[1].Amount)
	assert.False(t, got.CreatedAt.IsZero())
	assert.False(t, got.UpdatedAt.IsZero())
}

func TestEntStore_List(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	store := NewEntStore(client)

	require.NoError(t, store.Create(newEntTestEntry("escrow-a")))
	require.NoError(t, store.Create(newEntTestEntry("escrow-b")))
	require.NoError(t, store.Create(newEntTestEntry("escrow-c")))

	list := store.List()
	assert.Len(t, list, 3)
}

func TestEntStore_ListByPeer(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	store := NewEntStore(client)

	e1 := newEntTestEntry("escrow-p1")
	e1.BuyerDID = "did:example:alice"
	e1.SellerDID = "did:example:bob"
	require.NoError(t, store.Create(e1))

	e2 := newEntTestEntry("escrow-p2")
	e2.BuyerDID = "did:example:bob"
	e2.SellerDID = "did:example:carol"
	require.NoError(t, store.Create(e2))

	e3 := newEntTestEntry("escrow-p3")
	e3.BuyerDID = "did:example:carol"
	e3.SellerDID = "did:example:dave"
	require.NoError(t, store.Create(e3))

	tests := []struct {
		give    string
		wantLen int
	}{
		{give: "did:example:bob", wantLen: 2},
		{give: "did:example:alice", wantLen: 1},
		{give: "did:example:carol", wantLen: 2},
		{give: "did:example:dave", wantLen: 1},
		{give: "did:example:unknown", wantLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := store.ListByPeer(tt.give)
			assert.Len(t, result, tt.wantLen)
		})
	}
}

func TestEntStore_Update(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	store := NewEntStore(client)

	entry := newEntTestEntry("escrow-u1")
	require.NoError(t, store.Create(entry))

	got, err := store.Get("escrow-u1")
	require.NoError(t, err)

	got.Status = StatusActive
	got.TotalAmount = big.NewInt(2000)
	got.Milestones[0].Status = MilestoneCompleted
	got.DisputeNote = "updated note"
	require.NoError(t, store.Update(got))

	updated, err := store.Get("escrow-u1")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, updated.Status)
	assert.Equal(t, big.NewInt(2000), updated.TotalAmount)
	assert.Equal(t, MilestoneCompleted, updated.Milestones[0].Status)
	assert.Equal(t, "updated note", updated.DisputeNote)
}

func TestEntStore_Delete(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	store := NewEntStore(client)

	entry := newEntTestEntry("escrow-d1")
	require.NoError(t, store.Create(entry))

	require.NoError(t, store.Delete("escrow-d1"))

	_, err := store.Get("escrow-d1")
	assert.True(t, errors.Is(err, ErrEscrowNotFound))

	assert.Empty(t, store.List())
}

func TestEntStore_OnChainTracking(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	store := NewEntStore(client)

	entry := newEntTestEntry("escrow-oc1")
	require.NoError(t, store.Create(entry))

	// SetOnChainDealID
	require.NoError(t, store.SetOnChainDealID("escrow-oc1", "deal-42"))

	// GetByOnChainDealID
	got, err := store.GetByOnChainDealID("deal-42")
	require.NoError(t, err)
	assert.Equal(t, "escrow-oc1", got.ID)

	// SetTxHash
	require.NoError(t, store.SetTxHash("escrow-oc1", "deposit", "0xabc"))
	require.NoError(t, store.SetTxHash("escrow-oc1", "release", "0xdef"))
	require.NoError(t, store.SetTxHash("escrow-oc1", "refund", "0x123"))

	// Verify by re-reading the ent record directly
	deal, err := client.EscrowDeal.Query().Only(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "0xabc", deal.DepositTxHash)
	assert.Equal(t, "0xdef", deal.ReleaseTxHash)
	assert.Equal(t, "0x123", deal.RefundTxHash)

	// Unknown field should error
	err = store.SetTxHash("escrow-oc1", "invalid", "0x999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown field")
}

func TestEntStore_Errors(t *testing.T) {
	t.Parallel()
	client := testutil.TestEntClient(t)
	store := NewEntStore(client)

	t.Run("get not found", func(t *testing.T) {
		_, err := store.Get("nonexistent")
		assert.True(t, errors.Is(err, ErrEscrowNotFound))
	})

	t.Run("update not found", func(t *testing.T) {
		entry := newEntTestEntry("nonexistent")
		err := store.Update(entry)
		assert.True(t, errors.Is(err, ErrEscrowNotFound))
	})

	t.Run("delete not found", func(t *testing.T) {
		err := store.Delete("nonexistent")
		assert.True(t, errors.Is(err, ErrEscrowNotFound))
	})

	t.Run("duplicate create", func(t *testing.T) {
		entry := newEntTestEntry("escrow-dup")
		require.NoError(t, store.Create(entry))

		err := store.Create(newEntTestEntry("escrow-dup"))
		assert.True(t, errors.Is(err, ErrEscrowExists))
	})

	t.Run("set on-chain deal ID not found", func(t *testing.T) {
		err := store.SetOnChainDealID("nonexistent", "deal-1")
		assert.True(t, errors.Is(err, ErrEscrowNotFound))
	})

	t.Run("get by on-chain deal ID not found", func(t *testing.T) {
		_, err := store.GetByOnChainDealID("nonexistent")
		assert.True(t, errors.Is(err, ErrEscrowNotFound))
	})

	t.Run("set tx hash not found", func(t *testing.T) {
		err := store.SetTxHash("nonexistent", "deposit", "0xabc")
		assert.True(t, errors.Is(err, ErrEscrowNotFound))
	})
}
