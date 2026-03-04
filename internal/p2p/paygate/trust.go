// Package paygate implements trust-based payment tier routing.
package paygate

import "context"

// DefaultPostPayThreshold is the minimum trust score for a peer to qualify for
// post-pay (pay-after-execution). This constant is shared between paygate and
// team payment modules to avoid threshold drift.
const DefaultPostPayThreshold = 0.7

// ReputationFunc returns the trust score for a peer. The score is in [0, 1].
type ReputationFunc func(ctx context.Context, peerDID string) (float64, error)

// TrustConfig holds thresholds for trust-based payment tier decisions.
type TrustConfig struct {
	// PostPayMinScore is the minimum score to qualify for post-pay.
	PostPayMinScore float64
}

// DefaultTrustConfig returns a TrustConfig with production defaults.
func DefaultTrustConfig() TrustConfig {
	return TrustConfig{PostPayMinScore: DefaultPostPayThreshold}
}
