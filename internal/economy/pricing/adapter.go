package pricing

import (
	"context"
	"fmt"
	"math/big"
)

// USDCDecimals is the number of decimal places for USDC.
const USDCDecimals = 6

// AdaptToPricingFunc returns a function compatible with paygate.PricingFunc.
// Signature: func(toolName string) (price string, isFree bool)
// Uses a background context and empty peerDID for anonymous pricing lookups.
func (e *Engine) AdaptToPricingFunc() func(toolName string) (string, bool) {
	return func(toolName string) (string, bool) {
		quote, err := e.Quote(context.Background(), toolName, "")
		if err != nil || quote.IsFree {
			return "", true
		}
		return formatUSDC(quote.FinalPrice), false
	}
}

// AdaptToPricingFuncWithPeer returns a paygate-compatible PricingFunc that
// includes peer identity for trust-based pricing.
func (e *Engine) AdaptToPricingFuncWithPeer(peerDID string) func(toolName string) (string, bool) {
	return func(toolName string) (string, bool) {
		quote, err := e.Quote(context.Background(), toolName, peerDID)
		if err != nil || quote.IsFree {
			return "", true
		}
		return formatUSDC(quote.FinalPrice), false
	}
}

// formatUSDC converts smallest USDC units to decimal string.
// e.g., 1500000 → "1.50", 0 → "0.00", 50 → "0.000050"
func formatUSDC(amount *big.Int) string {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(USDCDecimals), nil)
	whole := new(big.Int).Div(amount, divisor)
	remainder := new(big.Int).Mod(amount, divisor)

	if remainder.Sign() == 0 {
		return fmt.Sprintf("%s.00", whole)
	}

	// Format remainder with leading zeros, then trim trailing zeros.
	fracStr := fmt.Sprintf("%06d", remainder.Int64())
	// Trim trailing zeros but keep at least 2 decimal places.
	trimmed := fracStr
	for len(trimmed) > 2 && trimmed[len(trimmed)-1] == '0' {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return fmt.Sprintf("%s.%s", whole, trimmed)
}

// MapToolPricer provides a simple way to set base prices from a map during
// engine construction. Call SetBasePrice on the engine directly for runtime updates.
type MapToolPricer struct {
	prices     map[string]*big.Int
	defaultVal *big.Int
}

// NewMapToolPricer creates a MapToolPricer backed by a map. Tools not in the map
// use the default price. If defaultPrice is nil, unlisted tools have no price.
func NewMapToolPricer(prices map[string]*big.Int, defaultPrice *big.Int) *MapToolPricer {
	copied := make(map[string]*big.Int, len(prices))
	for k, v := range prices {
		copied[k] = new(big.Int).Set(v)
	}
	return &MapToolPricer{
		prices:     copied,
		defaultVal: defaultPrice,
	}
}

// LoadInto sets all prices from this pricer into the engine.
func (m *MapToolPricer) LoadInto(e *Engine) {
	for name, price := range m.prices {
		e.SetBasePrice(name, price)
	}
}
