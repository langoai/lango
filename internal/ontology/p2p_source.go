package ontology

import (
	"context"
	"fmt"
	"math"

	"github.com/langoai/lango/internal/graph"
)

// P2PFactInput carries a fact received from a peer with reputation context.
type P2PFactInput struct {
	Triple     graph.Triple
	PeerDID    string  // DID of the asserting peer
	PeerTrust  float64 // 0.0-1.0 peer reputation score
	Confidence float64 // peer's claimed confidence
}

// P2PConfidenceScale caps effective confidence for P2P facts (max 80%).
const P2PConfidenceScale = 0.8

// assertP2PFact stores a fact received from a peer with damped confidence.
// Effective confidence = min(PeerTrust, Confidence) * P2PConfidenceScale.
func assertP2PFact(ctx context.Context, truth TruthMaintainer, input P2PFactInput) (*AssertionResult, error) {
	effectiveConf := math.Min(input.PeerTrust, input.Confidence) * P2PConfidenceScale

	if input.Triple.Metadata == nil {
		input.Triple.Metadata = make(map[string]string)
	}
	input.Triple.Metadata[MetaSource] = "p2p_exchange"
	input.Triple.Metadata[MetaRecordedBy] = input.PeerDID
	input.Triple.Metadata["_p2p_verified"] = "false"
	input.Triple.Metadata[MetaConfidence] = fmt.Sprintf("%.4f", effectiveConf)

	return truth.AssertFact(ctx, AssertionInput{
		Triple:     input.Triple,
		Source:     "p2p_exchange",
		Confidence: effectiveConf,
	})
}

// filterVerifiedTriples removes P2P facts with _p2p_verified="false" from results.
func filterVerifiedTriples(triples []graph.Triple) []graph.Triple {
	result := make([]graph.Triple, 0, len(triples))
	for _, t := range triples {
		if t.Metadata != nil && t.Metadata["_p2p_verified"] == "false" {
			continue
		}
		result = append(result, t)
	}
	return result
}

// filterVerifiedResultTriples removes unverified P2P facts from ResultTriple slices.
func filterVerifiedResultTriples(triples []ResultTriple) []ResultTriple {
	if len(triples) == 0 {
		return triples
	}
	result := make([]ResultTriple, 0, len(triples))
	for _, t := range triples {
		if t.Metadata != nil && t.Metadata["_p2p_verified"] == "false" {
			continue
		}
		result = append(result, t)
	}
	return result
}
