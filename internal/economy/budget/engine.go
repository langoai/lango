package budget

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/config"
)

var (
	ErrBudgetExceeded = errors.New("budget exceeded")
	ErrBudgetClosed   = errors.New("budget is closed")
	ErrInvalidAmount  = errors.New("invalid amount")
)

// RiskAssessor is a local interface to avoid importing the risk package directly.
type RiskAssessor func(ctx context.Context, peerDID string, amount *big.Int) error

// Engine implements the Guard interface with budget management logic.
type Engine struct {
	store         *Store
	cfg           config.BudgetConfig
	alertCallback func(taskID string, pct float64)
	riskAssessor  RiskAssessor
	defaultMax    *big.Int
	thresholds    []float64
	mu            sync.Mutex
	alertsSent    map[string]map[float64]bool
}

var _ Guard = (*Engine)(nil)

// NewEngine creates a new budget engine from config and options.
func NewEngine(store *Store, cfg config.BudgetConfig, opts ...Option) (*Engine, error) {
	e := &Engine{
		store:      store,
		cfg:        cfg,
		alertsSent: make(map[string]map[float64]bool),
	}

	if cfg.DefaultMax != "" {
		dm, ok := parseUSDC(cfg.DefaultMax)
		if !ok {
			return nil, fmt.Errorf("parse defaultMax %q: %w", cfg.DefaultMax, ErrInvalidAmount)
		}
		e.defaultMax = dm
	}

	if len(cfg.AlertThresholds) > 0 {
		e.thresholds = make([]float64, len(cfg.AlertThresholds))
		copy(e.thresholds, cfg.AlertThresholds)
		sort.Float64s(e.thresholds)
	}

	for _, opt := range opts {
		opt(e)
	}

	return e, nil
}

// Allocate creates a new task budget.
// If totalBudget is nil or zero, the default max from config is used.
func (e *Engine) Allocate(taskID string, totalBudget *big.Int) (*TaskBudget, error) {
	total := totalBudget
	if total == nil || total.Sign() <= 0 {
		if e.defaultMax == nil {
			return nil, fmt.Errorf("allocate %q: no budget specified and no default configured: %w",
				taskID, ErrInvalidAmount)
		}
		total = new(big.Int).Set(e.defaultMax)
	}
	return e.store.Allocate(taskID, total)
}

// Check verifies amount is within budget. If HardLimit is enabled (default),
// the check rejects amounts exceeding the remaining budget.
func (e *Engine) Check(taskID string, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return fmt.Errorf("check %q: %w", taskID, ErrInvalidAmount)
	}

	tb, err := e.store.Get(taskID)
	if err != nil {
		return err
	}

	if tb.Status == StatusClosed {
		return fmt.Errorf("check %q: %w", taskID, ErrBudgetClosed)
	}
	if tb.Status == StatusExhausted {
		return fmt.Errorf("check %q: %w", taskID, ErrBudgetExceeded)
	}

	if e.isHardLimit() {
		remaining := tb.Remaining()
		if amount.Cmp(remaining) > 0 {
			return fmt.Errorf("check %q: need %s but %s remaining: %w",
				taskID, amount, remaining, ErrBudgetExceeded)
		}
	}

	return nil
}

// Record records a spend entry, updates the budget, and checks threshold alerts.
func (e *Engine) Record(taskID string, entry SpendEntry) error {
	if entry.Amount == nil || entry.Amount.Sign() <= 0 {
		return fmt.Errorf("record %q: %w", taskID, ErrInvalidAmount)
	}

	tb, err := e.store.Get(taskID)
	if err != nil {
		return err
	}

	if tb.Status == StatusClosed {
		return fmt.Errorf("record %q: %w", taskID, ErrBudgetClosed)
	}

	if e.isHardLimit() {
		remaining := tb.Remaining()
		if entry.Amount.Cmp(remaining) > 0 {
			return fmt.Errorf("record %q: need %s but %s remaining: %w",
				taskID, entry.Amount, remaining, ErrBudgetExceeded)
		}
	}

	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	tb.Spent.Add(tb.Spent, entry.Amount)
	tb.Entries = append(tb.Entries, entry)

	if tb.Remaining().Sign() <= 0 {
		tb.Status = StatusExhausted
	}

	if err := e.store.Update(tb); err != nil {
		return err
	}

	e.checkThresholds(tb)
	return nil
}

// Reserve temporarily reserves an amount from the budget.
// Returns a release function that must be called to return the reserved funds.
func (e *Engine) Reserve(taskID string, amount *big.Int) (func(), error) {
	if amount.Sign() <= 0 {
		return nil, fmt.Errorf("reserve %q: %w", taskID, ErrInvalidAmount)
	}

	tb, err := e.store.Get(taskID)
	if err != nil {
		return nil, err
	}

	if tb.Status == StatusClosed {
		return nil, fmt.Errorf("reserve %q: %w", taskID, ErrBudgetClosed)
	}

	remaining := tb.Remaining()
	if amount.Cmp(remaining) > 0 {
		return nil, fmt.Errorf("reserve %q: need %s but %s remaining: %w",
			taskID, amount, remaining, ErrBudgetExceeded)
	}

	tb.Reserved.Add(tb.Reserved, amount)
	if err := e.store.Update(tb); err != nil {
		return nil, err
	}

	released := false
	releaseFunc := func() {
		if released {
			return
		}
		released = true
		if tb, err := e.store.Get(taskID); err == nil {
			tb.Reserved.Sub(tb.Reserved, amount)
			_ = e.store.Update(tb)
		}
	}

	return releaseFunc, nil
}

// SetProgress updates task completion progress (0.0 to 1.0).
func (e *Engine) SetProgress(taskID string, progress float64) error {
	if progress < 0 || progress > 1 {
		return fmt.Errorf("set progress %q: progress must be between 0.0 and 1.0", taskID)
	}

	tb, err := e.store.Get(taskID)
	if err != nil {
		return err
	}

	tb.Progress = progress
	return e.store.Update(tb)
}

// Close finalizes a budget and returns a report.
func (e *Engine) Close(taskID string) (*BudgetReport, error) {
	tb, err := e.store.Get(taskID)
	if err != nil {
		return nil, err
	}

	if tb.Status == StatusClosed {
		return nil, fmt.Errorf("close %q: %w", taskID, ErrBudgetClosed)
	}

	tb.Status = StatusClosed
	if err := e.store.Update(tb); err != nil {
		return nil, err
	}

	return &BudgetReport{
		TaskID:      tb.TaskID,
		TotalBudget: new(big.Int).Set(tb.TotalBudget),
		TotalSpent:  new(big.Int).Set(tb.Spent),
		EntryCount:  len(tb.Entries),
		Duration:    time.Since(tb.CreatedAt),
		Status:      StatusClosed,
	}, nil
}

// BurnRate returns the spending rate per minute for a task.
// Returns zero if no time has elapsed or nothing has been spent.
func (e *Engine) BurnRate(taskID string) (*big.Int, error) {
	tb, err := e.store.Get(taskID)
	if err != nil {
		return nil, err
	}

	if tb.Spent.Sign() == 0 || len(tb.Entries) == 0 {
		return new(big.Int), nil
	}

	elapsed := time.Since(tb.CreatedAt).Minutes()
	if elapsed <= 0 {
		return new(big.Int), nil
	}

	rate := new(big.Float).SetInt(tb.Spent)
	rate.Quo(rate, new(big.Float).SetFloat64(elapsed))

	result, _ := rate.Int(nil)
	return result, nil
}

// isHardLimit returns true if the hard limit is enabled (default: true).
func (e *Engine) isHardLimit() bool {
	return e.cfg.HardLimit == nil || *e.cfg.HardLimit
}

// checkThresholds fires alert callbacks when spent/total crosses configured thresholds.
func (e *Engine) checkThresholds(tb *TaskBudget) {
	if e.alertCallback == nil || tb.TotalBudget.Sign() == 0 {
		return
	}

	spent := new(big.Float).SetInt(tb.Spent)
	total := new(big.Float).SetInt(tb.TotalBudget)
	pct, _ := new(big.Float).Quo(spent, total).Float64()

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.alertsSent[tb.TaskID]; !ok {
		e.alertsSent[tb.TaskID] = make(map[float64]bool)
	}

	for _, threshold := range e.thresholds {
		if pct >= threshold && !e.alertsSent[tb.TaskID][threshold] {
			e.alertsSent[tb.TaskID][threshold] = true
			e.alertCallback(tb.TaskID, threshold)
		}
	}
}

// parseUSDC parses a decimal USDC string (e.g. "10.00") into the smallest unit (6 decimals).
func parseUSDC(s string) (*big.Int, bool) {
	f, _, err := new(big.Float).Parse(s, 10)
	if err != nil {
		return nil, false
	}
	multiplier := new(big.Float).SetInt64(1_000_000)
	f.Mul(f, multiplier)
	result, _ := f.Int(nil)
	if result.Sign() <= 0 {
		return nil, false
	}
	return result, true
}
