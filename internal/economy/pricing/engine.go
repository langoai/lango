package pricing

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/wallet"
)

// ReputationQuerier queries peer trust scores. Defined locally to avoid import cycles.
type ReputationQuerier func(ctx context.Context, peerDID string) (float64, error)

// DefaultQuoteExpiry is how long a price quote remains valid.
const DefaultQuoteExpiry = 5 * time.Minute

// Engine computes dynamic prices using rule-based evaluation.
type Engine struct {
	mu         sync.RWMutex
	ruleSet    *RuleSet
	cfg        config.DynamicPricingConfig
	reputation ReputationQuerier
	basePrices map[string]*big.Int // toolName -> base price in smallest USDC units
	minPrice   *big.Int
}

// New creates a pricing engine from config.
func New(cfg config.DynamicPricingConfig) (*Engine, error) {
	var minPrice *big.Int
	if cfg.MinPrice != "" {
		parsed, err := wallet.ParseUSDC(cfg.MinPrice)
		if err != nil {
			return nil, fmt.Errorf("parse min price %q: %w", cfg.MinPrice, err)
		}
		minPrice = parsed
	}
	if minPrice == nil {
		minPrice = new(big.Int)
	}

	if cfg.TrustDiscount == 0 {
		cfg.TrustDiscount = 0.10
	}
	if cfg.VolumeDiscount == 0 {
		cfg.VolumeDiscount = 0.05
	}

	return &Engine{
		ruleSet:    NewRuleSet(),
		cfg:        cfg,
		basePrices: make(map[string]*big.Int),
		minPrice:   minPrice,
	}, nil
}

// SetReputation sets the reputation querier for trust-based discounts.
func (e *Engine) SetReputation(fn ReputationQuerier) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.reputation = fn
}

// SetBasePrice sets the base price for a tool in smallest USDC units.
func (e *Engine) SetBasePrice(toolName string, price *big.Int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.basePrices[toolName] = new(big.Int).Set(price)
}

// SetBasePriceFromString parses a USDC decimal string and sets the base price.
func (e *Engine) SetBasePriceFromString(toolName, price string) error {
	parsed, err := wallet.ParseUSDC(price)
	if err != nil {
		return fmt.Errorf("parse price %q for %q: %w", price, toolName, err)
	}
	e.SetBasePrice(toolName, parsed)
	return nil
}

// AddRule adds a pricing rule to the engine.
func (e *Engine) AddRule(rule PricingRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ruleSet.Add(rule)
}

// RemoveRule removes a pricing rule by name.
func (e *Engine) RemoveRule(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ruleSet.Remove(name)
}

// Rules returns a snapshot of the current pricing rules.
func (e *Engine) Rules() []PricingRule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ruleSet.Rules()
}

// Quote computes a price quote for a tool invocation.
func (e *Engine) Quote(ctx context.Context, toolName, peerDID string) (*Quote, error) {
	e.mu.RLock()
	basePrice, ok := e.basePrices[toolName]
	if ok {
		basePrice = new(big.Int).Set(basePrice)
	}
	repFn := e.reputation
	e.mu.RUnlock()

	// Tool not priced or zero price → free.
	if !ok || basePrice.Sign() == 0 {
		return &Quote{
			ToolName:   toolName,
			BasePrice:  new(big.Int),
			FinalPrice: new(big.Int),
			Currency:   "USDC",
			IsFree:     true,
			ValidUntil: time.Now().Add(DefaultQuoteExpiry),
			PeerDID:    peerDID,
		}, nil
	}

	// Query reputation.
	trustScore, err := e.getTrustScore(ctx, repFn, peerDID)
	if err != nil {
		return nil, fmt.Errorf("get trust score for %q: %w", peerDID, err)
	}

	// Evaluate rules.
	e.mu.RLock()
	finalPrice, modifiers := e.ruleSet.Evaluate(toolName, trustScore, peerDID, basePrice)
	e.mu.RUnlock()

	// Apply trust discount if no explicit trust rule was matched and trust is high enough.
	if !hasModifierType(modifiers, ModifierTrustDiscount) && trustScore > 0.8 {
		factor := 1.0 - e.cfg.TrustDiscount
		finalPrice = applyModifier(finalPrice, factor)
		modifiers = append(modifiers, PriceModifier{
			Type:        ModifierTrustDiscount,
			Description: fmt.Sprintf("trust discount (score=%.2f, factor=%.2f)", trustScore, factor),
			Factor:      factor,
		})
	}

	// Enforce minimum price floor.
	if finalPrice.Cmp(e.minPrice) < 0 {
		finalPrice = new(big.Int).Set(e.minPrice)
	}

	return &Quote{
		ToolName:   toolName,
		BasePrice:  new(big.Int).Set(basePrice),
		FinalPrice: finalPrice,
		Currency:   "USDC",
		Modifiers:  modifiers,
		IsFree:     finalPrice.Sign() == 0,
		ValidUntil: time.Now().Add(DefaultQuoteExpiry),
		PeerDID:    peerDID,
	}, nil
}

// getTrustScore retrieves the trust score, returning 0 if no reputation querier is set.
func (e *Engine) getTrustScore(ctx context.Context, repFn ReputationQuerier, peerDID string) (float64, error) {
	if repFn == nil || peerDID == "" {
		return 0, nil
	}
	return repFn(ctx, peerDID)
}

// hasModifierType checks if any modifier in the list matches the given type.
func hasModifierType(mods []PriceModifier, t PriceModifierType) bool {
	for _, m := range mods {
		if m.Type == t {
			return true
		}
	}
	return false
}
