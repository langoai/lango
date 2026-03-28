## ADDED Requirements

### Requirement: Knowledge store FTS5 injection
The `knowledge.Store` SHALL provide a `SetFTS5Index(idx *search.FTS5Index)` method to inject the FTS5 index for knowledge search. A separate `SetLearningFTS5Index(idx *search.FTS5Index)` SHALL be provided for learning search. When no index is injected, search methods SHALL use the existing LIKE fallback.

#### Scenario: FTS5 index injected
- **WHEN** `SetFTS5Index` is called with a non-nil FTS5Index
- **THEN** subsequent `SearchKnowledge` calls SHALL use the FTS5 path

#### Scenario: No FTS5 index injected
- **WHEN** `SetFTS5Index` is never called (or called with nil)
- **THEN** `SearchKnowledge` SHALL use the existing LIKE-based search unchanged

#### Scenario: Learning FTS5 index injected
- **WHEN** `SetLearningFTS5Index` is called with a non-nil FTS5Index
- **THEN** subsequent `SearchLearnings` calls SHALL use the FTS5 path

### Requirement: FTS5-first search with LIKE fallback for knowledge
When an FTS5 index is available, `SearchKnowledge` SHALL query the FTS5 index first, then resolve full entries from the Ent store by key. Results SHALL be ordered by BM25 rank (from FTS5), not by static `RelevanceScore`. The category filter SHALL still be applied via Ent predicates after FTS5 retrieval.

#### Scenario: FTS5 search path
- **WHEN** `SearchKnowledge` is called and FTS5 index is available
- **THEN** the system SHALL search via FTS5 MATCH, collect matching keys, then load entries from Ent by those keys
- **AND** results SHALL be ordered by FTS5 BM25 rank

#### Scenario: FTS5 search with category filter
- **WHEN** `SearchKnowledge` is called with a category and FTS5 is available
- **THEN** the system SHALL get FTS5 results first, then filter by category when loading from Ent

#### Scenario: LIKE fallback when FTS5 unavailable
- **WHEN** `SearchKnowledge` is called and FTS5 index is nil
- **THEN** the system SHALL use existing per-keyword LIKE predicates with RelevanceScore ordering

#### Scenario: FTS5 search error falls back to LIKE
- **WHEN** `SearchKnowledge` via FTS5 returns an error
- **THEN** the system SHALL log a warning and fall back to the LIKE path for that query

### Requirement: FTS5-first search with LIKE fallback for learning
When an FTS5 index is available, `SearchLearnings` SHALL query the FTS5 index first, then resolve full entries from the Ent store. Results SHALL be ordered by BM25 rank. The category filter SHALL still be applied via Ent predicates.

#### Scenario: FTS5 search path for learning
- **WHEN** `SearchLearnings` is called and learning FTS5 index is available
- **THEN** the system SHALL search via FTS5 MATCH, collect matching IDs, then load entries from Ent

#### Scenario: LIKE fallback for learning
- **WHEN** `SearchLearnings` is called and learning FTS5 index is nil
- **THEN** the system SHALL use existing per-keyword LIKE predicates with Confidence ordering

### Requirement: Write-time FTS5 sync for knowledge
The `knowledge.Store` SHALL update the FTS5 index at write time. The FTS5 `source_id` SHALL always be `key` (not per-version). Only the latest version's content SHALL be indexed. On first save (version 1), the system SHALL insert into FTS5. On subsequent saves (version > 1), the system SHALL update (delete+re-insert) the existing FTS5 entry with the new content. On delete, the system SHALL delete the FTS5 entry. The FTS5 index columns for knowledge SHALL be `key` and `content`.

#### Scenario: New knowledge triggers FTS5 insert
- **WHEN** `SaveKnowledge` creates version 1 of a new entry and FTS5 index is available
- **THEN** the system SHALL call `fts5Index.Insert(ctx, key, [key, content])`

#### Scenario: New version triggers FTS5 update
- **WHEN** `SaveKnowledge` creates version N+1 of an existing entry and FTS5 index is available
- **THEN** the system SHALL call `fts5Index.Update(ctx, key, [key, content])` with the new version's content
- **AND** the FTS5 `source_id` SHALL remain `key` (not change per-version)

#### Scenario: Deleted knowledge triggers FTS5 delete
- **WHEN** `DeleteKnowledge` is called and FTS5 index is available
- **THEN** the system SHALL call `fts5Index.Delete(ctx, key)`

#### Scenario: FTS5 write failure does not block Ent write
- **WHEN** FTS5 insert/update/delete fails
- **THEN** the system SHALL log a warning but SHALL NOT return an error to the caller (Ent write succeeds regardless)

#### Scenario: FTS5 contains only latest content
- **WHEN** a key has 3 versions and FTS5 is queried
- **THEN** FTS5 SHALL contain only the version 3 content (one row per key)

### Requirement: Write-time FTS5 sync for learning
The `knowledge.Store` SHALL update the learning FTS5 index at write time. The FTS5 index columns for learning SHALL be `trigger`, `error_pattern`, and `fix`.

#### Scenario: New learning triggers FTS5 insert
- **WHEN** `SaveLearning` creates a new entry and learning FTS5 index is available
- **THEN** the system SHALL insert a row with the learning's ID as rowid and trigger/error_pattern/fix as columns

#### Scenario: Deleted learning triggers FTS5 delete
- **WHEN** `DeleteLearning` is called and learning FTS5 index is available
- **THEN** the system SHALL delete the FTS5 row by learning ID

### Requirement: FTS5 initial bulk index on startup
During app initialization, after FTS5 tables are created, the system SHALL bulk-index only the latest version (`is_latest=true`) of all knowledge entries into the FTS5 table. Learning bulk index is unchanged.

#### Scenario: Bulk index knowledge on startup (latest only)
- **WHEN** the app starts and FTS5 is available
- **THEN** only knowledge entries with `is_latest=true` SHALL be bulk-inserted into the knowledge FTS5 table
- **AND** the SQL query SHALL filter `WHERE is_latest = 1`

#### Scenario: Bulk index learning on startup
- **WHEN** the app starts and FTS5 is available
- **THEN** all existing learning entries SHALL be bulk-inserted into the learning FTS5 table

#### Scenario: Bulk index is idempotent
- **WHEN** the app restarts and FTS5 tables already contain data
- **THEN** the system SHALL clear and re-index without creating duplicates
