package app

import (
	"math/big"

	"github.com/langoai/lango/internal/finance"
)

// floatToMicroUSDC converts a float64 dollar amount to USDC smallest unit (6 decimals).
// 1 USDC = 1_000_000 micro-units.
func floatToMicroUSDC(amount float64) *big.Int {
	return finance.FloatToMicroUSDC(amount)
}
