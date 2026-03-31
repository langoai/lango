## MODIFIED Requirements

### Requirement: OntologyBridge implementation
The system SHALL provide an `OntologyBridge` struct in `internal/p2p/ontologybridge/` that implements `protocol.OntologyHandler`. The bridge SHALL use a `TrustScorer` interface (not concrete `*reputation.Store`) for trust checks. The bridge SHALL provide `SetReputation(TrustScorer)` and `SetEventBus(*eventbus.Bus)` setters. The bridge SHALL publish `SchemaExchangeEvent` after successful schema query (Direction="export") and schema propose (Direction="import") operations with PeerDID, TypeCount, PredCount, and ImportMode fields populated.

#### Scenario: Event published on successful export
- **WHEN** HandleSchemaQuery succeeds
- **THEN** a SchemaExchangeEvent with Direction="export", TypeCount, PredCount SHALL be published

#### Scenario: Event published on successful import
- **WHEN** HandleSchemaPropose imports types/predicates
- **THEN** a SchemaExchangeEvent with Direction="import", TypeCount, PredCount, ImportMode SHALL be published

#### Scenario: Trust rejection with TrustScorer
- **WHEN** a peer's trust score is below MinTrustForSchema
- **THEN** HandleSchemaQuery SHALL return an error

### Requirement: Peer principal context
The bridge SHALL set `peer:<did>` as the principal via `ctxkeys.WithPrincipal(ctx, "peer:"+peerDID)` before calling OntologyService methods.

#### Scenario: Principal set for audit
- **WHEN** the bridge calls ExportSchema or ImportSchema
- **THEN** the context SHALL carry principal `peer:<peerDID>`

### Requirement: Post-build wiring
The bridge SHALL be connected to the P2P protocol handler via post-build wiring in `app.go`. The bridge reference SHALL be passed through `intelligenceValues` to avoid module dependency changes.

#### Scenario: Handler connected at runtime
- **WHEN** both P2P and Ontology Exchange are enabled
- **THEN** the P2P handler SHALL have OntologyHandler set (schema_query requests SHALL NOT return "ontology handler not configured")
