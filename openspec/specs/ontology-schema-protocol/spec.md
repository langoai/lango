## Purpose

Capability spec for ontology-schema-protocol. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Schema query request type
The protocol package SHALL define a `RequestSchemaQuery` constant of type `RequestType` with value `"schema_query"`.

#### Scenario: Schema query constant exists
- **WHEN** the protocol package is compiled
- **THEN** `RequestSchemaQuery` SHALL equal `"schema_query"` and be of type `RequestType`

### Requirement: Schema propose request type
The protocol package SHALL define a `RequestSchemaPropose` constant of type `RequestType` with value `"schema_propose"`.

#### Scenario: Schema propose constant exists
- **WHEN** the protocol package is compiled
- **THEN** `RequestSchemaPropose` SHALL equal `"schema_propose"` and be of type `RequestType`

### Requirement: Schema query request/response structs
The protocol package SHALL define `SchemaQueryRequest` with fields `RequestedTypes []string` and `IncludePredicates bool`, and `SchemaQueryResponse` with field `Bundle json.RawMessage`.

#### Scenario: Schema query round-trip serialization
- **WHEN** a `SchemaQueryRequest` is JSON-marshaled and unmarshaled
- **THEN** all fields SHALL be preserved including empty `RequestedTypes`

### Requirement: Schema propose request/response structs
The protocol package SHALL define `SchemaProposeRequest` with fields `Bundle json.RawMessage` and `Reason string`, and `SchemaProposeResponse` with fields `Action string`, `Accepted []string`, `Rejected []string`, and `Result json.RawMessage`.

#### Scenario: Schema propose round-trip serialization
- **WHEN** a `SchemaProposeRequest` is JSON-marshaled and unmarshaled
- **THEN** all fields SHALL be preserved including the raw bundle bytes

### Requirement: OntologyHandler interface
The protocol package SHALL define an `OntologyHandler` interface with methods `HandleSchemaQuery(ctx context.Context, peerDID string, req SchemaQueryRequest) (*SchemaQueryResponse, error)` and `HandleSchemaPropose(ctx context.Context, peerDID string, req SchemaProposeRequest) (*SchemaProposeResponse, error)`.

#### Scenario: Interface is implementable
- **WHEN** an external package implements both methods of `OntologyHandler`
- **THEN** it SHALL satisfy the interface at compile time

### Requirement: Handler ontology dispatch
The `Handler` struct SHALL accept an `OntologyHandler` via `SetOntologyHandler` and dispatch `schema_query` and `schema_propose` requests to it in `handleRequest`.

#### Scenario: Schema query dispatch
- **WHEN** a request with type `schema_query` arrives and `OntologyHandler` is set
- **THEN** the handler SHALL decode the payload into `SchemaQueryRequest`, call `HandleSchemaQuery`, and return the response as JSON

#### Scenario: Schema propose dispatch
- **WHEN** a request with type `schema_propose` arrives and `OntologyHandler` is set
- **THEN** the handler SHALL decode the payload into `SchemaProposeRequest`, call `HandleSchemaPropose`, and return the response as JSON

#### Scenario: Ontology handler not configured
- **WHEN** a `schema_query` or `schema_propose` request arrives and `OntologyHandler` is nil
- **THEN** the handler SHALL return a `ResponseStatusError` with error message `"ontology handler not configured"`
