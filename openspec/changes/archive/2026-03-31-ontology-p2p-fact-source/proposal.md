# Proposal: Trust-Weighted P2P Fact Source

## Problem
The ontology subsystem has no mechanism for ingesting facts received from peer agents via P2P exchange. Facts from peers need trust-weighted confidence damping and verification tracking before they can be treated as reliable knowledge.

## Solution
1. Add `"p2p_exchange"` as the lowest-priority entry in `SourcePrecedence`
2. Implement `AssertP2PFact` — stores peer facts with damped confidence: `effective = min(PeerTrust, Confidence) * 0.8`
3. Implement `VerifyP2PFact` — marks a P2P fact as verified by updating `_p2p_verified` metadata
4. Add `exclude_unverified` parameter to query tools to filter out unverified P2P facts

## Scope
- Only `internal/ontology/` files
- No changes to `internal/p2p/` or `internal/app/`
- New file: `p2p_source.go` for P2P-specific logic
- Interface additions to `OntologyService`
- Tool parameter additions for filtering
