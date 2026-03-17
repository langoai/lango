package hub

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/eventbus"
)

func createTestEngine(t *testing.T) (*escrow.Engine, escrow.Store) {
	t.Helper()
	store := escrow.NewMemoryStore()
	cfg := escrow.DefaultEngineConfig()
	engine := escrow.NewEngine(store, escrow.NoopSettler{}, cfg)
	return engine, store
}

func TestDanglingDetector_Name(t *testing.T) {
	t.Parallel()
	engine, store := createTestEngine(t)
	dd := NewDanglingDetector(store, engine, eventbus.New())
	assert.Equal(t, "dangling-detector", dd.Name())
}

func TestDanglingDetector_ScanExpiresDangling(t *testing.T) {
	t.Parallel()
	engine, store := createTestEngine(t)
	bus := eventbus.New()

	var published []eventbus.EscrowDanglingEvent
	eventbus.SubscribeTyped(bus, func(ev eventbus.EscrowDanglingEvent) {
		published = append(published, ev)
	})

	// Create an escrow that is "old" (past maxPending).
	entry, err := engine.Create(context.Background(), escrow.CreateRequest{
		BuyerDID:  "did:test:buyer",
		SellerDID: "did:test:seller",
		Amount:    big.NewInt(500),
		Reason:    "test",
		Milestones: []escrow.MilestoneRequest{
			{Description: "m1", Amount: big.NewInt(500)},
		},
		ExpiresAt: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	})
	require.NoError(t, err)

	// Manually backdate CreatedAt to simulate old escrow.
	e, err := store.Get(entry.ID)
	require.NoError(t, err)
	e.CreatedAt = time.Now().Add(-15 * time.Minute)
	require.NoError(t, store.Update(e))

	dd := NewDanglingDetector(store, engine, bus,
		WithMaxPending(10*time.Minute),
		WithDanglingLogger(zap.NewNop().Sugar()),
	)

	// Run scan directly.
	dd.scan()

	// Escrow should be expired.
	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusExpired, updated.Status)

	// Event should be published.
	require.Len(t, published, 1)
	assert.Equal(t, entry.ID, published[0].EscrowID)
	assert.Equal(t, "expired", published[0].Action)
	assert.Equal(t, "did:test:buyer", published[0].BuyerDID)
}

func TestDanglingDetector_ScanIgnoresYoung(t *testing.T) {
	t.Parallel()
	engine, store := createTestEngine(t)
	bus := eventbus.New()

	// Create a fresh escrow (not old enough).
	entry, err := engine.Create(context.Background(), escrow.CreateRequest{
		BuyerDID:  "did:test:buyer",
		SellerDID: "did:test:seller",
		Amount:    big.NewInt(100),
		Reason:    "test",
		Milestones: []escrow.MilestoneRequest{
			{Description: "m1", Amount: big.NewInt(100)},
		},
		ExpiresAt: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	})
	require.NoError(t, err)

	dd := NewDanglingDetector(store, engine, bus,
		WithMaxPending(10*time.Minute),
	)

	dd.scan()

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusPending, updated.Status)
}

func TestDanglingDetector_ScanIgnoresNonPending(t *testing.T) {
	t.Parallel()
	engine, store := createTestEngine(t)
	bus := eventbus.New()

	entry, err := engine.Create(context.Background(), escrow.CreateRequest{
		BuyerDID:  "did:test:buyer",
		SellerDID: "did:test:seller",
		Amount:    big.NewInt(200),
		Reason:    "test",
		Milestones: []escrow.MilestoneRequest{
			{Description: "m1", Amount: big.NewInt(200)},
		},
		ExpiresAt: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
	})
	require.NoError(t, err)

	// Fund it so it's not Pending anymore.
	_, err = engine.Fund(context.Background(), entry.ID)
	require.NoError(t, err)

	// Backdate CreatedAt.
	e, err := store.Get(entry.ID)
	require.NoError(t, err)
	e.CreatedAt = time.Now().Add(-15 * time.Minute)
	require.NoError(t, store.Update(e))

	dd := NewDanglingDetector(store, engine, bus,
		WithMaxPending(10*time.Minute),
	)
	dd.scan()

	updated, err := engine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, escrow.StatusFunded, updated.Status)
}

func TestDanglingDetector_StartStop(t *testing.T) {
	t.Parallel()
	engine, store := createTestEngine(t)
	bus := eventbus.New()

	dd := NewDanglingDetector(store, engine, bus,
		WithScanInterval(50*time.Millisecond),
		WithDanglingLogger(zap.NewNop().Sugar()),
	)

	var wg sync.WaitGroup
	wg.Add(1)
	err := dd.Start(context.Background(), &wg)
	require.NoError(t, err)
	wg.Wait()

	// Let it run a couple ticks.
	time.Sleep(150 * time.Millisecond)

	err = dd.Stop(context.Background())
	require.NoError(t, err)
}

func TestDanglingDetector_Options(t *testing.T) {
	t.Parallel()
	engine, store := createTestEngine(t)
	bus := eventbus.New()
	logger := zap.NewNop().Sugar()

	dd := NewDanglingDetector(store, engine, bus,
		WithScanInterval(30*time.Second),
		WithMaxPending(20*time.Minute),
		WithDanglingLogger(logger),
	)

	assert.Equal(t, 30*time.Second, dd.scanInterval)
	assert.Equal(t, 20*time.Minute, dd.maxPending)
}
