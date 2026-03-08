package risk

import (
	"context"
	"math/big"
	"time"
)

// SessionPolicyRecommendation holds risk-driven policy parameters.
// This is a local struct to avoid importing the smartaccount package.
type SessionPolicyRecommendation struct {
	MaxSpendLimit    *big.Int      `json:"maxSpendLimit"`
	MaxDuration      time.Duration `json:"maxDuration"`
	RequireApproval  bool          `json:"requireApproval"`
	AllowedFunctions []string      `json:"allowedFunctions,omitempty"`
}

// PolicyAdapter converts risk assessments into session policy recommendations.
type PolicyAdapter struct {
	engine       *Engine
	fullBudget   *big.Int
	highTrustDur time.Duration
	medTrustDur  time.Duration
	lowTrustDur  time.Duration
}

// NewPolicyAdapter creates a risk-to-policy adapter.
func NewPolicyAdapter(engine *Engine, fullBudget *big.Int) *PolicyAdapter {
	return &PolicyAdapter{
		engine:       engine,
		fullBudget:   fullBudget,
		highTrustDur: 24 * time.Hour,
		medTrustDur:  6 * time.Hour,
		lowTrustDur:  1 * time.Hour,
	}
}

// Recommend generates a policy recommendation based on peer risk.
func (a *PolicyAdapter) Recommend(ctx context.Context, peerDID string, amount *big.Int) (*SessionPolicyRecommendation, error) {
	assessment, err := a.engine.Assess(ctx, peerDID, amount, VerifiabilityMedium)
	if err != nil {
		return nil, err
	}

	rec := &SessionPolicyRecommendation{}

	// Map risk level to spending limits.
	switch assessment.RiskLevel {
	case RiskLow:
		rec.MaxSpendLimit = new(big.Int).Set(a.fullBudget)
		rec.MaxDuration = a.highTrustDur
		rec.RequireApproval = false
	case RiskMedium:
		rec.MaxSpendLimit = new(big.Int).Div(a.fullBudget, big.NewInt(2))
		rec.MaxDuration = a.medTrustDur
		rec.RequireApproval = false
	case RiskHigh:
		rec.MaxSpendLimit = new(big.Int).Div(a.fullBudget, big.NewInt(10))
		rec.MaxDuration = a.lowTrustDur
		rec.RequireApproval = true
	case RiskCritical:
		rec.MaxSpendLimit = new(big.Int)
		rec.MaxDuration = 0
		rec.RequireApproval = true
	}

	return rec, nil
}

// AdaptToRiskPolicyFunc returns a function compatible with the policy engine callback type.
func (a *PolicyAdapter) AdaptToRiskPolicyFunc() func(ctx context.Context, peerDID string) (*SessionPolicyRecommendation, error) {
	return func(ctx context.Context, peerDID string) (*SessionPolicyRecommendation, error) {
		return a.Recommend(ctx, peerDID, a.fullBudget)
	}
}
