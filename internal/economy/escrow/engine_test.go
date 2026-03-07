package escrow

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSettler implements SettlementExecutor for tests.
type mockSettler struct {
	lockErr    error
	releaseErr error
	refundErr  error
	locked     []*big.Int
	released   []*big.Int
	refunded   []*big.Int
}

func (m *mockSettler) Lock(_ context.Context, _ string, amount *big.Int) error {
	if m.lockErr != nil {
		return m.lockErr
	}
	m.locked = append(m.locked, new(big.Int).Set(amount))
	return nil
}

func (m *mockSettler) Release(_ context.Context, _ string, amount *big.Int) error {
	if m.releaseErr != nil {
		return m.releaseErr
	}
	m.released = append(m.released, new(big.Int).Set(amount))
	return nil
}

func (m *mockSettler) Refund(_ context.Context, _ string, amount *big.Int) error {
	if m.refundErr != nil {
		return m.refundErr
	}
	m.refunded = append(m.refunded, new(big.Int).Set(amount))
	return nil
}

func newTestEngine(settler *mockSettler) *Engine {
	cfg := DefaultEngineConfig()
	e := NewEngine(NewMemoryStore(), settler, cfg)
	e.nowFunc = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	return e
}

func createFundedEscrow(t *testing.T, e *Engine, settler *mockSettler) *EscrowEntry {
	t.Helper()
	ctx := context.Background()
	entry, err := e.Create(ctx, CreateRequest{
		BuyerDID:  "did:buyer:1",
		SellerDID: "did:seller:1",
		Amount:    big.NewInt(1000),
		Reason:    "test",
		Milestones: []MilestoneRequest{
			{Description: "first half", Amount: big.NewInt(500)},
			{Description: "second half", Amount: big.NewInt(500)},
		},
	})
	require.NoError(t, err)

	entry, err = e.Fund(ctx, entry.ID)
	require.NoError(t, err)
	return entry
}

func TestEngineCreate(t *testing.T) {
	tests := []struct {
		give    string
		req     CreateRequest
		cfg     EngineConfig
		wantErr error
	}{
		{
			give: "success",
			req: CreateRequest{
				BuyerDID:  "did:buyer:1",
				SellerDID: "did:seller:1",
				Amount:    big.NewInt(1000),
				Reason:    "test",
				Milestones: []MilestoneRequest{
					{Description: "task", Amount: big.NewInt(1000)},
				},
			},
			cfg: DefaultEngineConfig(),
		},
		{
			give: "no milestones",
			req: CreateRequest{
				BuyerDID:   "did:buyer:1",
				SellerDID:  "did:seller:1",
				Amount:     big.NewInt(1000),
				Milestones: nil,
			},
			cfg:     DefaultEngineConfig(),
			wantErr: ErrNoMilestones,
		},
		{
			give: "too many milestones",
			req: CreateRequest{
				BuyerDID:  "did:buyer:1",
				SellerDID: "did:seller:1",
				Amount:    big.NewInt(200),
				Milestones: []MilestoneRequest{
					{Description: "a", Amount: big.NewInt(100)},
					{Description: "b", Amount: big.NewInt(100)},
				},
			},
			cfg:     EngineConfig{MaxMilestones: 1},
			wantErr: ErrTooManyMilestones,
		},
		{
			give: "amount mismatch",
			req: CreateRequest{
				BuyerDID:  "did:buyer:1",
				SellerDID: "did:seller:1",
				Amount:    big.NewInt(1000),
				Milestones: []MilestoneRequest{
					{Description: "a", Amount: big.NewInt(500)},
					{Description: "b", Amount: big.NewInt(400)},
				},
			},
			cfg:     DefaultEngineConfig(),
			wantErr: ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			e := NewEngine(NewMemoryStore(), &mockSettler{}, tt.cfg)
			entry, err := e.Create(context.Background(), tt.req)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, entry)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, StatusPending, entry.Status)
			assert.Equal(t, tt.req.BuyerDID, entry.BuyerDID)
			assert.Len(t, entry.Milestones, len(tt.req.Milestones))
		})
	}
}

func TestEngineCreate_CustomExpiry(t *testing.T) {
	e := newTestEngine(&mockSettler{})
	expiry := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	entry, err := e.Create(context.Background(), CreateRequest{
		BuyerDID:  "did:buyer:1",
		SellerDID: "did:seller:1",
		Amount:    big.NewInt(100),
		Milestones: []MilestoneRequest{
			{Description: "task", Amount: big.NewInt(100)},
		},
		ExpiresAt: &expiry,
	})
	require.NoError(t, err)
	assert.Equal(t, expiry, entry.ExpiresAt)
}

func TestEngineFund(t *testing.T) {
	tests := []struct {
		give      string
		setup     func(*Engine) string
		lockErr   error
		wantErr   bool
	}{
		{
			give: "success",
			setup: func(e *Engine) string {
				entry, _ := e.Create(context.Background(), CreateRequest{
					BuyerDID: "did:b", SellerDID: "did:s", Amount: big.NewInt(100),
					Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(100)}},
				})
				return entry.ID
			},
		},
		{
			give: "lock failure",
			setup: func(e *Engine) string {
				entry, _ := e.Create(context.Background(), CreateRequest{
					BuyerDID: "did:b", SellerDID: "did:s", Amount: big.NewInt(100),
					Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(100)}},
				})
				return entry.ID
			},
			lockErr: errors.New("insufficient funds"),
			wantErr: true,
		},
		{
			give: "not found",
			setup: func(e *Engine) string {
				return "nonexistent"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			settler := &mockSettler{lockErr: tt.lockErr}
			e := newTestEngine(settler)
			id := tt.setup(e)

			entry, err := e.Fund(context.Background(), id)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, StatusFunded, entry.Status)
			assert.Len(t, settler.locked, 1)
		})
	}
}

func TestEngineActivate(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	funded := createFundedEscrow(t, e, settler)

	entry, err := e.Activate(context.Background(), funded.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusActive, entry.Status)
}

func TestEngineActivate_InvalidTransition(t *testing.T) {
	e := newTestEngine(&mockSettler{})
	entry, _ := e.Create(context.Background(), CreateRequest{
		BuyerDID: "did:b", SellerDID: "did:s", Amount: big.NewInt(100),
		Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(100)}},
	})

	_, err := e.Activate(context.Background(), entry.ID)
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestEngineCompleteMilestone(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()
	funded := createFundedEscrow(t, e, settler)

	active, err := e.Activate(ctx, funded.ID)
	require.NoError(t, err)

	// Complete first milestone.
	entry, err := e.CompleteMilestone(ctx, active.ID, active.Milestones[0].ID, "done")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, entry.Status)
	assert.Equal(t, 1, entry.CompletedMilestones())

	// Complete second milestone -> status should become completed.
	entry, err = e.CompleteMilestone(ctx, active.ID, active.Milestones[1].ID, "also done")
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, entry.Status)
	assert.True(t, entry.AllMilestonesCompleted())
}

func TestEngineCompleteMilestone_AutoRelease(t *testing.T) {
	settler := &mockSettler{}
	cfg := DefaultEngineConfig()
	cfg.AutoRelease = true
	e := NewEngine(NewMemoryStore(), settler, cfg)
	e.nowFunc = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	ctx := context.Background()

	entry, err := e.Create(ctx, CreateRequest{
		BuyerDID: "did:b", SellerDID: "did:s", Amount: big.NewInt(100),
		Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(100)}},
	})
	require.NoError(t, err)

	entry, err = e.Fund(ctx, entry.ID)
	require.NoError(t, err)
	entry, err = e.Activate(ctx, entry.ID)
	require.NoError(t, err)

	entry, err = e.CompleteMilestone(ctx, entry.ID, entry.Milestones[0].ID, "proof")
	require.NoError(t, err)
	assert.Equal(t, StatusReleased, entry.Status)
	assert.Len(t, settler.released, 1)
}

func TestEngineCompleteMilestone_NotFound(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()
	funded := createFundedEscrow(t, e, settler)
	active, _ := e.Activate(ctx, funded.ID)

	_, err := e.CompleteMilestone(ctx, active.ID, "nonexistent", "proof")
	assert.ErrorIs(t, err, ErrMilestoneNotFound)
}

func TestEngineRelease(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()
	funded := createFundedEscrow(t, e, settler)
	active, _ := e.Activate(ctx, funded.ID)

	// Complete all milestones first.
	for _, m := range active.Milestones {
		active, _ = e.CompleteMilestone(ctx, active.ID, m.ID, "done")
	}
	require.Equal(t, StatusCompleted, active.Status)

	entry, err := e.Release(ctx, active.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusReleased, entry.Status)
	assert.Len(t, settler.released, 1)
}

func TestEngineDispute(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()
	funded := createFundedEscrow(t, e, settler)
	active, _ := e.Activate(ctx, funded.ID)

	entry, err := e.Dispute(ctx, active.ID, "bad delivery")
	require.NoError(t, err)
	assert.Equal(t, StatusDisputed, entry.Status)
	assert.Equal(t, "bad delivery", entry.DisputeNote)
}

func TestEngineRefund(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()
	funded := createFundedEscrow(t, e, settler)
	active, _ := e.Activate(ctx, funded.ID)
	disputed, _ := e.Dispute(ctx, active.ID, "issue")

	entry, err := e.Refund(ctx, disputed.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusRefunded, entry.Status)
	assert.Len(t, settler.refunded, 1)
}

func TestEngineRefund_InvalidTransition(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()
	funded := createFundedEscrow(t, e, settler)

	_, err := e.Refund(ctx, funded.ID)
	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestEngineExpire(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()
	funded := createFundedEscrow(t, e, settler)
	active, _ := e.Activate(ctx, funded.ID)

	entry, err := e.Expire(ctx, active.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusExpired, entry.Status)
	assert.Len(t, settler.refunded, 1)
}

func TestEngineExpire_PendingNoRefund(t *testing.T) {
	e := newTestEngine(&mockSettler{})
	ctx := context.Background()
	entry, _ := e.Create(ctx, CreateRequest{
		BuyerDID: "did:b", SellerDID: "did:s", Amount: big.NewInt(100),
		Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(100)}},
	})

	expired, err := e.Expire(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusExpired, expired.Status)
}

func TestEngineCheckExpiry(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()

	entry, _ := e.Create(ctx, CreateRequest{
		BuyerDID: "did:b", SellerDID: "did:s", Amount: big.NewInt(100),
		Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(100)}},
	})

	// Move time past expiry.
	e.nowFunc = func() time.Time { return time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC) }

	_, err := e.Fund(ctx, entry.ID)
	assert.ErrorIs(t, err, ErrEscrowExpired)
}

func TestEngineListAndGet(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()

	e.Create(ctx, CreateRequest{
		BuyerDID: "did:b1", SellerDID: "did:s1", Amount: big.NewInt(100),
		Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(100)}},
	})
	e.Create(ctx, CreateRequest{
		BuyerDID: "did:b2", SellerDID: "did:s2", Amount: big.NewInt(200),
		Milestones: []MilestoneRequest{{Description: "t", Amount: big.NewInt(200)}},
	})

	assert.Len(t, e.List(), 2)
	assert.Len(t, e.ListByPeer("did:b1"), 1)
	assert.Len(t, e.ListByPeer("did:nobody"), 0)
}

func TestEngineFullLifecycle(t *testing.T) {
	settler := &mockSettler{}
	e := newTestEngine(settler)
	ctx := context.Background()

	// Create
	entry, err := e.Create(ctx, CreateRequest{
		BuyerDID:  "did:buyer:1",
		SellerDID: "did:seller:1",
		Amount:    big.NewInt(1000),
		Reason:    "full lifecycle test",
		Milestones: []MilestoneRequest{
			{Description: "milestone 1", Amount: big.NewInt(600)},
			{Description: "milestone 2", Amount: big.NewInt(400)},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, StatusPending, entry.Status)

	// Fund
	entry, err = e.Fund(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusFunded, entry.Status)

	// Activate
	entry, err = e.Activate(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusActive, entry.Status)

	// Complete milestones
	entry, err = e.CompleteMilestone(ctx, entry.ID, entry.Milestones[0].ID, "delivered part 1")
	require.NoError(t, err)
	assert.Equal(t, StatusActive, entry.Status)

	entry, err = e.CompleteMilestone(ctx, entry.ID, entry.Milestones[1].ID, "delivered part 2")
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, entry.Status)

	// Release
	entry, err = e.Release(ctx, entry.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusReleased, entry.Status)

	// Verify settlement calls.
	assert.Len(t, settler.locked, 1)
	assert.Len(t, settler.released, 1)
	assert.Equal(t, big.NewInt(1000), settler.locked[0])
	assert.Equal(t, big.NewInt(1000), settler.released[0])
}
