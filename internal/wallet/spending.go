package wallet

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/langoai/lango/internal/ent"
	"github.com/langoai/lango/internal/ent/paymenttx"
	"github.com/langoai/lango/internal/finance"
)

// SpendingLimiter enforces per-transaction and daily spending limits.
type SpendingLimiter interface {
	// Check verifies that spending amount is within limits without recording it.
	Check(ctx context.Context, amount *big.Int) error

	// Record records a spent amount for daily tracking.
	Record(ctx context.Context, amount *big.Int) error

	// DailySpent returns the total amount spent today.
	DailySpent(ctx context.Context) (*big.Int, error)

	// DailyRemaining returns the remaining daily budget.
	DailyRemaining(ctx context.Context) (*big.Int, error)

	// IsAutoApprovable checks whether the given amount can be auto-approved
	// without explicit user confirmation, based on the autoApproveBelow threshold
	// and spending limits.
	IsAutoApprovable(ctx context.Context, amount *big.Int) (bool, error)
}

// SpendingUsageStore provides the payment usage data needed for limit checks.
type SpendingUsageStore interface {
	DailySpendSince(ctx context.Context, since time.Time) ([]string, error)
}

// USDCDecimals is the number of decimal places for USDC (6).
// Deprecated: Use finance.USDCDecimals instead.
const USDCDecimals = finance.USDCDecimals

// ParseUSDC converts a decimal string (e.g. "1.50") to the smallest USDC unit.
// Deprecated: Use finance.ParseUSDC instead.
func ParseUSDC(amount string) (*big.Int, error) {
	return finance.ParseUSDC(amount)
}

// FormatUSDC converts smallest USDC units back to a decimal string.
// Deprecated: Use finance.FormatUSDC instead.
func FormatUSDC(amount *big.Int) string {
	return finance.FormatUSDC(amount)
}

// EntSpendingLimiter uses Ent PaymentTx records to enforce spending limits.
type EntSpendingLimiter struct {
	store            SpendingUsageStore
	maxPerTx         *big.Int
	maxDaily         *big.Int
	autoApproveBelow *big.Int
}

// NewEntSpendingLimiter creates a spending limiter backed by Ent PaymentTx records.
// autoApproveBelow is the USDC amount threshold below which transactions can be
// auto-approved without explicit user confirmation. Pass "" or "0" to disable.
func NewEntSpendingLimiter(client *ent.Client, maxPerTx, maxDaily, autoApproveBelow string) (*EntSpendingLimiter, error) {
	return NewStoreSpendingLimiter(entUsageStore{client: client}, maxPerTx, maxDaily, autoApproveBelow)
}

// NewStoreSpendingLimiter creates a spending limiter backed by a payment tx store.
func NewStoreSpendingLimiter(store SpendingUsageStore, maxPerTx, maxDaily, autoApproveBelow string) (*EntSpendingLimiter, error) {
	perTx, err := ParseUSDC(maxPerTx)
	if err != nil {
		return nil, fmt.Errorf("parse maxPerTx: %w", err)
	}

	daily, err := ParseUSDC(maxDaily)
	if err != nil {
		return nil, fmt.Errorf("parse maxDaily: %w", err)
	}

	autoApprove := big.NewInt(0)
	if autoApproveBelow != "" {
		parsed, err := ParseUSDC(autoApproveBelow)
		if err != nil {
			return nil, fmt.Errorf("parse autoApproveBelow: %w", err)
		}
		autoApprove = parsed
	}

	return &EntSpendingLimiter{
		store:            store,
		maxPerTx:         perTx,
		maxDaily:         daily,
		autoApproveBelow: autoApprove,
	}, nil
}

// Check verifies that the amount does not exceed per-tx or daily limits.
func (l *EntSpendingLimiter) Check(ctx context.Context, amount *big.Int) error {
	if amount.Cmp(l.maxPerTx) > 0 {
		return fmt.Errorf("amount %s exceeds per-transaction limit %s",
			FormatUSDC(amount), FormatUSDC(l.maxPerTx))
	}

	spent, err := l.DailySpent(ctx)
	if err != nil {
		return fmt.Errorf("check daily spent: %w", err)
	}

	projected := new(big.Int).Add(spent, amount)
	if projected.Cmp(l.maxDaily) > 0 {
		return fmt.Errorf("amount %s would exceed daily limit %s (already spent %s today)",
			FormatUSDC(amount), FormatUSDC(l.maxDaily), FormatUSDC(spent))
	}

	return nil
}

// Record is a no-op: spending is tracked via PaymentTx records created by PaymentService.
func (l *EntSpendingLimiter) Record(_ context.Context, _ *big.Int) error {
	return nil
}

// DailySpent sums confirmed and submitted transaction amounts for today.
func (l *EntSpendingLimiter) DailySpent(ctx context.Context) (*big.Int, error) {
	startOfDay := startOfToday()
	amounts, err := l.store.DailySpendSince(ctx, startOfDay)
	if err != nil {
		return nil, fmt.Errorf("query daily transactions: %w", err)
	}

	total := new(big.Int)
	for _, amount := range amounts {
		amt, err := ParseUSDC(amount)
		if err != nil {
			continue
		}
		total.Add(total, amt)
	}

	return total, nil
}

// DailyRemaining returns how much can still be spent today.
func (l *EntSpendingLimiter) DailyRemaining(ctx context.Context) (*big.Int, error) {
	spent, err := l.DailySpent(ctx)
	if err != nil {
		return nil, err
	}

	remaining := new(big.Int).Sub(l.maxDaily, spent)
	if remaining.Sign() < 0 {
		return big.NewInt(0), nil
	}

	return remaining, nil
}

// MaxPerTx returns the per-transaction limit.
func (l *EntSpendingLimiter) MaxPerTx() *big.Int {
	return new(big.Int).Set(l.maxPerTx)
}

// MaxDaily returns the daily spending limit.
func (l *EntSpendingLimiter) MaxDaily() *big.Int {
	return new(big.Int).Set(l.maxDaily)
}

// IsAutoApprovable checks whether amount can be auto-approved without user confirmation.
// Returns false when auto-approve is disabled (threshold is 0), when amount exceeds the
// threshold, or when spending limits would be exceeded.
func (l *EntSpendingLimiter) IsAutoApprovable(ctx context.Context, amount *big.Int) (bool, error) {
	if l.autoApproveBelow.Sign() == 0 {
		return false, nil
	}

	if amount.Cmp(l.autoApproveBelow) > 0 {
		return false, nil
	}

	if err := l.Check(ctx, amount); err != nil {
		return false, err
	}

	return true, nil
}

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

var _ SpendingLimiter = (*EntSpendingLimiter)(nil)

type entUsageStore struct {
	client *ent.Client
}

func (s entUsageStore) DailySpendSince(ctx context.Context, since time.Time) ([]string, error) {
	if s.client == nil {
		return nil, fmt.Errorf("payment usage store unavailable")
	}
	rows, err := s.client.PaymentTx.Query().
		Where(
			paymenttx.CreatedAtGTE(since),
			paymenttx.StatusIn(paymenttx.StatusPending, paymenttx.StatusSubmitted, paymenttx.StatusConfirmed),
		).
		Select(paymenttx.FieldAmount).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Amount)
	}
	return out, nil
}
