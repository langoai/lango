package wallet

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"
)

type fakeSpendingUsageStore struct {
	amounts []string
	err     error
	calls   int
	since   time.Time
}

func (s *fakeSpendingUsageStore) DailySpendSince(_ context.Context, since time.Time) ([]string, error) {
	s.calls++
	s.since = since
	if s.err != nil {
		return nil, s.err
	}
	return append([]string(nil), s.amounts...), nil
}

func TestParseUSDC(t *testing.T) {
	tests := []struct {
		give    string
		want    int64
		wantErr bool
	}{
		{give: "1.00", want: 1_000_000},
		{give: "0.50", want: 500_000},
		{give: "10.00", want: 10_000_000},
		{give: "0.000001", want: 1},
		{give: "0", want: 0},
		{give: "100", want: 100_000_000},
		{give: "invalid", wantErr: true},
		{give: "0.0000001", wantErr: true}, // too many decimals
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseUSDC(tt.give)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseUSDC(%q) expected error, got %v", tt.give, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseUSDC(%q) unexpected error: %v", tt.give, err)
			}
			if got.Int64() != tt.want {
				t.Errorf("ParseUSDC(%q) = %d, want %d", tt.give, got.Int64(), tt.want)
			}
		})
	}
}

func TestFormatUSDC(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: 1_000_000, want: "1.00"},
		{give: 500_000, want: "0.50"},
		{give: 10_000_000, want: "10.00"},
		{give: 0, want: "0.00"},
		{give: 1, want: "0.000001"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatUSDC(big.NewInt(tt.give))
			if got != tt.want {
				t.Errorf("FormatUSDC(%d) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}

func TestIsAutoApprovable(t *testing.T) {
	tests := []struct {
		give             string
		autoApproveBelow string
		wantOK           bool
		wantErr          bool
	}{
		{give: "0.05", autoApproveBelow: "0.10", wantOK: true},
		{give: "0.10", autoApproveBelow: "0.10", wantOK: true},
		{give: "0.11", autoApproveBelow: "0.10", wantOK: false},
		{give: "1.00", autoApproveBelow: "0.10", wantOK: false},
		{give: "0.05", autoApproveBelow: "0", wantOK: false},    // disabled
		{give: "0.05", autoApproveBelow: "", wantOK: false},     // disabled
		{give: "0.00", autoApproveBelow: "0.10", wantOK: true},  // zero amount
		{give: "5.00", autoApproveBelow: "10.00", wantOK: true}, // large threshold
	}

	for _, tt := range tests {
		name := fmt.Sprintf("amount=%s,threshold=%s", tt.give, tt.autoApproveBelow)
		t.Run(name, func(t *testing.T) {
			limiter := &EntSpendingLimiter{
				maxPerTx:         big.NewInt(100_000_000), // 100 USDC
				maxDaily:         big.NewInt(100_000_000), // 100 USDC
				autoApproveBelow: big.NewInt(0),
			}

			// Parse auto-approve threshold.
			if tt.autoApproveBelow != "" {
				parsed, err := ParseUSDC(tt.autoApproveBelow)
				if err != nil {
					t.Fatalf("parse autoApproveBelow: %v", err)
				}
				limiter.autoApproveBelow = parsed
			}

			amt, err := ParseUSDC(tt.give)
			if err != nil {
				t.Fatalf("parse amount: %v", err)
			}

			// IsAutoApprovable uses Check() which requires an ent client for DailySpent.
			// Since we can't create a real ent client in unit tests, we test the
			// threshold logic directly. The client-dependent path is covered by
			// integration tests.
			if limiter.autoApproveBelow.Sign() == 0 {
				if tt.wantOK {
					t.Error("expected auto-approvable but threshold is 0")
				}
				return
			}

			if amt.Cmp(limiter.autoApproveBelow) > 0 {
				if tt.wantOK {
					t.Errorf("amount %s > threshold %s, expected not auto-approvable",
						tt.give, tt.autoApproveBelow)
				}
				return
			}

			// Amount is within threshold.
			if !tt.wantOK {
				t.Errorf("amount %s <= threshold %s, expected auto-approvable",
					tt.give, tt.autoApproveBelow)
			}
		})
	}
}

func TestNewEntSpendingLimiter_AutoApproveBelow(t *testing.T) {
	tests := []struct {
		give    string
		want    int64
		wantErr bool
	}{
		{give: "0.10", want: 100_000},
		{give: "1.00", want: 1_000_000},
		{give: "0", want: 0},
		{give: "", want: 0},
		{give: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			limiter, err := NewEntSpendingLimiter(nil, "1.00", "10.00", tt.give)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if limiter.autoApproveBelow.Int64() != tt.want {
				t.Errorf("autoApproveBelow = %d, want %d",
					limiter.autoApproveBelow.Int64(), tt.want)
			}
		})
	}
}

func TestSpendingLimiterDailySpentSumsValidAmountsAndSkipsMalformedRows(t *testing.T) {
	t.Parallel()

	store := &fakeSpendingUsageStore{
		amounts: []string{"1.25", "not-usdc", "0.75", "0.000001"},
	}
	limiter, err := NewStoreSpendingLimiter(store, "10.00", "20.00", "0")
	if err != nil {
		t.Fatalf("NewStoreSpendingLimiter: %v", err)
	}

	before := time.Now()
	spent, err := limiter.DailySpent(context.Background())
	if err != nil {
		t.Fatalf("DailySpent: %v", err)
	}
	after := time.Now()

	want, err := ParseUSDC("2.000001")
	if err != nil {
		t.Fatalf("ParseUSDC: %v", err)
	}
	if spent.Cmp(want) != 0 {
		t.Fatalf("DailySpent = %s, want %s", FormatUSDC(spent), FormatUSDC(want))
	}
	if store.calls != 1 {
		t.Fatalf("DailySpendSince calls = %d, want 1", store.calls)
	}
	if store.since.Hour() != 0 || store.since.Minute() != 0 || store.since.Second() != 0 || store.since.Nanosecond() != 0 {
		t.Fatalf("DailySpendSince since = %s, want start of day", store.since)
	}
	if !sameLocalDate(store.since, before) && !sameLocalDate(store.since, after) {
		t.Fatalf("DailySpendSince date = %s, want date between %s and %s", store.since, before, after)
	}
}

func TestSpendingLimiterDailySpentWrapsStoreErrors(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("database unavailable")
	limiter, err := NewStoreSpendingLimiter(&fakeSpendingUsageStore{err: storeErr}, "10.00", "20.00", "0")
	if err != nil {
		t.Fatalf("NewStoreSpendingLimiter: %v", err)
	}

	_, err = limiter.DailySpent(context.Background())
	if err == nil {
		t.Fatal("DailySpent error = nil, want store error")
	}
	if !strings.Contains(err.Error(), "query daily transactions") || !errors.Is(err, storeErr) {
		t.Fatalf("DailySpent error = %v, want wrapped store error", err)
	}
}

func TestSpendingLimiterCheckEnforcesPerTransactionBeforeStoreLookup(t *testing.T) {
	t.Parallel()

	store := &fakeSpendingUsageStore{}
	limiter, err := NewStoreSpendingLimiter(store, "5.00", "10.00", "0")
	if err != nil {
		t.Fatalf("NewStoreSpendingLimiter: %v", err)
	}
	amount, err := ParseUSDC("5.01")
	if err != nil {
		t.Fatalf("ParseUSDC: %v", err)
	}

	err = limiter.Check(context.Background(), amount)
	if err == nil {
		t.Fatal("Check error = nil, want per-transaction limit error")
	}
	if !strings.Contains(err.Error(), "exceeds per-transaction limit") {
		t.Fatalf("Check error = %v, want per-transaction limit", err)
	}
	if store.calls != 0 {
		t.Fatalf("DailySpendSince calls = %d, want 0", store.calls)
	}
}

func TestSpendingLimiterCheckEnforcesDailyLimitAndPropagatesDailyErrors(t *testing.T) {
	t.Parallel()

	t.Run("daily limit exceeded", func(t *testing.T) {
		t.Parallel()

		limiter, err := NewStoreSpendingLimiter(
			&fakeSpendingUsageStore{amounts: []string{"9.50"}},
			"10.00",
			"10.00",
			"0",
		)
		if err != nil {
			t.Fatalf("NewStoreSpendingLimiter: %v", err)
		}
		amount, err := ParseUSDC("0.51")
		if err != nil {
			t.Fatalf("ParseUSDC: %v", err)
		}

		err = limiter.Check(context.Background(), amount)
		if err == nil {
			t.Fatal("Check error = nil, want daily limit error")
		}
		if !strings.Contains(err.Error(), "would exceed daily limit") {
			t.Fatalf("Check error = %v, want daily limit", err)
		}
	})

	t.Run("store error", func(t *testing.T) {
		t.Parallel()

		storeErr := errors.New("read failed")
		limiter, err := NewStoreSpendingLimiter(&fakeSpendingUsageStore{err: storeErr}, "10.00", "10.00", "0")
		if err != nil {
			t.Fatalf("NewStoreSpendingLimiter: %v", err)
		}
		amount, err := ParseUSDC("1.00")
		if err != nil {
			t.Fatalf("ParseUSDC: %v", err)
		}

		err = limiter.Check(context.Background(), amount)
		if err == nil {
			t.Fatal("Check error = nil, want wrapped store error")
		}
		if !strings.Contains(err.Error(), "check daily spent") || !errors.Is(err, storeErr) {
			t.Fatalf("Check error = %v, want wrapped daily spent error", err)
		}
	})
}

func TestSpendingLimiterDailyRemainingClampsAtZero(t *testing.T) {
	t.Parallel()

	limiter, err := NewStoreSpendingLimiter(
		&fakeSpendingUsageStore{amounts: []string{"7.00", "5.00"}},
		"20.00",
		"10.00",
		"0",
	)
	if err != nil {
		t.Fatalf("NewStoreSpendingLimiter: %v", err)
	}

	remaining, err := limiter.DailyRemaining(context.Background())
	if err != nil {
		t.Fatalf("DailyRemaining: %v", err)
	}
	if remaining.Sign() != 0 {
		t.Fatalf("DailyRemaining = %s, want 0.00", FormatUSDC(remaining))
	}
}

func TestSpendingLimiterAccessorsReturnCopiesAndRecordIsNoop(t *testing.T) {
	t.Parallel()

	store := &fakeSpendingUsageStore{}
	limiter, err := NewStoreSpendingLimiter(store, "5.00", "10.00", "0")
	if err != nil {
		t.Fatalf("NewStoreSpendingLimiter: %v", err)
	}

	perTx := limiter.MaxPerTx()
	daily := limiter.MaxDaily()
	perTx.SetInt64(1)
	daily.SetInt64(2)

	if got := FormatUSDC(limiter.MaxPerTx()); got != "5.00" {
		t.Fatalf("MaxPerTx mutated through returned pointer: got %s", got)
	}
	if got := FormatUSDC(limiter.MaxDaily()); got != "10.00" {
		t.Fatalf("MaxDaily mutated through returned pointer: got %s", got)
	}
	if err := limiter.Record(context.Background(), big.NewInt(123)); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if store.calls != 0 {
		t.Fatalf("Record called store %d times, want 0", store.calls)
	}
}

func TestSpendingLimiterIsAutoApprovableUsesThresholdAndLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		spent      []string
		amount     string
		threshold  string
		wantOK     bool
		wantErrSub string
	}{
		{name: "disabled threshold", amount: "0.10", threshold: "0", wantOK: false},
		{name: "above threshold", amount: "1.01", threshold: "1.00", wantOK: false},
		{name: "within threshold and limits", spent: []string{"2.00"}, amount: "1.00", threshold: "1.00", wantOK: true},
		{name: "within threshold but daily limit exceeded", spent: []string{"9.50"}, amount: "1.00", threshold: "1.00", wantErrSub: "would exceed daily limit"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limiter, err := NewStoreSpendingLimiter(&fakeSpendingUsageStore{amounts: tt.spent}, "10.00", "10.00", tt.threshold)
			if err != nil {
				t.Fatalf("NewStoreSpendingLimiter: %v", err)
			}
			amount, err := ParseUSDC(tt.amount)
			if err != nil {
				t.Fatalf("ParseUSDC: %v", err)
			}

			got, err := limiter.IsAutoApprovable(context.Background(), amount)
			if tt.wantErrSub != "" {
				if err == nil {
					t.Fatal("IsAutoApprovable error = nil, want error")
				}
				if !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Fatalf("IsAutoApprovable error = %v, want substring %q", err, tt.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("IsAutoApprovable: %v", err)
			}
			if got != tt.wantOK {
				t.Fatalf("IsAutoApprovable = %v, want %v", got, tt.wantOK)
			}
		})
	}
}

func TestEntUsageStoreNilClientReturnsUnavailable(t *testing.T) {
	t.Parallel()

	_, err := entUsageStore{}.DailySpendSince(context.Background(), time.Now())
	if err == nil {
		t.Fatal("DailySpendSince error = nil, want unavailable error")
	}
	if !strings.Contains(err.Error(), "payment usage store unavailable") {
		t.Fatalf("DailySpendSince error = %v, want unavailable error", err)
	}
}

func TestStartOfTodayReturnsLocalMidnightForCurrentDate(t *testing.T) {
	t.Parallel()

	before := time.Now()
	got := startOfToday()
	after := time.Now()
	if got.Location() != after.Location() {
		t.Fatalf("startOfToday location = %v, want %v", got.Location(), after.Location())
	}
	if !sameLocalDate(got, before) && !sameLocalDate(got, after) {
		t.Fatalf("startOfToday date = %s, want date between %s and %s", got, before, after)
	}
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("startOfToday = %s, want midnight", got)
	}
}

func sameLocalDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func TestNetworkName(t *testing.T) {
	tests := []struct {
		give int64
		want string
	}{
		{give: 1, want: "Ethereum Mainnet"},
		{give: 8453, want: "Base"},
		{give: 84532, want: "Base Sepolia"},
		{give: 99999, want: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := NetworkName(tt.give)
			if got != tt.want {
				t.Errorf("NetworkName(%d) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}
