## MODIFIED Requirements

### Requirement: Triple represents a Subject-Predicate-Object relationship
The Triple struct SHALL include Subject, Predicate, Object, SubjectType, ObjectType, and Metadata fields. SubjectType and ObjectType SHALL default to empty string when not provided.

#### Scenario: Triple struct backward compatibility
- **WHEN** existing code creates `graph.Triple{Subject: "a", Predicate: "rel", Object: "b"}` without type fields
- **THEN** the code compiles and SubjectType/ObjectType are empty strings

### Requirement: BoltStore stores and retrieves triples with type metadata
BoltStore SHALL persist SubjectType/ObjectType as `_subject_type`/`_object_type` in the Metadata map during putTriple. On retrieval (tripleFromSPOKey, tripleFromOSPKey), BoltStore SHALL restore SubjectType/ObjectType from metadata keys.

#### Scenario: Round-trip typed triple through BoltDB
- **WHEN** `AddTriple(ctx, Triple{Subject: "x", Predicate: "rel", Object: "y", SubjectType: "A", ObjectType: "B"})` is called
- **AND** `QueryBySubject(ctx, "x")` is called
- **THEN** the returned triple has SubjectType "A" and ObjectType "B"
