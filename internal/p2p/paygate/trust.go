// Package paygate implements trust-based payment tier routing.
package paygate

import "context"

// ReputationFunc returns the trust score for a peer. The score is in [0, 1].
type ReputationFunc func(ctx context.Context, peerDID string) (float64, error)

// TrustConfig holds thresholds for trust-based payment tier decisions.
type TrustConfig struct {
	// PostPayMinScore is the minimum score to qualify for post-pay (default: 0.8).
	PostPayMinScore float64
}

// DefaultTrustConfig returns a TrustConfig with production defaults.
func DefaultTrustConfig() TrustConfig {
	return TrustConfig{PostPayMinScore: 0.8}
}
