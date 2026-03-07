package budget

import (
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func newTestEngine(cfg config.BudgetConfig, opts ...Option) (*Engine, *Store) {
	s := NewStore()
	e, err := NewEngine(s, cfg, opts...)
	if err != nil {
		panic(err)
	}
	return e, s
}

func defaultCfg() config.BudgetConfig {
	return config.BudgetConfig{
		DefaultMax:      "10.00",
		AlertThresholds: []float64{0.5, 0.8, 0.95},
	}
}

func TestEngine_Allocate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give      string
		giveTotal int64
	}{
		{give: "valid allocation", giveTotal: 1000000},
		{give: "small allocation", giveTotal: 1},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			e, _ := newTestEngine(defaultCfg())
			tb, err := e.Allocate("task-1", big.NewInt(tt.giveTotal))
			require.NoError(t, err)
			assert.Equal(t, 0, tb.TotalBudget.Cmp(big.NewInt(tt.giveTotal)))
			assert.Equal(t, StatusActive, tb.Status)
		})
	}
}

func TestEngine_Allocate_DefaultMax(t *testing.T) {
	t.Parallel()

	e, _ := newTestEngine(defaultCfg())
	tb, err := e.Allocate("task-1", nil)
	require.NoError(t, err)

	want := big.NewInt(10_000_000)
	assert.Equal(t, 0, tb.TotalBudget.Cmp(want))
}

func TestEngine_Allocate_NoDefaultNoAmount(t *testing.T) {
	t.Parallel()

	e, _ := newTestEngine(config.BudgetConfig{})
	_, err := e.Allocate("task-1", nil)
	require.ErrorIs(t, err, ErrInvalidAmount)
}

func TestEngine_Allocate_Duplicate(t *testing.T) {
	t.Parallel()

	e, _ := newTestEngine(defaultCfg())
	_, _ = e.Allocate("task-1", big.NewInt(1000000))
	_, err := e.Allocate("task-1", big.NewInt(500000))
	require.ErrorIs(t, err, ErrBudgetExists)
}

func TestEngine_Check(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveAmount int64
		giveStatus BudgetStatus
		giveSpent  int64
		wantErr    error
	}{
		{give: "sufficient budget", giveAmount: 100000, giveStatus: StatusActive},
		{give: "exact remaining", giveAmount: 1000000, giveStatus: StatusActive},
		{give: "exceeds budget", giveAmount: 1000001, giveStatus: StatusActive, wantErr: ErrBudgetExceeded},
		{give: "closed budget", giveAmount: 100, giveStatus: StatusClosed, wantErr: ErrBudgetClosed},
		{give: "exhausted budget", giveAmount: 100, giveStatus: StatusExhausted, wantErr: ErrBudgetExceeded},
		{give: "insufficient after spending", giveAmount: 600000, giveStatus: StatusActive, giveSpent: 500000, wantErr: ErrBudgetExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			e, s := newTestEngine(defaultCfg())
			_, _ = s.Allocate("task-1", big.NewInt(1000000))

			tb, _ := s.Get("task-1")
			tb.Status = tt.giveStatus
			if tt.giveSpent > 0 {
				tb.Spent = big.NewInt(tt.giveSpent)
			}
			_ = s.Update(tb)

			err := e.Check("task-1", big.NewInt(tt.giveAmount))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestEngine_Check_InvalidAmount(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	err := e.Check("task-1", big.NewInt(0))
	require.ErrorIs(t, err, ErrInvalidAmount)
}

func TestEngine_Check_NotFound(t *testing.T) {
	t.Parallel()

	e, _ := newTestEngine(defaultCfg())
	err := e.Check("nonexistent", big.NewInt(100))
	require.ErrorIs(t, err, ErrBudgetNotFound)
}

func TestEngine_Record(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		giveAmount int64
		giveSetup  func(*Store)
		wantErr    error
		wantSpent  int64
	}{
		{give: "valid record", giveAmount: 100000, wantSpent: 100000},
		{give: "exceeds remaining", giveAmount: 1000001, wantErr: ErrBudgetExceeded},
		{give: "zero amount", giveAmount: 0, wantErr: ErrInvalidAmount},
		{
			give:       "closed budget",
			giveAmount: 100,
			giveSetup: func(s *Store) {
				tb, _ := s.Get("task-1")
				tb.Status = StatusClosed
				_ = s.Update(tb)
			},
			wantErr: ErrBudgetClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			e, s := newTestEngine(defaultCfg())
			_, _ = s.Allocate("task-1", big.NewInt(1000000))
			if tt.giveSetup != nil {
				tt.giveSetup(s)
			}

			entry := SpendEntry{
				Amount:   big.NewInt(tt.giveAmount),
				PeerDID:  "did:peer:123",
				ToolName: "compute",
				Reason:   "test",
			}

			err := e.Record("task-1", entry)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			tb, _ := s.Get("task-1")
			assert.Equal(t, 0, tb.Spent.Cmp(big.NewInt(tt.wantSpent)))
			assert.Len(t, tb.Entries, 1)
			assert.NotEmpty(t, tb.Entries[0].ID)
		})
	}
}

func TestEngine_Record_ExhaustsOnFullSpend(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000))

	err := e.Record("task-1", SpendEntry{Amount: big.NewInt(1000), PeerDID: "did:peer:123"})
	require.NoError(t, err)

	tb, _ := s.Get("task-1")
	assert.Equal(t, StatusExhausted, tb.Status)
}

func TestEngine_Record_MultipleEntries(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	for i := range 3 {
		err := e.Record("task-1", SpendEntry{
			Amount:  big.NewInt(100000),
			PeerDID: "did:peer:123",
			Reason:  "entry",
			ID:      "id-" + string(rune('a'+i)),
		})
		require.NoError(t, err)
	}

	tb, _ := s.Get("task-1")
	assert.Equal(t, 0, tb.Spent.Cmp(big.NewInt(300000)))
	assert.Len(t, tb.Entries, 3)
}

func TestEngine_Reserve(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	release, err := e.Reserve("task-1", big.NewInt(500000))
	require.NoError(t, err)

	tb, _ := s.Get("task-1")
	assert.Equal(t, 0, tb.Reserved.Cmp(big.NewInt(500000)))

	release()

	tb, _ = s.Get("task-1")
	assert.Equal(t, 0, tb.Reserved.Sign())
}

func TestEngine_Reserve_ExceedsRemaining(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	_, err := e.Reserve("task-1", big.NewInt(1000001))
	require.ErrorIs(t, err, ErrBudgetExceeded)
}

func TestEngine_Reserve_ReleaseIdempotent(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	release, _ := e.Reserve("task-1", big.NewInt(500000))
	release()
	release()

	tb, _ := s.Get("task-1")
	assert.Equal(t, 0, tb.Reserved.Sign())
}

func TestEngine_SetProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		giveProgress float64
		wantErr      bool
	}{
		{give: "zero", giveProgress: 0.0},
		{give: "half", giveProgress: 0.5},
		{give: "full", giveProgress: 1.0},
		{give: "negative", giveProgress: -0.1, wantErr: true},
		{give: "over 100%", giveProgress: 1.1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			e, s := newTestEngine(defaultCfg())
			_, _ = s.Allocate("task-1", big.NewInt(1000000))

			err := e.SetProgress("task-1", tt.giveProgress)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tb, _ := s.Get("task-1")
			assert.InDelta(t, tt.giveProgress, tb.Progress, 0.001)
		})
	}
}

func TestEngine_Close(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(300000), PeerDID: "did:peer:123"})
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(200000), PeerDID: "did:peer:456"})

	report, err := e.Close("task-1")
	require.NoError(t, err)
	assert.Equal(t, 0, report.TotalSpent.Cmp(big.NewInt(500000)))
	assert.Equal(t, 2, report.EntryCount)
	assert.Equal(t, StatusClosed, report.Status)
}

func TestEngine_Close_AlreadyClosed(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))
	_, _ = e.Close("task-1")
	_, err := e.Close("task-1")
	require.ErrorIs(t, err, ErrBudgetClosed)
}

func TestEngine_BurnRate(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	rate, err := e.BurnRate("task-1")
	require.NoError(t, err)
	assert.Equal(t, 0, rate.Sign())
}

func TestEngine_ThresholdAlerts(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var alerts []float64

	cfg := config.BudgetConfig{
		DefaultMax:      "10.00",
		AlertThresholds: []float64{0.5, 0.8},
	}
	e, s := newTestEngine(cfg, WithAlertCallback(func(_ string, pct float64) {
		mu.Lock()
		defer mu.Unlock()
		alerts = append(alerts, pct)
	}))
	_, _ = s.Allocate("task-1", big.NewInt(1000))

	// Spend 500 -> 50% -> triggers 0.5
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(500), PeerDID: "did:peer:123"})

	mu.Lock()
	require.Len(t, alerts, 1)
	assert.Equal(t, 0.5, alerts[0])
	mu.Unlock()

	// Spend 310 -> 81% -> triggers 0.8
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(310), PeerDID: "did:peer:123"})

	mu.Lock()
	require.Len(t, alerts, 2)
	assert.Equal(t, 0.8, alerts[1])
	mu.Unlock()

	// No re-trigger
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(10), PeerDID: "did:peer:123"})
	mu.Lock()
	assert.Len(t, alerts, 2)
	mu.Unlock()
}

func TestEngine_GuardInterface(t *testing.T) {
	t.Parallel()

	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	var g Guard = e

	require.NoError(t, g.Check("task-1", big.NewInt(100)))
	release, err := g.Reserve("task-1", big.NewInt(200000))
	require.NoError(t, err)
	release()
	require.NoError(t, g.Record("task-1", SpendEntry{Amount: big.NewInt(100), PeerDID: "did:peer:123"}))
}
