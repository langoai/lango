package pricing

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	t.Parallel()

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
			t.Parallel()
			_, err := New(tt.giveCfg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEngine_Quote(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
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
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantFree, quote.IsFree)
			if tt.wantFree {
				return
			}

			assert.True(t, quote.FinalPrice.Int64() >= tt.wantPriceMin && quote.FinalPrice.Int64() <= tt.wantPriceMax,
				"FinalPrice = %s, want [%d, %d]", quote.FinalPrice, tt.wantPriceMin, tt.wantPriceMax)
			assert.GreaterOrEqual(t, len(quote.Modifiers), tt.wantModMin)
		})
	}
}

func TestEngine_Quote_Fields(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))
	engine.SetReputation(mockReputation(map[string]float64{"did:key:alice": 0.5}))

	quote, err := engine.Quote(context.Background(), "search", "did:key:alice")
	require.NoError(t, err)

	assert.Equal(t, "search", quote.ToolName)
	assert.Equal(t, "USDC", quote.Currency)
	assert.Equal(t, "did:key:alice", quote.PeerDID)
	assert.Equal(t, 0, quote.BasePrice.Cmp(usdc(1)))
	assert.True(t, quote.ValidUntil.After(time.Now()))
}

func TestEngine_Quote_ReputationError(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))
	engine.SetReputation(mockReputationErr(errors.New("db down")))

	_, err := engine.Quote(context.Background(), "search", "did:key:alice")
	require.Error(t, err)
}

func TestEngine_Quote_NilReputation(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	engine.SetBasePrice("search", usdc(1))

	quote, err := engine.Quote(context.Background(), "search", "did:key:alice")
	require.NoError(t, err)
	assert.Equal(t, 0, quote.FinalPrice.Cmp(usdc(1)))
}

func TestEngine_Quote_BasePriceNotMutated(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{})
	original := usdc(1)
	engine.SetBasePrice("search", original)
	engine.SetReputation(mockReputation(map[string]float64{"peer": 0.9}))

	_, err := engine.Quote(context.Background(), "search", "peer")
	require.NoError(t, err)

	// The original value should not be mutated (SetBasePrice copies).
	assert.Equal(t, 0, original.Cmp(usdc(1)))
}

func TestEngine_SetBasePriceFromString(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			engine := newTestEngine(t, config.DynamicPricingConfig{})
			err := engine.SetBasePriceFromString("tool", tt.givePrice)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			quote, err := engine.Quote(context.Background(), "tool", "")
			require.NoError(t, err)
			assert.Equal(t, tt.wantPrice, quote.FinalPrice.Int64())
		})
	}
}

func TestEngine_AddRemoveRule(t *testing.T) {
	t.Parallel()

	engine := newTestEngine(t, config.DynamicPricingConfig{})

	engine.AddRule(PricingRule{
		Name:     "surge",
		Priority: 1,
		Enabled:  true,
		Modifier: PriceModifier{Type: ModifierSurge, Factor: 2.0},
	})

	rules := engine.Rules()
	require.Len(t, rules, 1)

	engine.RemoveRule("surge")
	rules = engine.Rules()
	assert.Len(t, rules, 0)
}

func TestHasModifierType(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			assert.Equal(t, tt.want, hasModifierType(tt.giveMods, tt.giveType))
		})
	}
}
