package pricing

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/finance"
)

func TestAdaptToPricingFunc_FreeTool(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{})

	fn := engine.AdaptToPricingFunc()
	price, isFree := fn("unknown")
	assert.True(t, isFree)
	assert.Empty(t, price)
}

func TestAdaptToPricingFunc_PaidTool(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))

	fn := engine.AdaptToPricingFunc()
	price, isFree := fn("search")
	assert.False(t, isFree)
	assert.Equal(t, "1.00", price)
}

func TestAdaptToPricingFuncWithPeer(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{TrustDiscount: 0.10})
	engine.SetBasePrice("search", usdc(1))
	engine.SetReputation(mockReputation(map[string]float64{"did:key:alice": 0.9}))

	fn := engine.AdaptToPricingFuncWithPeer("did:key:alice")
	price, isFree := fn("search")
	assert.False(t, isFree)
	// Trust score 0.9 > 0.8 threshold -> 10% discount -> 0.90
	assert.Equal(t, "0.90", price)
}

func TestFormatUSDC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give *big.Int
		want string
	}{
		{give: big.NewInt(0), want: "0.00"},
		{give: big.NewInt(1_000_000), want: "1.00"},
		{give: big.NewInt(1_500_000), want: "1.50"},
		{give: big.NewInt(10_000), want: "0.01"},
		{give: big.NewInt(50), want: "0.00005"},
		{give: big.NewInt(100_000_000), want: "100.00"},
		{give: big.NewInt(1_234_567), want: "1.234567"},
		{give: big.NewInt(500_000), want: "0.50"},
		{give: big.NewInt(1_200_000), want: "1.20"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, finance.FormatUSDC(tt.give))
		})
	}
}

func TestMapToolPricer_LoadInto(t *testing.T) {
	t.Parallel()

	prices := map[string]*big.Int{
		"search":  usdc(1),
		"compute": usdc(5),
	}
	pricer := NewMapToolPricer(prices, nil)

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	pricer.LoadInto(engine)

	quote, err := engine.Quote(context.Background(), "search", "")
	require.NoError(t, err)
	assert.Equal(t, 0, quote.FinalPrice.Cmp(usdc(1)))

	quote, err = engine.Quote(context.Background(), "compute", "")
	require.NoError(t, err)
	assert.Equal(t, 0, quote.FinalPrice.Cmp(usdc(5)))
}

func TestMapToolPricer_DefensiveCopy(t *testing.T) {
	t.Parallel()

	original := map[string]*big.Int{"search": usdc(1)}
	pricer := NewMapToolPricer(original, nil)

	// Mutate original map value.
	original["search"].SetInt64(0)

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	pricer.LoadInto(engine)

	quote, err := engine.Quote(context.Background(), "search", "")
	require.NoError(t, err)
	assert.Equal(t, 0, quote.FinalPrice.Cmp(usdc(1)))
}
