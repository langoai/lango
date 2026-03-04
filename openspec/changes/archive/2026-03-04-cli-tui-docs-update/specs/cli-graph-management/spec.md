## MODIFIED Requirements

### Requirement: AllTriples method on Store interface
The graph.Store interface SHALL include an `AllTriples(ctx context.Context) ([]Triple, error)` method that returns every triple in the store. This method is required to support the graph export command.

#### Scenario: AllTriples on populated store
- **WHEN** AllTriples() is called on a store containing N triples
- **THEN** the method returns a slice of exactly N Triple values with no error

#### Scenario: AllTriples on empty store
- **WHEN** AllTriples() is called on an empty store
- **THEN** the method returns an empty slice with no error

### Requirement: BoltDB AllTriples implementation
The BoltDB-backed Store implementation SHALL implement AllTriples() by scanning the SPO index bucket and returning all triples.

#### Scenario: Full scan
- **WHEN** AllTriples() is called on a BoltDB store with triples
- **THEN** the implementation iterates the SPO bucket, decodes all entries, and returns the complete list

### Requirement: Backward compatibility
The addition of AllTriples() to the Store interface SHALL NOT change the behavior of any existing Store methods. All existing tests for QueryBySubject, QueryByObject, QueryBySubjectPredicate, Count, PredicateStats, and ClearAll SHALL continue to pass.

#### Scenario: Existing tests pass
- **WHEN** `go test ./internal/graph/...` is run after the interface addition
- **THEN** all existing tests pass without modification
