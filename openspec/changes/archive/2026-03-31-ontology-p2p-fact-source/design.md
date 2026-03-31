# Design: Trust-Weighted P2P Fact Source

## Confidence Damping

P2P facts are inherently less reliable than local knowledge. The effective confidence is:

```
effectiveConf = min(PeerTrust, ClaimedConfidence) * P2PConfidenceScale
```

Where `P2PConfidenceScale = 0.8` (80% ceiling). This ensures:
- A peer with trust=0.5 claiming confidence=0.9 gets effective=0.4
- A peer with trust=0.9 claiming confidence=0.9 gets effective=0.72
- No P2P fact can exceed 80% confidence

## Metadata

P2P facts carry additional metadata:
- `_source`: `"p2p_exchange"` — lowest SourcePrecedence (1)
- `_recorded_by`: PeerDID — identifies the asserting peer
- `_p2p_verified`: `"false"` initially, `"true"` after admin verification

## Verification

`VerifyP2PFact` requires `PermAdmin` and flips `_p2p_verified` from `"false"` to `"true"`.
This is an idempotent operation — verifying an already-verified fact is a no-op.

## Tool Filtering

Query tools (`ontology_facts_at`, `ontology_get_entity`, `ontology_query_entities`) gain an
`exclude_unverified` boolean parameter (default: true). When true, triples with
`_p2p_verified="false"` are excluded from results.

## File Layout

- `types.go`: Add `"p2p_exchange": 1` to SourcePrecedence
- `p2p_source.go`: P2PFactInput, P2PConfidenceScale, assertP2PFact, filterVerifiedTriples
- `service.go`: AssertP2PFact, VerifyP2PFact interface + impl
- `tools.go`: exclude_unverified parameter on query tools
