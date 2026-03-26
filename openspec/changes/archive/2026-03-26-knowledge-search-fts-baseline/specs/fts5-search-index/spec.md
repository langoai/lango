## ADDED Requirements

### Requirement: FTS5 runtime probe
The system SHALL provide a `ProbeFTS5(db)` function that tests whether the current SQLite connection supports FTS5. The probe SHALL execute a temporary FTS5 table creation and immediately drop it. The function SHALL return a boolean and SHALL NOT panic or log errors on unsupported builds.

#### Scenario: FTS5 available
- **WHEN** `ProbeFTS5` is called on a SQLite build with FTS5 enabled
- **THEN** it SHALL return `true`

#### Scenario: FTS5 unavailable
- **WHEN** `ProbeFTS5` is called on a SQLite build without FTS5
- **THEN** it SHALL return `false` without error or panic

#### Scenario: Probe is idempotent
- **WHEN** `ProbeFTS5` is called multiple times on the same connection
- **THEN** each call SHALL return the same result and leave no residual tables

### Requirement: FTS5Index table lifecycle
The system SHALL provide an `FTS5Index` type in `internal/search/` that manages an FTS5 virtual table. The constructor SHALL accept a `*sql.DB`, a table name, and a list of column names. The `EnsureTable()` method SHALL create the FTS5 virtual table if it does not exist, using `unicode61` tokenizer.

#### Scenario: Create FTS5 table
- **WHEN** `EnsureTable()` is called and the table does not exist
- **THEN** the system SHALL execute `CREATE VIRTUAL TABLE IF NOT EXISTS <name> USING fts5(<columns>, source_id UNINDEXED, tokenize='unicode61')` where `source_id` is an UNINDEXED column for row identification

#### Scenario: Table already exists
- **WHEN** `EnsureTable()` is called and the table already exists
- **THEN** the system SHALL succeed without error (idempotent)

#### Scenario: Drop table
- **WHEN** `DropTable()` is called
- **THEN** the system SHALL execute `DROP TABLE IF EXISTS <name>` and succeed even if the table does not exist

### Requirement: FTS5Index CRUD operations
The `FTS5Index` SHALL provide `Insert`, `Update`, `Delete`, and `BulkInsert` methods. Each record is identified by a `rowid` (string, maps to the source entity's key). The `FTS5Index` SHALL NOT import or reference any domain types — it operates on raw string columns.

#### Scenario: Insert a record
- **WHEN** `Insert(ctx, rowid, columns)` is called with column values
- **THEN** the system SHALL insert a new FTS5 row with the given rowid and column values

#### Scenario: Update a record
- **WHEN** `Update(ctx, rowid, columns)` is called
- **THEN** the system SHALL delete the old row by rowid and insert the new one (delete+insert pattern for FTS5)

#### Scenario: Delete a record
- **WHEN** `Delete(ctx, rowid)` is called
- **THEN** the system SHALL remove the FTS5 row matching the rowid

#### Scenario: Bulk insert records
- **WHEN** `BulkInsert(ctx, records)` is called with a slice of records
- **THEN** the system SHALL insert all records within a single transaction for performance

#### Scenario: No domain type dependency
- **WHEN** `FTS5Index` is compiled
- **THEN** it SHALL NOT import packages from `internal/knowledge/`, `internal/ent/`, or any domain package

### Requirement: FTS5Index search with BM25 ranking
The `FTS5Index` SHALL provide a `Search(ctx, query, limit)` method that returns results ranked by FTS5 BM25 relevance. The method SHALL support phrase queries (quoted strings), prefix queries (trailing `*`), and plain keyword queries.

#### Scenario: Keyword search
- **WHEN** `Search(ctx, "deploy server", 10)` is called
- **THEN** the system SHALL execute an FTS5 MATCH query and return up to 10 results ordered by BM25 rank (most relevant first)

#### Scenario: Phrase search
- **WHEN** `Search(ctx, "\"deploy server\"", 10)` is called
- **THEN** the system SHALL match the exact phrase "deploy server" in sequence

#### Scenario: Prefix search
- **WHEN** `Search(ctx, "dep*", 10)` is called
- **THEN** the system SHALL match terms starting with "dep" (e.g., "deploy", "dependency")

#### Scenario: Empty query
- **WHEN** `Search(ctx, "", 10)` is called
- **THEN** the system SHALL return an empty result slice without error

#### Scenario: Search result structure
- **WHEN** a search returns results
- **THEN** each result SHALL contain `RowID` (string) and `Rank` (float64, BM25 score)

### Requirement: FTS5Index is domain-agnostic
The `FTS5Index` SHALL NOT contain any knowledge-specific, learning-specific, or temporal logic. It SHALL operate purely on table names, column names, rowids, and string values. Domain semantics (e.g., which entries to index, when to delete) are the caller's responsibility.

#### Scenario: No is_latest awareness
- **WHEN** the FTS5Index is used for knowledge search
- **THEN** the index itself SHALL NOT filter by `is_latest` or any domain predicate — the caller controls what is indexed via Insert/Delete calls

#### Scenario: Reusable across collections
- **WHEN** separate FTS5Index instances are created for knowledge and learning
- **THEN** each instance SHALL operate independently with its own table name and column configuration
