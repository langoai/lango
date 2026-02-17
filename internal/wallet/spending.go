package wallet

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/langowarny/lango/internal/ent"
	"github.com/langowarny/lango/internal/ent/paymenttx"
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
}

// USDCDecimals is the number of decimal places for USDC (6).
const USDCDecimals = 6

// ParseUSDC converts a decimal string (e.g. "1.50") to the smallest USDC unit.
func ParseUSDC(amount string) (*big.Int, error) {
	// Parse as a rational number to handle decimals.
	rat := new(big.Rat)
	if _, ok := rat.SetString(amount); !ok {
		return nil, fmt.Errorf("invalid USDC amount: %q", amount)
	}

	// Multiply by 10^6 to get smallest unit.
	multiplier := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(USDCDecimals), nil))
	rat.Mul(rat, multiplier)

	if !rat.IsInt() {
		return nil, fmt.Errorf("USDC amount %q has too many decimal places", amount)
	}

	return rat.Num(), nil
}

// FormatUSDC converts smallest USDC units back to a decimal string.
func FormatUSDC(amount *big.Int) string {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(USDCDecimals), nil)
	whole := new(big.Int).Div(amount, divisor)
	remainder := new(big.Int).Mod(amount, divisor)

	if remainder.Sign() == 0 {
		return whole.String() + ".00"
	}

	return fmt.Sprintf("%s.%06s", whole.String(), remainder.String())
}

// EntSpendingLimiter uses Ent PaymentTx records to enforce spending limits.
type EntSpendingLimiter struct {
	client   *ent.Client
	maxPerTx *big.Int
	maxDaily *big.Int
}

// NewEntSpendingLimiter creates a spending limiter backed by Ent PaymentTx records.
func NewEntSpendingLimiter(client *ent.Client, maxPerTx, maxDaily string) (*EntSpendingLimiter, error) {
	perTx, err := ParseUSDC(maxPerTx)
	if err != nil {
		return nil, fmt.Errorf("parse maxPerTx: %w", err)
	}

	daily, err := ParseUSDC(maxDaily)
	if err != nil {
		return nil, fmt.Errorf("parse maxDaily: %w", err)
	}

	return &EntSpendingLimiter{
		client:   client,
		maxPerTx: perTx,
		maxDaily: daily,
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

	// Query all non-failed transactions from today.
	txs, err := l.client.PaymentTx.Query().
		Where(
			paymenttx.CreatedAtGTE(startOfDay),
			paymenttx.StatusIn(paymenttx.StatusPending, paymenttx.StatusSubmitted, paymenttx.StatusConfirmed),
		).
		Select(paymenttx.FieldAmount).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query daily transactions: %w", err)
	}

	total := new(big.Int)
	for _, tx := range txs {
		amt, err := ParseUSDC(tx.Amount)
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

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

var _ SpendingLimiter = (*EntSpendingLimiter)(nil)
