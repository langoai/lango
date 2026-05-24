# Spec: Trust-Weighted P2P Fact Source

## Purpose

Capability spec for ontology-p2p-fact-source. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: Source Precedence
- `"p2p_exchange"` MUST be added to `SourcePrecedence` with value `1` (lowest)
- Auto-resolution MUST favor all other sources over P2P facts

#### Scenario: P2P facts have the lowest precedence
- **WHEN** a P2P-exchanged fact conflicts with a fact from any other source
- **THEN** auto-resolution MUST favor the non-P2P source

### REQ-2: AssertP2PFact
`AssertP2PFact` SHALL compute effective confidence as `min(PeerTrust, Confidence) * P2PConfidenceScale`, with `P2PConfidenceScale = 0.8`. It SHALL store metadata `_source="p2p_exchange"`, `_recorded_by=PeerDID`, `_p2p_verified="false"`, and SHALL require `PermWrite`.

#### Scenario: AssertP2PFact stores scaled confidence and metadata
- **WHEN** `AssertP2PFact` is called with peer trust, confidence, and a peer DID
- **THEN** the stored fact MUST use the scaled effective confidence
- **AND** attach the documented `_source`, `_recorded_by`, and `_p2p_verified` metadata

### REQ-3: VerifyP2PFact
`VerifyP2PFact` SHALL flip `_p2p_verified` from `"false"` to `"true"`, SHALL require `PermAdmin`, and SHALL be a no-op if the fact is already verified or is not a P2P fact.

#### Scenario: VerifyP2PFact marks an unverified P2P fact as verified
- **WHEN** an admin verifies a stored P2P fact that is currently unverified
- **THEN** `_p2p_verified` MUST change from `"false"` to `"true"`

### REQ-4: Query Filtering
`ontology_facts_at`, `ontology_get_entity`, and `ontology_query_entities` SHALL accept `exclude_unverified` (default: true). When it is true, triples with `_p2p_verified="false"` SHALL be excluded.

#### Scenario: Query filtering excludes unverified P2P facts by default
- **WHEN** ontology fact or entity queries run without overriding `exclude_unverified`
- **THEN** triples with `_p2p_verified="false"` MUST be excluded from the results

## Interfaces

```go
// Added to OntologyService
AssertP2PFact(ctx context.Context, input P2PFactInput) (*AssertionResult, error)
VerifyP2PFact(ctx context.Context, subject, predicate, object string) error
```
