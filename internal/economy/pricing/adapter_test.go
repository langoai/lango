package pricing

import (
	"math/big"
	"testing"

	"github.com/langoai/lango/internal/config"
)

func TestAdaptToPricingFunc_FreeTool(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{})

	fn := engine.AdaptToPricingFunc()
	price, isFree := fn("unknown")
	if !isFree {
		t.Error("expected free tool")
	}
	if price != "" {
		t.Errorf("price = %q, want empty", price)
	}
}

func TestAdaptToPricingFunc_PaidTool(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))

	fn := engine.AdaptToPricingFunc()
	price, isFree := fn("search")
	if isFree {
		t.Error("expected paid tool")
	}
	if price != "1.00" {
		t.Errorf("price = %q, want %q", price, "1.00")
	}
}

func TestAdaptToPricingFuncWithPeer(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{TrustDiscount: 0.10})
	engine.SetBasePrice("search", usdc(1))
	engine.SetReputation(mockReputation(map[string]float64{"did:key:alice": 0.9}))

	fn := engine.AdaptToPricingFuncWithPeer("did:key:alice")
	price, isFree := fn("search")
	if isFree {
		t.Error("expected paid tool")
	}
	// Trust score 0.9 > 0.8 threshold → 10% discount → 0.90
	if price != "0.90" {
		t.Errorf("price = %q, want %q", price, "0.90")
	}
}

func TestFormatUSDC(t *testing.T) {
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
			got := formatUSDC(tt.give)
			if got != tt.want {
				t.Errorf("formatUSDC(%s) = %q, want %q", tt.give, got, tt.want)
			}
		})
	}
}

func TestMapToolPricer_LoadInto(t *testing.T) {
	prices := map[string]*big.Int{
		"search":  usdc(1),
		"compute": usdc(5),
	}
	pricer := NewMapToolPricer(prices, nil)

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	pricer.LoadInto(engine)

	quote, err := engine.Quote(nil, "search", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.FinalPrice.Cmp(usdc(1)) != 0 {
		t.Errorf("search price = %s, want %s", quote.FinalPrice, usdc(1))
	}

	quote, err = engine.Quote(nil, "compute", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.FinalPrice.Cmp(usdc(5)) != 0 {
		t.Errorf("compute price = %s, want %s", quote.FinalPrice, usdc(5))
	}
}

func TestMapToolPricer_DefensiveCopy(t *testing.T) {
	original := map[string]*big.Int{"search": usdc(1)}
	pricer := NewMapToolPricer(original, nil)

	// Mutate original map value.
	original["search"].SetInt64(0)

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	pricer.LoadInto(engine)

	quote, err := engine.Quote(nil, "search", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.FinalPrice.Cmp(usdc(1)) != 0 {
		t.Errorf("pricer should have defensive copy, got %s", quote.FinalPrice)
	}
}
