## ADDED Requirements

### Requirement: Count triples
The Store interface SHALL provide a `Count(ctx) (int, error)` method that returns the total number of triples in the store.

#### Scenario: Count on empty store
- **WHEN** Count is called on an empty store
- **THEN** result is 0 with nil error

#### Scenario: Count on populated store
- **WHEN** Count is called after adding N triples
- **THEN** result is N with nil error

### Requirement: Predicate statistics
The Store interface SHALL provide a `PredicateStats(ctx) (map[string]int, error)` method that returns the count of triples grouped by predicate type.

#### Scenario: Stats on populated store
- **WHEN** PredicateStats is called on a store with triples of different predicate types
- **THEN** result is a map where each key is a predicate and each value is the count of triples with that predicate

### Requirement: Clear all triples
The Store interface SHALL provide a `ClearAll(ctx) error` method that removes all triples from all index buckets atomically.

#### Scenario: Clear and verify
- **WHEN** ClearAll is called on a populated store
- **THEN** Count returns 0 and all queries return empty results
