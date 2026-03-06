package pricing

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
)

func usdc(n int64) *big.Int {
	return big.NewInt(n * 1_000_000) // 6 decimal places
}

func newTestEngine(t *testing.T, cfg config.DynamicPricingConfig) *Engine {
	t.Helper()
	e, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return e
}

func mockReputation(scores map[string]float64) ReputationQuerier {
	return func(_ context.Context, peerDID string) (float64, error) {
		return scores[peerDID], nil
	}
}

func mockReputationErr(e error) ReputationQuerier {
	return func(_ context.Context, _ string) (float64, error) {
		return 0, e
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		give    string
		giveCfg config.DynamicPricingConfig
		wantErr bool
	}{
		{
			give:    "default config",
			giveCfg: config.DynamicPricingConfig{},
		},
		{
			give: "with min price",
			giveCfg: config.DynamicPricingConfig{
				MinPrice: "0.01",
			},
		},
		{
			give: "invalid min price",
			giveCfg: config.DynamicPricingConfig{
				MinPrice: "not-a-number",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			_, err := New(tt.giveCfg)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEngine_Quote(t *testing.T) {
	tests := []struct {
		give         string
		givePrices   map[string]*big.Int
		giveTool     string
		givePeerDID  string
		giveRepFn    ReputationQuerier
		giveRules    []PricingRule
		giveCfg      config.DynamicPricingConfig
		wantFree     bool
		wantPriceMin int64
		wantPriceMax int64
		wantModMin   int
		wantErr      bool
	}{
		{
			give:       "free tool (not in price list)",
			givePrices: map[string]*big.Int{"search": usdc(1)},
			giveTool:   "unknown_tool",
			wantFree:   true,
		},
		{
			give:       "zero-priced tool is free",
			givePrices: map[string]*big.Int{"free_tool": big.NewInt(0)},
			giveTool:   "free_tool",
			wantFree:   true,
		},
		{
			give:         "base price without reputation",
			givePrices:   map[string]*big.Int{"search": usdc(1)},
			giveTool:     "search",
			wantFree:     false,
			wantPriceMin: 1_000_000,
			wantPriceMax: 1_000_000,
			wantModMin:   0,
		},
		{
			give:         "trust discount applied for high-trust peer",
			givePrices:   map[string]*big.Int{"search": usdc(1)},
			giveTool:     "search",
			givePeerDID:  "did:key:trusted",
			giveRepFn:    mockReputation(map[string]float64{"did:key:trusted": 0.9}),
			giveCfg:      config.DynamicPricingConfig{TrustDiscount: 0.10},
			wantFree:     false,
			wantPriceMin: 900_000, // 10% discount
			wantPriceMax: 900_000,
			wantModMin:   1,
		},
		{
			give:         "no discount for low trust",
			givePrices:   map[string]*big.Int{"search": usdc(1)},
			giveTool:     "search",
			givePeerDID:  "did:key:new",
			giveRepFn:    mockReputation(map[string]float64{"did:key:new": 0.5}),
			wantFree:     false,
			wantPriceMin: 1_000_000,
			wantPriceMax: 1_000_000,
			wantModMin:   0,
		},
		{
			give:         "no discount for zero trust",
			givePrices:   map[string]*big.Int{"search": usdc(1)},
			giveTool:     "search",
			givePeerDID:  "did:key:new",
			giveRepFn:    mockReputation(map[string]float64{"did:key:new": 0.0}),
			wantFree:     false,
			wantPriceMin: 1_000_000,
			wantPriceMax: 1_000_000,
			wantModMin:   0,
		},
		{
			give:       "rule-based surge pricing",
			givePrices: map[string]*big.Int{"compute": usdc(2)},
			giveTool:   "compute",
			giveRules: []PricingRule{
				{
					Name:     "compute_surge",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						ToolPattern: "compute",
					},
					Modifier: PriceModifier{
						Type:   ModifierSurge,
						Factor: 1.5,
					},
				},
			},
			wantFree:     false,
			wantPriceMin: 3_000_000, // 2 USDC * 1.5
			wantPriceMax: 3_000_000,
			wantModMin:   1,
		},
		{
			give:        "rule trust discount suppresses auto trust discount",
			givePrices:  map[string]*big.Int{"search": usdc(1)},
			giveTool:    "search",
			givePeerDID: "did:key:trusted",
			giveRepFn:   mockReputation(map[string]float64{"did:key:trusted": 0.9}),
			giveRules: []PricingRule{
				{
					Name:     "explicit_trust",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						MinTrustScore: 0.8,
					},
					Modifier: PriceModifier{
						Type:   ModifierTrustDiscount,
						Factor: 0.8,
					},
				},
			},
			wantFree:     false,
			wantPriceMin: 800_000, // 20% rule discount, no auto discount
			wantPriceMax: 800_000,
			wantModMin:   1,
		},
		{
			give:        "min price floor enforced",
			givePrices:  map[string]*big.Int{"search": big.NewInt(100)},
			giveTool:    "search",
			givePeerDID: "did:key:trusted",
			giveRepFn:   mockReputation(map[string]float64{"did:key:trusted": 0.95}),
			giveCfg: config.DynamicPricingConfig{
				MinPrice:      "0.000050", // 50 units
				TrustDiscount: 0.5,        // large discount to push below floor
			},
			wantFree:     false,
			wantPriceMin: 50,
			wantPriceMax: 100,
			wantModMin:   0,
		},
		{
			give:       "multiple modifiers stacking",
			givePrices: map[string]*big.Int{"search": usdc(10)},
			giveTool:   "search",
			giveRules: []PricingRule{
				{
					Name:     "surge",
					Priority: 1,
					Enabled:  true,
					Modifier: PriceModifier{Type: ModifierSurge, Factor: 1.5},
				},
				{
					Name:     "volume",
					Priority: 2,
					Enabled:  true,
					Modifier: PriceModifier{Type: ModifierVolumeDiscount, Factor: 0.8},
				},
			},
			wantFree:     false,
			wantPriceMin: 12_000_000, // 10 * 1.5 = 15, * 0.8 = 12
			wantPriceMax: 12_000_000,
			wantModMin:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			engine := newTestEngine(t, tt.giveCfg)
			for name, price := range tt.givePrices {
				engine.SetBasePrice(name, price)
			}
			if tt.giveRepFn != nil {
				engine.SetReputation(tt.giveRepFn)
			}
			for _, r := range tt.giveRules {
				engine.AddRule(r)
			}

			quote, err := engine.Quote(context.Background(), tt.giveTool, tt.givePeerDID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if quote.IsFree != tt.wantFree {
				t.Errorf("IsFree = %v, want %v", quote.IsFree, tt.wantFree)
			}
			if tt.wantFree {
				return
			}

			if quote.FinalPrice.Int64() < tt.wantPriceMin || quote.FinalPrice.Int64() > tt.wantPriceMax {
				t.Errorf("FinalPrice = %s, want [%d, %d]",
					quote.FinalPrice, tt.wantPriceMin, tt.wantPriceMax)
			}
			if len(quote.Modifiers) < tt.wantModMin {
				t.Errorf("modifier count = %d, want >= %d", len(quote.Modifiers), tt.wantModMin)
			}
		})
	}
}

func TestEngine_Quote_Fields(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))
	engine.SetReputation(mockReputation(map[string]float64{"did:key:alice": 0.5}))

	quote, err := engine.Quote(context.Background(), "search", "did:key:alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if quote.ToolName != "search" {
		t.Errorf("ToolName = %q, want %q", quote.ToolName, "search")
	}
	if quote.Currency != "USDC" {
		t.Errorf("Currency = %q, want %q", quote.Currency, "USDC")
	}
	if quote.PeerDID != "did:key:alice" {
		t.Errorf("PeerDID = %q, want %q", quote.PeerDID, "did:key:alice")
	}
	if quote.BasePrice.Cmp(usdc(1)) != 0 {
		t.Errorf("BasePrice = %s, want %s", quote.BasePrice, usdc(1))
	}
	if quote.ValidUntil.Before(time.Now()) {
		t.Error("ValidUntil should be in the future")
	}
}

func TestEngine_Quote_ReputationError(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))
	engine.SetReputation(mockReputationErr(errors.New("db down")))

	_, err := engine.Quote(context.Background(), "search", "did:key:alice")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEngine_Quote_NilReputation(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))

	quote, err := engine.Quote(context.Background(), "search", "did:key:alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quote.FinalPrice.Cmp(usdc(1)) != 0 {
		t.Errorf("FinalPrice = %s, want %s (no discount without reputation)", quote.FinalPrice, usdc(1))
	}
}

func TestEngine_Quote_BasePriceNotMutated(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{})
	original := usdc(1)
	engine.SetBasePrice("search", original)
	engine.SetReputation(mockReputation(map[string]float64{"peer": 0.9}))

	_, err := engine.Quote(context.Background(), "search", "peer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The original value should not be mutated (SetBasePrice copies).
	if original.Cmp(usdc(1)) != 0 {
		t.Errorf("original price mutated to %s", original)
	}
}

func TestEngine_SetBasePriceFromString(t *testing.T) {
	tests := []struct {
		give      string
		givePrice string
		wantErr   bool
		wantPrice int64
	}{
		{give: "valid decimal", givePrice: "1.50", wantPrice: 1_500_000},
		{give: "integer", givePrice: "5", wantPrice: 5_000_000},
		{give: "small amount", givePrice: "0.01", wantPrice: 10_000},
		{give: "invalid", givePrice: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			engine := newTestEngine(t, config.DynamicPricingConfig{})
			err := engine.SetBasePriceFromString("tool", tt.givePrice)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			quote, err := engine.Quote(context.Background(), "tool", "")
			if err != nil {
				t.Fatalf("Quote error: %v", err)
			}
			if quote.FinalPrice.Int64() != tt.wantPrice {
				t.Errorf("FinalPrice = %d, want %d", quote.FinalPrice.Int64(), tt.wantPrice)
			}
		})
	}
}

func TestEngine_AddRemoveRule(t *testing.T) {
	engine := newTestEngine(t, config.DynamicPricingConfig{})

	engine.AddRule(PricingRule{
		Name:     "surge",
		Priority: 1,
		Enabled:  true,
		Modifier: PriceModifier{Type: ModifierSurge, Factor: 2.0},
	})

	rules := engine.Rules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}

	engine.RemoveRule("surge")
	rules = engine.Rules()
	if len(rules) != 0 {
		t.Errorf("expected 0 rules after remove, got %d", len(rules))
	}
}

func TestHasModifierType(t *testing.T) {
	tests := []struct {
		give     string
		giveMods []PriceModifier
		giveType PriceModifierType
		want     bool
	}{
		{
			give:     "empty list",
			giveMods: nil,
			giveType: ModifierSurge,
			want:     false,
		},
		{
			give:     "type present",
			giveMods: []PriceModifier{{Type: ModifierSurge}},
			giveType: ModifierSurge,
			want:     true,
		},
		{
			give:     "type not present",
			giveMods: []PriceModifier{{Type: ModifierSurge}},
			giveType: ModifierTrustDiscount,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := hasModifierType(tt.giveMods, tt.giveType)
			if got != tt.want {
				t.Errorf("hasModifierType() = %v, want %v", got, tt.want)
			}
		})
	}
}
