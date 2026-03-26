## MODIFIED Requirements

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
