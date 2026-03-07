package pricing

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuleSet(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet()
	assert.Len(t, rs.Rules(), 0)
}

func TestRuleSet_Add_SortsByPriority(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet()
	rs.Add(PricingRule{Name: "low", Priority: 10, Enabled: true})
	rs.Add(PricingRule{Name: "high", Priority: 1, Enabled: true})
	rs.Add(PricingRule{Name: "mid", Priority: 5, Enabled: true})

	rules := rs.Rules()
	require.Len(t, rules, 3)
	assert.Equal(t, "high", rules[0].Name)
	assert.Equal(t, "mid", rules[1].Name)
	assert.Equal(t, "low", rules[2].Name)
}

func TestRuleSet_Remove(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet()
	rs.Add(PricingRule{Name: "a", Priority: 1, Enabled: true})
	rs.Add(PricingRule{Name: "b", Priority: 2, Enabled: true})

	rs.Remove("a")
	rules := rs.Rules()
	require.Len(t, rules, 1)
	assert.Equal(t, "b", rules[0].Name)
}

func TestRuleSet_Remove_Nonexistent(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet()
	rs.Add(PricingRule{Name: "a", Priority: 1, Enabled: true})
	rs.Remove("nonexistent")

	assert.Len(t, rs.Rules(), 1)
}

func TestRuleSet_Rules_ReturnsCopy(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet()
	rs.Add(PricingRule{Name: "a", Priority: 1, Enabled: true})

	rules := rs.Rules()
	rules[0].Name = "mutated"

	assert.Equal(t, "a", rs.Rules()[0].Name)
}

func TestRuleSet_Evaluate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		giveRules      []PricingRule
		giveToolName   string
		giveTrustScore float64
		givePeerDID    string
		giveBasePrice  int64
		wantPrice      int64
		wantModCount   int
	}{
		{
			give:          "no rules returns base price",
			giveToolName:  "search",
			giveBasePrice: 100000,
			wantPrice:     100000,
			wantModCount:  0,
		},
		{
			give: "single trust discount",
			giveRules: []PricingRule{
				{
					Name:     "trust10",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						MinTrustScore: 0.8,
					},
					Modifier: PriceModifier{
						Type:   ModifierTrustDiscount,
						Factor: 0.9,
					},
				},
			},
			giveToolName:   "search",
			giveTrustScore: 0.85,
			giveBasePrice:  100000,
			wantPrice:      90000,
			wantModCount:   1,
		},
		{
			give: "trust score below threshold skips rule",
			giveRules: []PricingRule{
				{
					Name:     "trust10",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						MinTrustScore: 0.8,
					},
					Modifier: PriceModifier{
						Type:   ModifierTrustDiscount,
						Factor: 0.9,
					},
				},
			},
			giveToolName:   "search",
			giveTrustScore: 0.5,
			giveBasePrice:  100000,
			wantPrice:      100000,
			wantModCount:   0,
		},
		{
			give: "tool pattern match",
			giveRules: []PricingRule{
				{
					Name:     "search_surge",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						ToolPattern: "search_*",
					},
					Modifier: PriceModifier{
						Type:   ModifierSurge,
						Factor: 1.2,
					},
				},
			},
			giveToolName:  "search_web",
			giveBasePrice: 100000,
			wantPrice:     120000,
			wantModCount:  1,
		},
		{
			give: "tool pattern no match",
			giveRules: []PricingRule{
				{
					Name:     "search_surge",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						ToolPattern: "search_*",
					},
					Modifier: PriceModifier{
						Type:   ModifierSurge,
						Factor: 1.2,
					},
				},
			},
			giveToolName:  "compute",
			giveBasePrice: 100000,
			wantPrice:     100000,
			wantModCount:  0,
		},
		{
			give: "peer DID match",
			giveRules: []PricingRule{
				{
					Name:     "vip_peer",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						PeerDID: "did:key:z6Mk123",
					},
					Modifier: PriceModifier{
						Type:   ModifierCustom,
						Factor: 0.5,
					},
				},
			},
			giveToolName:  "search",
			givePeerDID:   "did:key:z6Mk123",
			giveBasePrice: 100000,
			wantPrice:     50000,
			wantModCount:  1,
		},
		{
			give: "peer DID no match",
			giveRules: []PricingRule{
				{
					Name:     "vip_peer",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						PeerDID: "did:key:z6Mk123",
					},
					Modifier: PriceModifier{
						Type:   ModifierCustom,
						Factor: 0.5,
					},
				},
			},
			giveToolName:  "search",
			givePeerDID:   "did:key:z6MkOTHER",
			giveBasePrice: 100000,
			wantPrice:     100000,
			wantModCount:  0,
		},
		{
			give: "disabled rule is skipped",
			giveRules: []PricingRule{
				{
					Name:     "disabled",
					Priority: 1,
					Enabled:  false,
					Modifier: PriceModifier{
						Type:   ModifierSurge,
						Factor: 2.0,
					},
				},
			},
			giveToolName:  "search",
			giveBasePrice: 100000,
			wantPrice:     100000,
			wantModCount:  0,
		},
		{
			give: "multiple rules applied in priority order",
			giveRules: []PricingRule{
				{
					Name:     "surge",
					Priority: 1,
					Enabled:  true,
					Modifier: PriceModifier{
						Type:   ModifierSurge,
						Factor: 1.5,
					},
				},
				{
					Name:     "trust_discount",
					Priority: 2,
					Enabled:  true,
					Condition: RuleCondition{
						MinTrustScore: 0.5,
					},
					Modifier: PriceModifier{
						Type:   ModifierTrustDiscount,
						Factor: 0.8,
					},
				},
			},
			giveToolName:   "search",
			giveTrustScore: 0.9,
			giveBasePrice:  100000,
			wantPrice:      120000, // 100000 * 1.5 = 150000, then 150000 * 0.8 = 120000
			wantModCount:   2,
		},
		{
			give: "max trust score filters correctly",
			giveRules: []PricingRule{
				{
					Name:     "new_peer_surcharge",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						MaxTrustScore: 0.3,
					},
					Modifier: PriceModifier{
						Type:   ModifierSurge,
						Factor: 1.5,
					},
				},
			},
			giveToolName:   "search",
			giveTrustScore: 0.1,
			giveBasePrice:  100000,
			wantPrice:      150000,
			wantModCount:   1,
		},
		{
			give: "max trust score exceeded skips rule",
			giveRules: []PricingRule{
				{
					Name:     "new_peer_surcharge",
					Priority: 1,
					Enabled:  true,
					Condition: RuleCondition{
						MaxTrustScore: 0.3,
					},
					Modifier: PriceModifier{
						Type:   ModifierSurge,
						Factor: 1.5,
					},
				},
			},
			giveToolName:   "search",
			giveTrustScore: 0.8,
			giveBasePrice:  100000,
			wantPrice:      100000,
			wantModCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			rs := NewRuleSet()
			for _, r := range tt.giveRules {
				rs.Add(r)
			}

			gotPrice, gotMods := rs.Evaluate(
				tt.giveToolName,
				tt.giveTrustScore,
				tt.givePeerDID,
				big.NewInt(tt.giveBasePrice),
			)

			assert.Equal(t, 0, gotPrice.Cmp(big.NewInt(tt.wantPrice)), "price = %s, want %d", gotPrice, tt.wantPrice)
			assert.Len(t, gotMods, tt.wantModCount)
		})
	}
}

func TestRuleSet_Evaluate_DoesNotMutateBasePrice(t *testing.T) {
	t.Parallel()

	rs := NewRuleSet()
	rs.Add(PricingRule{
		Name:     "discount",
		Priority: 1,
		Enabled:  true,
		Modifier: PriceModifier{Factor: 0.5},
	})

	basePrice := big.NewInt(100000)
	rs.Evaluate("tool", 0.5, "", basePrice)

	assert.Equal(t, 0, basePrice.Cmp(big.NewInt(100000)))
}
