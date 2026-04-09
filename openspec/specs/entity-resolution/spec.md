## Purpose

Capability spec for entity-resolution. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Entity alias resolution
The system SHALL provide a `Resolve(rawID)` method that returns the canonical ID for a given raw ID. If no alias exists, the raw ID itself SHALL be returned unchanged.

#### Scenario: Resolve with no alias
- **WHEN** `Resolve("error:timeout")` is called with no aliases registered
- **THEN** `"error:timeout"` is returned

#### Scenario: Resolve with alias
- **GIVEN** an alias `error:api_timeout → error:timeout` is registered
- **WHEN** `Resolve("error:api_timeout")` is called
- **THEN** `"error:timeout"` is returned

### Requirement: DeclareSameAs
`DeclareSameAs(nodeA, nodeB, source)` SHALL declare that two node IDs refer to the same entity and register an alias. The second argument is treated as canonical on tie.

### Requirement: Entity merge
`MergeEntities(canonical, duplicate)` SHALL:
1. Snapshot all outgoing and incoming triples of the duplicate
2. Replicate them with canonical IDs (subject or object replaced)
3. Retract original triples via `RetractFact` (soft delete with ValidTo)
4. Register alias (duplicate → canonical) as the LAST step

#### Scenario: Merge moves triples
- **GIVEN** `dup:x` has outgoing and incoming triples
- **WHEN** `MergeEntities("canon:x", "dup:x")` is called
- **THEN** canonical has the replicated triples, original triples have ValidTo set

### Requirement: Entity split
`SplitEntity(canonical, splitOut)` SHALL remove the alias for `splitOut`. Relationship restoration is manual.

### Requirement: Write-path canonicalization
`StoreTriple` SHALL resolve subject and object through the entity resolver before storing. This ensures all triples are stored under canonical IDs.

### Requirement: Read-path canonicalization
`QueryTriples(subject)` SHALL resolve the subject through the entity resolver before querying the graph store. This ensures alias-based queries return canonical results.
