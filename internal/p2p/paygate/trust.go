// Package paygate implements trust-based payment tier routing.
package paygate

import (
	"github.com/langoai/lango/internal/p2p/trustpolicy"
	"github.com/langoai/lango/internal/types"
)

// DefaultPostPayThreshold is the minimum trust score for a peer to qualify for
// post-pay (pay-after-execution). This constant is shared with trustpolicy to
// avoid threshold drift across payment layers.
const DefaultPostPayThreshold = trustpolicy.DefaultPostPayThreshold

// ReputationFunc is an alias for types.ReputationQuerier for backward
// compatibility within the paygate package.
type ReputationFunc = types.ReputationQuerier

// TrustConfig holds thresholds for trust-based payment tier decisions.
type TrustConfig struct {
	// PostPayMinScore is the minimum score to qualify for post-pay.
	PostPayMinScore float64
}

// DefaultTrustConfig returns a TrustConfig with production defaults.
func DefaultTrustConfig() TrustConfig {
	return TrustConfig{PostPayMinScore: DefaultPostPayThreshold}
}
