package budget

import (
	"errors"
	"math/big"
	"sync"
	"testing"

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
	tests := []struct {
		give      string
		giveTotal int64
	}{
		{give: "valid allocation", giveTotal: 1000000},
		{give: "small allocation", giveTotal: 1},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			e, _ := newTestEngine(defaultCfg())
			tb, err := e.Allocate("task-1", big.NewInt(tt.giveTotal))
			if err != nil {
				t.Fatalf("Allocate() unexpected error: %v", err)
			}
			if tb.TotalBudget.Cmp(big.NewInt(tt.giveTotal)) != 0 {
				t.Errorf("TotalBudget = %s, want %d", tb.TotalBudget, tt.giveTotal)
			}
			if tb.Status != StatusActive {
				t.Errorf("Status = %q, want %q", tb.Status, StatusActive)
			}
		})
	}
}

func TestEngine_Allocate_DefaultMax(t *testing.T) {
	e, _ := newTestEngine(defaultCfg())
	tb, err := e.Allocate("task-1", nil)
	if err != nil {
		t.Fatalf("Allocate() with default max: %v", err)
	}
	want := big.NewInt(10_000_000)
	if tb.TotalBudget.Cmp(want) != 0 {
		t.Errorf("TotalBudget = %s, want %s", tb.TotalBudget, want)
	}
}

func TestEngine_Allocate_NoDefaultNoAmount(t *testing.T) {
	e, _ := newTestEngine(config.BudgetConfig{})
	_, err := e.Allocate("task-1", nil)
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("Allocate() nil with no default: got %v, want ErrInvalidAmount", err)
	}
}

func TestEngine_Allocate_Duplicate(t *testing.T) {
	e, _ := newTestEngine(defaultCfg())
	_, _ = e.Allocate("task-1", big.NewInt(1000000))
	_, err := e.Allocate("task-1", big.NewInt(500000))
	if !errors.Is(err, ErrBudgetExists) {
		t.Fatalf("duplicate Allocate() error = %v, want ErrBudgetExists", err)
	}
}

func TestEngine_Check(t *testing.T) {
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
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Check() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Check() unexpected error: %v", err)
			}
		})
	}
}

func TestEngine_Check_InvalidAmount(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	err := e.Check("task-1", big.NewInt(0))
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("Check(0) error = %v, want ErrInvalidAmount", err)
	}
}

func TestEngine_Check_NotFound(t *testing.T) {
	e, _ := newTestEngine(defaultCfg())
	err := e.Check("nonexistent", big.NewInt(100))
	if !errors.Is(err, ErrBudgetNotFound) {
		t.Fatalf("Check() error = %v, want ErrBudgetNotFound", err)
	}
}

func TestEngine_Record(t *testing.T) {
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
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Record() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Record() unexpected error: %v", err)
			}

			tb, _ := s.Get("task-1")
			if tb.Spent.Cmp(big.NewInt(tt.wantSpent)) != 0 {
				t.Errorf("Spent = %s, want %d", tb.Spent, tt.wantSpent)
			}
			if len(tb.Entries) != 1 {
				t.Errorf("Entries count = %d, want 1", len(tb.Entries))
			}
			if tb.Entries[0].ID == "" {
				t.Error("Entry ID should be auto-generated")
			}
		})
	}
}

func TestEngine_Record_ExhaustsOnFullSpend(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000))

	err := e.Record("task-1", SpendEntry{Amount: big.NewInt(1000), PeerDID: "did:peer:123"})
	if err != nil {
		t.Fatalf("Record() unexpected error: %v", err)
	}

	tb, _ := s.Get("task-1")
	if tb.Status != StatusExhausted {
		t.Errorf("Status = %q, want %q", tb.Status, StatusExhausted)
	}
}

func TestEngine_Record_MultipleEntries(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	for i := range 3 {
		err := e.Record("task-1", SpendEntry{
			Amount:  big.NewInt(100000),
			PeerDID: "did:peer:123",
			Reason:  "entry",
			ID:      "id-" + string(rune('a'+i)),
		})
		if err != nil {
			t.Fatalf("Record() entry %d: unexpected error: %v", i, err)
		}
	}

	tb, _ := s.Get("task-1")
	if tb.Spent.Cmp(big.NewInt(300000)) != 0 {
		t.Errorf("Spent = %s, want 300000", tb.Spent)
	}
	if len(tb.Entries) != 3 {
		t.Errorf("Entries count = %d, want 3", len(tb.Entries))
	}
}

func TestEngine_Reserve(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	release, err := e.Reserve("task-1", big.NewInt(500000))
	if err != nil {
		t.Fatalf("Reserve() unexpected error: %v", err)
	}

	tb, _ := s.Get("task-1")
	if tb.Reserved.Cmp(big.NewInt(500000)) != 0 {
		t.Errorf("Reserved = %s, want 500000", tb.Reserved)
	}

	release()

	tb, _ = s.Get("task-1")
	if tb.Reserved.Sign() != 0 {
		t.Errorf("Reserved after release = %s, want 0", tb.Reserved)
	}
}

func TestEngine_Reserve_ExceedsRemaining(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	_, err := e.Reserve("task-1", big.NewInt(1000001))
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("Reserve() error = %v, want ErrBudgetExceeded", err)
	}
}

func TestEngine_Reserve_ReleaseIdempotent(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	release, _ := e.Reserve("task-1", big.NewInt(500000))
	release()
	release()

	tb, _ := s.Get("task-1")
	if tb.Reserved.Sign() != 0 {
		t.Errorf("Reserved after double release = %s, want 0", tb.Reserved)
	}
}

func TestEngine_SetProgress(t *testing.T) {
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
			e, s := newTestEngine(defaultCfg())
			_, _ = s.Allocate("task-1", big.NewInt(1000000))

			err := e.SetProgress("task-1", tt.giveProgress)
			if tt.wantErr {
				if err == nil {
					t.Fatal("SetProgress() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("SetProgress() unexpected error: %v", err)
			}
			tb, _ := s.Get("task-1")
			if tb.Progress != tt.giveProgress {
				t.Errorf("Progress = %f, want %f", tb.Progress, tt.giveProgress)
			}
		})
	}
}

func TestEngine_Close(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(300000), PeerDID: "did:peer:123"})
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(200000), PeerDID: "did:peer:456"})

	report, err := e.Close("task-1")
	if err != nil {
		t.Fatalf("Close() unexpected error: %v", err)
	}
	if report.TotalSpent.Cmp(big.NewInt(500000)) != 0 {
		t.Errorf("TotalSpent = %s, want 500000", report.TotalSpent)
	}
	if report.EntryCount != 2 {
		t.Errorf("EntryCount = %d, want 2", report.EntryCount)
	}
	if report.Status != StatusClosed {
		t.Errorf("Status = %q, want %q", report.Status, StatusClosed)
	}
}

func TestEngine_Close_AlreadyClosed(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))
	_, _ = e.Close("task-1")
	_, err := e.Close("task-1")
	if !errors.Is(err, ErrBudgetClosed) {
		t.Fatalf("second Close() error = %v, want ErrBudgetClosed", err)
	}
}

func TestEngine_BurnRate(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	rate, err := e.BurnRate("task-1")
	if err != nil {
		t.Fatalf("BurnRate() unexpected error: %v", err)
	}
	if rate.Sign() != 0 {
		t.Errorf("BurnRate() with no spending = %s, want 0", rate)
	}
}

func TestEngine_ThresholdAlerts(t *testing.T) {
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

	// Spend 500 → 50% → triggers 0.5
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(500), PeerDID: "did:peer:123"})

	mu.Lock()
	if len(alerts) != 1 || alerts[0] != 0.5 {
		t.Fatalf("after 50%% spend: alerts = %v, want [0.5]", alerts)
	}
	mu.Unlock()

	// Spend 310 → 81% → triggers 0.8
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(310), PeerDID: "did:peer:123"})

	mu.Lock()
	if len(alerts) != 2 || alerts[1] != 0.8 {
		t.Fatalf("after 81%% spend: alerts = %v, want [0.5, 0.8]", alerts)
	}
	mu.Unlock()

	// No re-trigger
	_ = e.Record("task-1", SpendEntry{Amount: big.NewInt(10), PeerDID: "did:peer:123"})
	mu.Lock()
	if len(alerts) != 2 {
		t.Errorf("expected no re-trigger, got %d alerts", len(alerts))
	}
	mu.Unlock()
}

func TestEngine_GuardInterface(t *testing.T) {
	e, s := newTestEngine(defaultCfg())
	_, _ = s.Allocate("task-1", big.NewInt(1000000))

	var g Guard = e

	if err := g.Check("task-1", big.NewInt(100)); err != nil {
		t.Fatalf("Guard.Check() unexpected error: %v", err)
	}
	release, err := g.Reserve("task-1", big.NewInt(200000))
	if err != nil {
		t.Fatalf("Guard.Reserve() unexpected error: %v", err)
	}
	release()
	if err := g.Record("task-1", SpendEntry{Amount: big.NewInt(100), PeerDID: "did:peer:123"}); err != nil {
		t.Fatalf("Guard.Record() unexpected error: %v", err)
	}
}
