## ADDED Requirements

### Requirement: Shared predicate validity source
The runtime SHALL use the ontology service predicate validator closure as the primary predicate-validity source for observe-only admission decisions and graph-store validation.

#### Scenario: Graph admission and graph-store validation use the same validator closure
- **WHEN** observe-only admission and graph-store validation both perform predicate checks
- **THEN** the runtime SHALL obtain those checks from the same ontology service predicate validator closure
- **AND** observe-only admission SHALL identify that closure with the stable validator-source value `ontology_registry`

#### Scenario: Ontology init failure preserves existing graph validation behavior
- **WHEN** ontology is disabled or ontology initialization fails
- **THEN** graph-store validation SHALL continue to use the built-in hardcoded graph predicate validator
- **AND** observe-only admission SHALL switch to the stable validator-source value `unavailable` rather than blocking current graph writes
