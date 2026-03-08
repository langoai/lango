package policy

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// HarnessPolicy defines the off-chain harness constraints.
type HarnessPolicy struct {
	MaxTxAmount       *big.Int         `json:"maxTxAmount"`
	DailyLimit        *big.Int         `json:"dailyLimit"`
	MonthlyLimit      *big.Int         `json:"monthlyLimit"`
	AllowedTargets    []common.Address `json:"allowedTargets"`
	AllowedFunctions  []string         `json:"allowedFunctions"`
	RequiredRiskScore float64          `json:"requiredRiskScore"`
	AutoApproveBelow  *big.Int         `json:"autoApproveBelow"`
}

// SpendTracker tracks cumulative spending.
type SpendTracker struct {
	DailySpent       *big.Int  `json:"dailySpent"`
	MonthlySpent     *big.Int  `json:"monthlySpent"`
	LastDailyReset   time.Time `json:"lastDailyReset"`
	LastMonthlyReset time.Time `json:"lastMonthlyReset"`
}

// NewSpendTracker creates a zeroed spend tracker with current reset times.
func NewSpendTracker() *SpendTracker {
	now := time.Now()
	return &SpendTracker{
		DailySpent:       new(big.Int),
		MonthlySpent:     new(big.Int),
		LastDailyReset:   now,
		LastMonthlyReset: now,
	}
}

// ResetIfNeeded resets daily/monthly counters if their windows have expired.
func (st *SpendTracker) ResetIfNeeded(now time.Time) {
	if now.Sub(st.LastDailyReset) >= 24*time.Hour {
		st.DailySpent = new(big.Int)
		st.LastDailyReset = now
	}
	if now.Sub(st.LastMonthlyReset) >= 30*24*time.Hour {
		st.MonthlySpent = new(big.Int)
		st.LastMonthlyReset = now
	}
}
