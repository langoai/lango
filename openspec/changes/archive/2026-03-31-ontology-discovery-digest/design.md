# Design: Ontology Discovery Digest

## Data Model

```go
type OntologyDigest struct {
    SchemaVersion  int      `json:"schemaVersion"`
    Digest         string   `json:"digest"`
    TypeCount      int      `json:"typeCount"`
    PredicateCount int      `json:"predicateCount"`
    TypeNames      []string `json:"typeNames,omitempty"`
}
```

### Field Semantics

| Field | Purpose |
|-------|---------|
| `SchemaVersion` | Monotonic version counter for the ontology schema |
| `Digest` | Cryptographic hash (SHA-256) of the serialized schema |
| `TypeCount` | Number of entity types defined |
| `PredicateCount` | Number of predicates/relations defined |
| `TypeNames` | Optional list of type names for discoverability (privacy-sensitive) |

## Integration Points

- `GossipCard.OntologyDigest *OntologyDigest` — optional, `json:"ontologyDigest,omitempty"`
- `AgentAd.OntologyDigest *OntologyDigest` — optional, `json:"ontologyDigest,omitempty"`

## Backward Compatibility

Both fields are pointer types with `omitempty`. Old peers that do not understand the field will:
1. Ignore it when deserializing (standard JSON behavior)
2. Not include it when serializing (nil pointer + omitempty = absent)

No protocol version bump is required.

## Privacy Considerations

`TypeNames` is optional and controlled by a future `AdvertiseTypeNames` configuration flag (default: `false`). When disabled, only counts and the opaque digest are shared, preventing schema structure leakage.

## Placement

The `OntologyDigest` type is defined in `internal/p2p/discovery/` alongside `GossipCard` since it is a discovery-layer concept. The ontology subsystem will populate it via a setter or constructor parameter — that wiring is out of scope for this change.
