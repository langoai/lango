package policy

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpendTracker(t *testing.T) {
	t.Parallel()

	before := time.Now()
	st := NewSpendTracker()
	after := time.Now()

	require.NotNil(t, st)
	assert.Equal(t, int64(0), st.DailySpent.Int64())
	assert.Equal(t, int64(0), st.MonthlySpent.Int64())
	assert.False(t, st.LastDailyReset.Before(before))
	assert.False(t, st.LastDailyReset.After(after))
	assert.False(t, st.LastMonthlyReset.Before(before))
	assert.False(t, st.LastMonthlyReset.After(after))
}

func TestSpendTracker_ResetIfNeeded(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give             string
		dailySpent       int64
		monthlySpent     int64
		lastDailyReset   time.Duration // offset from "now"
		lastMonthlyReset time.Duration // offset from "now"
		wantDailyReset   bool
		wantMonthlyReset bool
	}{
		{
			give:             "no_reset_within_windows",
			dailySpent:       500,
			monthlySpent:     2000,
			lastDailyReset:   -12 * time.Hour,
			lastMonthlyReset: -15 * 24 * time.Hour,
			wantDailyReset:   false,
			wantMonthlyReset: false,
		},
		{
			give:             "daily_reset_only",
			dailySpent:       500,
			monthlySpent:     2000,
			lastDailyReset:   -25 * time.Hour,
			lastMonthlyReset: -15 * 24 * time.Hour,
			wantDailyReset:   true,
			wantMonthlyReset: false,
		},
		{
			give:             "monthly_reset_only",
			dailySpent:       500,
			monthlySpent:     2000,
			lastDailyReset:   -12 * time.Hour,
			lastMonthlyReset: -31 * 24 * time.Hour,
			wantDailyReset:   false,
			wantMonthlyReset: true,
		},
		{
			give:             "both_reset",
			dailySpent:       500,
			monthlySpent:     2000,
			lastDailyReset:   -25 * time.Hour,
			lastMonthlyReset: -31 * 24 * time.Hour,
			wantDailyReset:   true,
			wantMonthlyReset: true,
		},
		{
			give:             "exact_daily_boundary",
			dailySpent:       100,
			monthlySpent:     100,
			lastDailyReset:   -24 * time.Hour,
			lastMonthlyReset: -1 * time.Hour,
			wantDailyReset:   true,
			wantMonthlyReset: false,
		},
		{
			give:             "exact_monthly_boundary",
			dailySpent:       100,
			monthlySpent:     100,
			lastDailyReset:   -1 * time.Hour,
			lastMonthlyReset: -30 * 24 * time.Hour,
			wantDailyReset:   false,
			wantMonthlyReset: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			now := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)

			st := &SpendTracker{
				DailySpent:       big.NewInt(tt.dailySpent),
				MonthlySpent:     big.NewInt(tt.monthlySpent),
				LastDailyReset:   now.Add(tt.lastDailyReset),
				LastMonthlyReset: now.Add(tt.lastMonthlyReset),
			}

			st.ResetIfNeeded(now)

			if tt.wantDailyReset {
				assert.Equal(t, int64(0), st.DailySpent.Int64())
				assert.Equal(t, now, st.LastDailyReset)
			} else {
				assert.Equal(t, tt.dailySpent, st.DailySpent.Int64())
			}

			if tt.wantMonthlyReset {
				assert.Equal(t, int64(0), st.MonthlySpent.Int64())
				assert.Equal(t, now, st.LastMonthlyReset)
			} else {
				assert.Equal(t, tt.monthlySpent, st.MonthlySpent.Int64())
			}
		})
	}
}
