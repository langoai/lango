package app

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
)

// bridgeTestSetup creates a bus, escrow engine, and wires the bridge.
func bridgeTestSetup(t *testing.T) (*eventbus.Bus, *escrow.Engine, escrow.Store) {
	t.Helper()
	bus := eventbus.New()
	store := escrow.NewMemoryStore()
	settler := escrow.NoopSettler{}
	cfg := escrow.DefaultEngineConfig()
	engine := escrow.NewEngine(store, settler, cfg)
	initOnChainEscrowBridge(bus, engine, zap.NewNop().Sugar())
	return bus, engine, store
}

// createTestEscrow creates a pending escrow for testing.
func createTestEscrow(t *testing.T, engine *escrow.Engine) *escrow.EscrowEntry {
	t.Helper()
	entry, err := engine.Create(context.Background(), escrow.CreateRequest{
		BuyerDID:  "did:test:buyer",
		SellerDID: "did:test:seller",
		Amount:    big.NewInt(1000),
		Reason:    "test",
		Milestones: []escrow.MilestoneRequest{
			{Description: "milestone1", Amount: big.NewInt(1000)},
		},
		ExpiresAt: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	})
	require.NoError(t, err)
	return entry
}

func TestBridge_DepositEvent_FundsAndActivates(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)
	assert.Equal(t, escrow.StatusPending, entry.Status)

	bus.Publish(eventbus.EscrowOnChainDepositEvent{
		EscrowID: entry.ID,
		DealID:   "1",
		Buyer:    "0xBuyer",
		Amount:   big.NewInt(1000),
		TxHash:   "0xdeposittx",
	})

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusActive, updated.Status)
}

func TestBridge_DepositEvent_EmptyEscrowID(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	bus.Publish(eventbus.EscrowOnChainDepositEvent{
		EscrowID: "",
		DealID:   "1",
		Buyer:    "0xBuyer",
		Amount:   big.NewInt(1000),
		TxHash:   "0xtx",
	})

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusPending, updated.Status)
}

func TestBridge_DepositEvent_Idempotent(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	ev := eventbus.EscrowOnChainDepositEvent{
		EscrowID: entry.ID,
		DealID:   "1",
		Buyer:    "0xBuyer",
		Amount:   big.NewInt(1000),
		TxHash:   "0xtx",
	}
	bus.Publish(ev)
	bus.Publish(ev) // second time should not panic or error

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusActive, updated.Status)
}

func TestBridge_ReleaseEvent(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	// Bring escrow to completed state.
	_, err := engine.Fund(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Activate(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.CompleteMilestone(context.Background(), entry.ID, entry.Milestones[0].ID, "done")
	require.NoError(t, err)

	bus.Publish(eventbus.EscrowOnChainReleaseEvent{
		EscrowID: entry.ID,
		DealID:   "1",
		Seller:   "0xSeller",
		Amount:   big.NewInt(1000),
		TxHash:   "0xreleasetx",
	})

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusReleased, updated.Status)
}

func TestBridge_RefundEvent(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	// Bring escrow to disputed state.
	_, err := engine.Fund(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Activate(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Dispute(context.Background(), entry.ID, "test dispute")
	require.NoError(t, err)

	bus.Publish(eventbus.EscrowOnChainRefundEvent{
		EscrowID: entry.ID,
		DealID:   "1",
		Buyer:    "0xBuyer",
		Amount:   big.NewInt(1000),
		TxHash:   "0xrefundtx",
	})

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusRefunded, updated.Status)
}

func TestBridge_DisputeEvent(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	// Bring escrow to active state.
	_, err := engine.Fund(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Activate(context.Background(), entry.ID)
	require.NoError(t, err)

	bus.Publish(eventbus.EscrowOnChainDisputeEvent{
		EscrowID:  entry.ID,
		DealID:    "1",
		Initiator: "0xBuyer",
		TxHash:    "0xdisputetx",
	})

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusDisputed, updated.Status)
}

func TestBridge_ResolvedEvent_SellerFavor(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	// Bring escrow to disputed state.
	_, err := engine.Fund(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Activate(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Dispute(context.Background(), entry.ID, "test")
	require.NoError(t, err)

	bus.Publish(eventbus.EscrowOnChainResolvedEvent{
		EscrowID:    entry.ID,
		DealID:      "1",
		SellerFavor: true,
		TxHash:      "0xresolvedtx",
	})

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusReleased, updated.Status)
}

func TestBridge_ResolvedEvent_BuyerFavor(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	// Bring escrow to disputed state.
	_, err := engine.Fund(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Activate(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Dispute(context.Background(), entry.ID, "test")
	require.NoError(t, err)

	bus.Publish(eventbus.EscrowOnChainResolvedEvent{
		EscrowID:    entry.ID,
		DealID:      "1",
		SellerFavor: false,
		TxHash:      "0xresolvedtx",
	})

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusRefunded, updated.Status)
}

func TestBridge_ReleaseEvent_Idempotent(t *testing.T) {
	t.Parallel()
	bus, engine, _ := bridgeTestSetup(t)

	entry := createTestEscrow(t, engine)

	_, err := engine.Fund(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.Activate(context.Background(), entry.ID)
	require.NoError(t, err)
	_, err = engine.CompleteMilestone(context.Background(), entry.ID, entry.Milestones[0].ID, "done")
	require.NoError(t, err)

	ev := eventbus.EscrowOnChainReleaseEvent{
		EscrowID: entry.ID,
		DealID:   "1",
		Seller:   "0xSeller",
		Amount:   big.NewInt(1000),
		TxHash:   "0xtx",
	}
	bus.Publish(ev)
	bus.Publish(ev) // second time — already released, should not panic

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusReleased, updated.Status)
}

func TestIsAlreadyTransitioned(t *testing.T) {
	t.Parallel()
	assert.False(t, isAlreadyTransitioned(nil))
	assert.True(t, isAlreadyTransitioned(escrow.ErrInvalidTransition))
	assert.False(t, isAlreadyTransitioned(escrow.ErrEscrowNotFound))
}
