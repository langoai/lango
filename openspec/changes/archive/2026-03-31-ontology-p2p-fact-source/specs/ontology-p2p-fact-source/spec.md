# Spec: Trust-Weighted P2P Fact Source

## Requirements

### REQ-1: Source Precedence
- `"p2p_exchange"` MUST be added to `SourcePrecedence` with value `1` (lowest)
- Auto-resolution MUST favor all other sources over P2P facts

### REQ-2: AssertP2PFact
- Effective confidence = `min(PeerTrust, Confidence) * P2PConfidenceScale`
- `P2PConfidenceScale` = 0.8
- Metadata `_source` = `"p2p_exchange"`
- Metadata `_recorded_by` = PeerDID
- Metadata `_p2p_verified` = `"false"`
- Requires `PermWrite`

### REQ-3: VerifyP2PFact
- Flips `_p2p_verified` from `"false"` to `"true"`
- Requires `PermAdmin`
- No-op if already verified or not a P2P fact

### REQ-4: Query Filtering
- `ontology_facts_at`, `ontology_get_entity`, `ontology_query_entities` accept `exclude_unverified` (default: true)
- When true, triples with `_p2p_verified="false"` are excluded

## Interfaces

```go
// Added to OntologyService
AssertP2PFact(ctx context.Context, input P2PFactInput) (*AssertionResult, error)
VerifyP2PFact(ctx context.Context, subject, predicate, object string) error
```
