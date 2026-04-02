// Package finance provides shared financial types and utilities for USDC
// operations. It is a leaf package with no internal dependencies, designed
// to break the coupling between wallet (key management) and packages that
// only need monetary parsing/formatting.
package finance

import (
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"
)

// USDCDecimals is the number of decimal places for USDC (6).
const USDCDecimals = 6

// CurrencyUSDC is the ticker symbol for the USDC stablecoin.
const CurrencyUSDC = "USDC"

// usdcMultiplier is 10^6, used to convert between dollar amounts and micro-units.
var usdcMultiplier = decimal.New(1, USDCDecimals)

// ParseUSDC converts a decimal string (e.g. "1.50") to the smallest USDC unit.
// Returns an error if the string is not a valid decimal or has more than 6
// decimal places.
func ParseUSDC(amount string) (*big.Int, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid USDC amount: %q", amount)
	}

	micro := d.Mul(usdcMultiplier)
	if !micro.Equal(micro.Truncate(0)) {
		return nil, fmt.Errorf("USDC amount %q has too many decimal places", amount)
	}

	return micro.BigInt(), nil
}

// FormatUSDC converts smallest USDC units back to a decimal string.
// Trailing zeros are trimmed but at least 2 decimal places are kept.
// e.g., 1500000 -> "1.50", 0 -> "0.00", 50 -> "0.00005"
func FormatUSDC(amount *big.Int) string {
	d := decimal.NewFromBigInt(amount, 0).Div(usdcMultiplier)

	// Use full precision string, then trim trailing zeros keeping at least 2.
	s := d.StringFixed(USDCDecimals)
	// Trim trailing zeros but keep at least 2 decimal places.
	dotIdx := len(s)
	for i, c := range s {
		if c == '.' {
			dotIdx = i
			break
		}
	}
	trimmed := s
	minLen := dotIdx + 3 // "X.00" = dot + 2 digits
	for len(trimmed) > minLen && trimmed[len(trimmed)-1] == '0' {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return trimmed
}

// FloatToMicroUSDC converts a float64 dollar amount to the smallest USDC unit
// (6 decimals). Uses shopspring/decimal for exact conversion — no floating-point
// rounding hacks needed.
func FloatToMicroUSDC(amount float64) *big.Int {
	d := decimal.NewFromFloat(amount)
	micro := d.Mul(usdcMultiplier).Round(0)
	return micro.BigInt()
}
