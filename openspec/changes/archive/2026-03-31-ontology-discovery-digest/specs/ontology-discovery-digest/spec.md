# Spec: Ontology Discovery Digest

## Status: ADDED

## Requirements

### REQ-ODD-1: OntologyDigest Type

The `discovery` package MUST define an `OntologyDigest` struct with the following fields:

| Field | Type | JSON Tag | Required |
|-------|------|----------|----------|
| `SchemaVersion` | `int` | `"schemaVersion"` | Yes |
| `Digest` | `string` | `"digest"` | Yes |
| `TypeCount` | `int` | `"typeCount"` | Yes |
| `PredicateCount` | `int` | `"predicateCount"` | Yes |
| `TypeNames` | `[]string` | `"typeNames,omitempty"` | No |

### REQ-ODD-2: GossipCard Field

`GossipCard` MUST include an optional `OntologyDigest` field:
- Type: `*OntologyDigest`
- JSON tag: `"ontologyDigest,omitempty"`
- When nil, the field MUST be omitted from JSON serialization.

### REQ-ODD-3: AgentAd Field

`AgentAd` MUST include an optional `OntologyDigest` field:
- Type: `*OntologyDigest`
- JSON tag: `"ontologyDigest,omitempty"`
- When nil, the field MUST be omitted from JSON serialization.

### REQ-ODD-4: Backward Compatibility

- Peers that do not recognize the `ontologyDigest` field MUST be able to deserialize cards/ads without error.
- Cards/ads without the field MUST be accepted without error by peers that do recognize it.

### REQ-ODD-5: TypeNames Privacy

- `TypeNames` MUST be omitted from JSON when empty (via `omitempty`).
- Future `AdvertiseTypeNames` config flag (default `false`) will control population. This change does NOT implement the config flag.
