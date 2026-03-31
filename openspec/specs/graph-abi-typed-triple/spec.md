## ADDED Requirements

### Requirement: Triple struct carries type information
The `graph.Triple` struct SHALL include `SubjectType string` and `ObjectType string` fields. Empty string SHALL mean "untyped" and existing code creating triples without these fields SHALL continue to compile and function normally.

#### Scenario: Create typed triple
- **WHEN** a triple is created with `SubjectType: "ErrorPattern"` and `ObjectType: "Tool"`
- **THEN** the triple carries these type values through storage and retrieval

#### Scenario: Create untyped triple (backward compatibility)
- **WHEN** a triple is created without SubjectType/ObjectType fields (e.g., `graph.Triple{Subject: "a", Predicate: "rel", Object: "b"}`)
- **THEN** SubjectType and ObjectType default to empty string and the triple is stored and retrieved normally

### Requirement: Type information persisted in BoltDB metadata
BoltStore SHALL store SubjectType and ObjectType as `_subject_type` and `_object_type` entries in the triple's Metadata map before persisting to BoltDB. On read, BoltStore SHALL restore SubjectType/ObjectType fields from metadata. Empty type values SHALL NOT be stored in metadata.

#### Scenario: Store and retrieve typed triple
- **WHEN** a triple with `SubjectType: "ErrorPattern"` is stored via `AddTriple`
- **AND** the same triple is retrieved via `QueryBySubject`
- **THEN** the retrieved triple has `SubjectType: "ErrorPattern"` and `Metadata["_subject_type"]: "ErrorPattern"`

#### Scenario: Retrieve legacy untyped triple
- **WHEN** a triple was stored before this change (no `_subject_type` in metadata)
- **AND** it is retrieved via any query method
- **THEN** SubjectType and ObjectType are empty strings (no error, no panic)

### Requirement: Registry-backed predicate validation on BoltStore
BoltStore SHALL accept a `PredicateValidatorFunc` via `SetPredicateValidator`. When set, `putTriple` SHALL validate the predicate before storing. If the predicate is not recognized by the validator, `putTriple` SHALL return an error. When no validator is set, all predicates SHALL be accepted (backward compatible).

#### Scenario: Valid predicate accepted
- **WHEN** a validator is set that recognizes "caused_by"
- **AND** `AddTriple` is called with predicate "caused_by"
- **THEN** the triple is stored successfully

#### Scenario: Invalid predicate rejected
- **WHEN** a validator is set that does not recognize "fake_pred"
- **AND** `AddTriple` is called with predicate "fake_pred"
- **THEN** `AddTriple` returns an error containing "unknown predicate"

#### Scenario: No validator set (backward compatible)
- **WHEN** no validator has been set on BoltStore
- **AND** `AddTriple` is called with any predicate
- **THEN** the triple is stored successfully regardless of predicate value

### Requirement: Extractor uses ontology-backed predicate validation
Extractor SHALL accept an optional `PredicateValidatorFunc` via `WithPredicateValidator` functional option. When set, `isValidPredicate` SHALL delegate to the validator. When not set, it SHALL fall back to the existing hardcoded 9-predicate switch. Rejected predicates SHALL be logged at warn level with predicate name and source context.

#### Scenario: Extractor with ontology validator rejects unknown predicate
- **WHEN** Extractor is created with `WithPredicateValidator(v)` where v only knows "caused_by"
- **AND** the LLM extracts a triple with predicate "invented_rel"
- **THEN** the triple is skipped and a warn log is emitted

#### Scenario: Extractor without validator uses hardcoded fallback
- **WHEN** Extractor is created without `WithPredicateValidator`
- **AND** the LLM extracts a triple with predicate "caused_by"
- **THEN** the triple is accepted (hardcoded list includes "caused_by")

### Requirement: All triple producers set SubjectType and ObjectType
GraphEngine (recordErrorGraph, RecordFix), MemoryGraphHooks (OnObservation, OnReflection), and wiring event handlers SHALL set SubjectType and ObjectType on all triples they create, using the ObjectType names from the ontology seed (ErrorPattern, Tool, Fix, Session, Observation, Reflection).

#### Scenario: GraphEngine recordErrorGraph sets types
- **WHEN** GraphEngine records an error graph for tool "api_call" with pattern "timeout"
- **THEN** the error triple has `SubjectType: "ErrorPattern"` and `ObjectType: "Tool"`

#### Scenario: MemoryGraphHooks OnObservation sets types
- **WHEN** MemoryGraphHooks records an observation triple
- **THEN** the observation triple has `SubjectType: "Observation"` and `ObjectType: "Session"`

### Requirement: EventBus Triple mirror carries type fields
The `eventbus.Triple` struct SHALL include `SubjectType` and `ObjectType` fields. Conversion code between `graph.Triple` and `eventbus.Triple` in wiring SHALL copy these fields.

#### Scenario: Triples extracted event preserves types
- **WHEN** a TriplesExtractedEvent is published with typed triples
- **AND** the subscriber converts them to graph.Triple
- **THEN** SubjectType and ObjectType are preserved in the converted triples
