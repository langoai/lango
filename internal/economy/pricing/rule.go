package pricing

import (
	"math/big"
	"path"
	"sort"
)

// RuleCondition defines when a pricing rule applies.
type RuleCondition struct {
	ToolPattern   string  `json:"toolPattern,omitempty"`   // glob pattern for tool name
	MinTrustScore float64 `json:"minTrustScore,omitempty"`
	MaxTrustScore float64 `json:"maxTrustScore,omitempty"`
	PeerDID       string  `json:"peerDid,omitempty"` // specific peer
}

// PricingRule defines a pricing rule with condition and modifier.
type PricingRule struct {
	Name      string        `json:"name"`
	Priority  int           `json:"priority"` // lower = higher priority
	Condition RuleCondition `json:"condition"`
	Modifier  PriceModifier `json:"modifier"`
	Enabled   bool          `json:"enabled"`
}

// RuleSet holds an ordered collection of pricing rules.
type RuleSet struct {
	rules []PricingRule
}

// NewRuleSet creates a new empty RuleSet.
func NewRuleSet() *RuleSet {
	return &RuleSet{}
}

// Add inserts a rule and keeps the list sorted by priority.
func (rs *RuleSet) Add(rule PricingRule) {
	rs.rules = append(rs.rules, rule)
	sort.Slice(rs.rules, func(i, j int) bool {
		return rs.rules[i].Priority < rs.rules[j].Priority
	})
}

// Remove deletes a rule by name.
func (rs *RuleSet) Remove(name string) {
	for i, r := range rs.rules {
		if r.Name == name {
			rs.rules = append(rs.rules[:i], rs.rules[i+1:]...)
			return
		}
	}
}

// Rules returns a copy of all rules sorted by priority.
func (rs *RuleSet) Rules() []PricingRule {
	out := make([]PricingRule, len(rs.rules))
	copy(out, rs.rules)
	return out
}

// Evaluate walks rules in priority order, applies matching modifiers, and
// returns the final price and the list of applied modifiers.
func (rs *RuleSet) Evaluate(toolName string, trustScore float64, peerDID string, basePrice *big.Int) (*big.Int, []PriceModifier) {
	price := new(big.Int).Set(basePrice)
	var applied []PriceModifier

	for _, r := range rs.rules {
		if !r.Enabled {
			continue
		}
		if !matchesCondition(r.Condition, toolName, trustScore, peerDID) {
			continue
		}
		price = applyModifier(price, r.Modifier.Factor)
		applied = append(applied, r.Modifier)
	}

	return price, applied
}

// matchesCondition checks whether the given context satisfies the rule condition.
func matchesCondition(c RuleCondition, toolName string, trustScore float64, peerDID string) bool {
	if c.ToolPattern != "" {
		matched, err := path.Match(c.ToolPattern, toolName)
		if err != nil || !matched {
			return false
		}
	}
	if c.MinTrustScore != 0 && trustScore < c.MinTrustScore {
		return false
	}
	if c.MaxTrustScore != 0 && trustScore > c.MaxTrustScore {
		return false
	}
	if c.PeerDID != "" && c.PeerDID != peerDID {
		return false
	}
	return true
}

// applyModifier multiplies the price by the given factor using integer arithmetic.
// Factor is a float64 multiplier (e.g., 0.9 for 10% discount).
func applyModifier(price *big.Int, factor float64) *big.Int {
	// Convert factor to basis points (10000 = 1.0x) for integer arithmetic.
	bps := int64(factor * 10000)
	result := new(big.Int).Mul(price, big.NewInt(bps))
	result.Div(result, big.NewInt(10000))
	return result
}
