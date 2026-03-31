## ADDED Requirements

### Requirement: OntologyBridge implementation
The system SHALL provide an `OntologyBridge` struct in `internal/p2p/ontologybridge/` that implements `protocol.OntologyHandler`. The bridge SHALL use a `TrustScorer` interface (not concrete `*reputation.Store`) for trust checks. The bridge SHALL provide `SetReputation(TrustScorer)` and `SetEventBus(*eventbus.Bus)` setters. The bridge SHALL publish `SchemaExchangeEvent` after successful operations with PeerDID, TypeCount, PredCount, and ImportMode fields populated.

#### Scenario: Schema query returns exported bundle
- **WHEN** `HandleSchemaQuery` is called with a peer DID that has trust >= MinTrustForSchema
- **THEN** it SHALL return a SchemaQueryResponse containing the exported SchemaBundle JSON

#### Scenario: Schema query rejected for low trust
- **WHEN** `HandleSchemaQuery` is called with a peer DID that has trust < MinTrustForSchema
- **THEN** it SHALL return an error

#### Scenario: Schema propose imports as shadow
- **WHEN** `HandleSchemaPropose` is called with AutoImportMode="shadow"
- **THEN** it SHALL import the proposed types with shadow status and return the ImportResult

### Requirement: Peer principal context
The bridge SHALL set `peer:<did>` as the principal via `ctxkeys.WithPrincipal(ctx, "peer:"+peerDID)` before calling OntologyService methods.

#### Scenario: Principal set for audit
- **WHEN** the bridge calls ExportSchema or ImportSchema
- **THEN** the context SHALL carry principal `peer:<peerDID>`

### Requirement: OntologyExchangeConfig
The system SHALL support config `ontology.exchange` with fields: Enabled (bool), MinTrustForSchema (float64, default 0.5), MinTrustForFacts (float64, default 0.7), AutoImportMode (string, default "shadow"), MaxTypesPerImport (int, default 10).

#### Scenario: Exchange disabled
- **WHEN** `ontology.exchange.enabled` is false
- **THEN** no OntologyBridge SHALL be created

### Requirement: SchemaExchangeEvent
The system SHALL publish a `SchemaExchangeEvent` to the event bus after successful schema exchange operations.

#### Scenario: Event published on export
- **WHEN** a peer successfully queries schema
- **THEN** a SchemaExchangeEvent with Direction="export" SHALL be published
