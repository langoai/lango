## ADDED Requirements

### Requirement: Per-entity property storage
The system SHALL provide `SetEntityProperty(ctx, entityID, entityType, property, value)` to store property values per entity instance. Properties SHALL be persisted in an EAV table (entity_id, property, value). The entity_id SHALL be canonicalized via EntityResolver before storage.

#### Scenario: Set and get property
- **WHEN** `SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "tool_name", "http_client")` is called
- **THEN** `GetEntityProperties(ctx, "error:timeout")` returns `{"tool_name": "http_client"}`

#### Scenario: Alias-aware storage
- **GIVEN** alias `error:api_timeout → error:timeout`
- **WHEN** `SetEntityProperty(ctx, "error:api_timeout", "ErrorPattern", "tool_name", "http")` is called
- **THEN** the property is stored under canonical ID `error:timeout`

### Requirement: Property schema validation
`SetEntityProperty` SHALL validate that the property name exists in the ObjectType's Properties schema. Unknown property names SHALL be rejected with an error. The entityType parameter SHALL match a registered ObjectType.

#### Scenario: Unknown property rejected
- **WHEN** `SetEntityProperty(ctx, "error:timeout", "ErrorPattern", "nonexistent_prop", "val")` is called
- **THEN** an error is returned indicating the property is not defined in ErrorPattern

#### Scenario: Unknown type rejected
- **WHEN** `SetEntityProperty(ctx, "x", "UnknownType", "prop", "val")` is called
- **THEN** an error is returned indicating the type is not registered

### Requirement: Structured entity query
The system SHALL provide `QueryEntities(ctx, PropertyQuery)` that returns entities matching type + property filters. Filters use AND semantics. Supported FilterOps: `eq` (exact match), `neq` (not equal), `contains` (substring). Results include entity properties and outgoing triples.

#### Scenario: Query by type and property
- **GIVEN** two ErrorPattern entities, one with tool_name="http_client" and one with tool_name="db_client"
- **WHEN** `QueryEntities(ctx, {EntityType: "ErrorPattern", Filters: [{Property: "tool_name", Op: "eq", Value: "http_client"}]})` is called
- **THEN** only the http_client entity is returned

#### Scenario: Query with contains filter
- **GIVEN** entity with tool_name="http_client_v2"
- **WHEN** filter `{Property: "tool_name", Op: "contains", Value: "http"}` is applied
- **THEN** the entity is included in results

### Requirement: Single entity retrieval
`GetEntity(ctx, entityID)` SHALL return an EntityResult containing properties (from PropertyStore), outgoing triples (subject=entityID), and incoming triples (object=entityID). The entityID SHALL be canonicalized via EntityResolver.

#### Scenario: Get entity with alias
- **GIVEN** alias `error:api_timeout → error:timeout` and properties stored under canonical
- **WHEN** `GetEntity(ctx, "error:api_timeout")` is called
- **THEN** properties and triples for `error:timeout` are returned

### Requirement: Property deletion
`DeleteProperties` removes all properties for an entity (used by entity lifecycle management).
