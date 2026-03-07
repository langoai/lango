package risk

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/wallet"
)

// ReputationQuerier queries peer trust scores. Defined locally to avoid import cycles.
type ReputationQuerier func(ctx context.Context, peerDID string) (float64, error)

// Engine implements Assessor using a 3-variable risk matrix:
// trust score x transaction value x output verifiability.
type Engine struct {
	cfg             config.RiskConfig
	reputation      ReputationQuerier
	escrowThreshold *big.Int
	highTrust       float64
	medTrust        float64
}

var _ Assessor = (*Engine)(nil)

// New creates a risk assessment engine.
func New(cfg config.RiskConfig, reputation ReputationQuerier) (*Engine, error) {
	highTrust := cfg.HighTrustScore
	if highTrust == 0 {
		highTrust = 0.8
	}
	medTrust := cfg.MediumTrustScore
	if medTrust == 0 {
		medTrust = 0.5
	}

	threshold, err := wallet.ParseUSDC(cfg.EscrowThreshold)
	if err != nil || threshold.Sign() <= 0 {
		threshold = big.NewInt(5_000_000) // 5 USDC default (6 decimals)
	}

	return &Engine{
		cfg:             cfg,
		reputation:      reputation,
		escrowThreshold: threshold,
		highTrust:       highTrust,
		medTrust:        medTrust,
	}, nil
}

// Assess evaluates risk for a transaction and recommends a strategy.
func (e *Engine) Assess(ctx context.Context, peerDID string, amount *big.Int, v Verifiability) (*Assessment, error) {
	trustScore, err := e.reputation(ctx, peerDID)
	if err != nil {
		return nil, fmt.Errorf("query trust score for %q: %w", peerDID, err)
	}

	factors := computeFactors(trustScore, amount, e.escrowThreshold, v)
	riskScore := computeRiskScore(factors)
	level := classifyRisk(riskScore)
	strategy := e.selectStrategy(trustScore, amount, v)
	explanation := e.explain(trustScore, amount, v, strategy)

	return &Assessment{
		PeerDID:       peerDID,
		Amount:        new(big.Int).Set(amount),
		TrustScore:    trustScore,
		Verifiability: v,
		RiskLevel:     level,
		RiskScore:     riskScore,
		Strategy:      strategy,
		Factors:       factors,
		Explanation:   explanation,
		AssessedAt:    time.Now(),
	}, nil
}

// explain generates a human-readable explanation of the assessment.
func (e *Engine) explain(trust float64, amount *big.Int, v Verifiability, s Strategy) string {
	trustLabel := "low"
	switch {
	case trust >= e.highTrust:
		trustLabel = "high"
	case trust >= e.medTrust:
		trustLabel = "medium"
	}

	valueLabel := "low-value"
	if amount.Cmp(e.escrowThreshold) > 0 {
		valueLabel = "high-value"
	}

	return fmt.Sprintf("peer trust is %s, %s transaction with %s verifiability; recommending %s",
		trustLabel, valueLabel, string(v), string(s))
}

// clamp restricts a value to [0.0, 1.0].
func clamp(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
