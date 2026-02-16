## ADDED Requirements

### Requirement: BoltDB Store Initialization
`NewBoltStore(path)` SHALL expand tilde prefixes and ensure parent directories exist before opening the BoltDB database.

The function SHALL:
1. If `path` starts with `~/`, replace the prefix with the user's home directory obtained from `os.UserHomeDir()`
2. Create parent directories using `os.MkdirAll(filepath.Dir(path), 0o700)` if they do not exist
3. Open the BoltDB file at the resolved path

#### Scenario: Path with tilde prefix
- **WHEN** `NewBoltStore("~/.lango/graph.db")` is called
- **THEN** the tilde SHALL be expanded to the user's home directory and the database SHALL be opened at the absolute path

#### Scenario: Parent directory does not exist
- **WHEN** `NewBoltStore` is called with a path whose parent directory does not exist
- **THEN** the parent directory SHALL be created with `0o700` permissions before opening the database

#### Scenario: Home directory resolution failure
- **WHEN** `os.UserHomeDir()` returns an error during tilde expansion
- **THEN** `NewBoltStore` SHALL return a wrapped error without attempting to open the database

#### Scenario: Directory creation failure
- **WHEN** `os.MkdirAll` returns an error
- **THEN** `NewBoltStore` SHALL return a wrapped error without attempting to open the database

### Requirement: BoltDB-backed triple store
The system SHALL provide a graph store backed by BoltDB that stores Subject-Predicate-Object triples with SPO, POS, and OSP index buckets for efficient lookups.

#### Scenario: Add and query a triple
- **WHEN** a triple `{Subject: "error:timeout", Predicate: "caused_by", Object: "tool:exec"}` is added
- **THEN** the triple SHALL be retrievable by subject, by object, and by subject+predicate queries

#### Scenario: Batch add triples atomically
- **WHEN** multiple triples are submitted via `AddTriples`
- **THEN** all triples SHALL be persisted in a single BoltDB transaction

#### Scenario: BFS traversal
- **WHEN** `Traverse("node-A", maxDepth=2, predicates=["related_to"])` is called
- **THEN** the store SHALL return all triples reachable within 2 hops following only `related_to` edges

### Requirement: Predicate vocabulary
The graph store SHALL define these standard predicates: `related_to`, `caused_by`, `resolved_by`, `follows`, `similar_to`, `contains`, `in_session`, `reflects_on`, `learned_from`.

#### Scenario: Invalid predicate rejected by extractor
- **WHEN** the entity extractor produces a triple with predicate `"foo_bar"`
- **THEN** the triple SHALL be skipped during parsing

### Requirement: Async graph buffer
The system SHALL provide a `GraphBuffer` that batches graph update requests and processes them on a background goroutine with Start/Enqueue/Stop lifecycle.

#### Scenario: GraphBuffer lifecycle
- **WHEN** the application starts with `graph.enabled: true`
- **THEN** GraphBuffer.Start() SHALL be called, and on shutdown GraphBuffer.Stop() SHALL drain remaining items before exiting

#### Scenario: Non-blocking enqueue
- **WHEN** the buffer queue is full (capacity 256)
- **THEN** Enqueue SHALL drop the request without blocking

### Requirement: Graph store close on shutdown
The system SHALL call `GraphStore.Close()` during application shutdown after the WaitGroup completes.

#### Scenario: Clean shutdown
- **WHEN** the application is stopping
- **THEN** GraphBuffer.Stop() SHALL be called before wg.Wait(), and GraphStore.Close() SHALL be called after wg.Wait()

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
