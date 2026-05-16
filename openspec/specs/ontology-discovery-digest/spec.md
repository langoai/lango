# Spec: Ontology Discovery Digest

## Status: ADDED

## Purpose

Capability spec for ontology-discovery-digest. See requirements below for scope and behavior contracts.

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

#### Scenario: OntologyDigest exposes the documented fields
- **WHEN** the discovery package constructs an `OntologyDigest`
- **THEN** it MUST expose the documented schema, digest, counts, and optional type names fields

### REQ-ODD-2: GossipCard Field

`GossipCard` MUST include an optional `OntologyDigest` field:
- Type: `*OntologyDigest`
- JSON tag: `"ontologyDigest,omitempty"`
- When nil, the field MUST be omitted from JSON serialization.

#### Scenario: GossipCard omits ontologyDigest when nil
- **WHEN** a `GossipCard` has no ontology digest attached
- **THEN** JSON serialization MUST omit the `ontologyDigest` field

### REQ-ODD-3: AgentAd Field

`AgentAd` MUST include an optional `OntologyDigest` field:
- Type: `*OntologyDigest`
- JSON tag: `"ontologyDigest,omitempty"`
- When nil, the field MUST be omitted from JSON serialization.

#### Scenario: AgentAd omits ontologyDigest when nil
- **WHEN** an `AgentAd` has no ontology digest attached
- **THEN** JSON serialization MUST omit the `ontologyDigest` field

### REQ-ODD-4: Backward Compatibility

- Peers that do not recognize the `ontologyDigest` field MUST be able to deserialize cards/ads without error.
- Cards/ads without the field MUST be accepted without error by peers that do recognize it.

#### Scenario: Unknown ontologyDigest field does not break peers
- **WHEN** a peer receives a card or ad containing an `ontologyDigest` field it does not recognize
- **THEN** deserialization MUST continue without error

#### Scenario: Missing ontologyDigest remains acceptable
- **WHEN** a peer that understands ontology digests receives a card or ad without that field
- **THEN** it MUST accept the payload without error

### REQ-ODD-5: TypeNames Privacy

- `TypeNames` MUST be omitted from JSON when empty (via `omitempty`).
- Future `AdvertiseTypeNames` config flag (default `false`) will control population. This change does NOT implement the config flag.

#### Scenario: Empty TypeNames are omitted from JSON
- **WHEN** `TypeNames` is empty on an `OntologyDigest`
- **THEN** JSON serialization MUST omit the `typeNames` field
