## Purpose

Define the append-only version history for knowledge entries, including schema versioning fields, transactional save semantics, concurrency handling, version history retrieval, domain type extensions, event versioning, and agent tool contract updates.

## Requirements

### Requirement: Knowledge version schema
The Knowledge Ent schema SHALL include a `version` field (int, default 1) and an `is_latest` field (bool, default true). The `key` field SHALL NOT have a UNIQUE constraint. A composite unique index SHALL exist on `(key, version)`. A non-unique index SHALL exist on `(key, is_latest)` for latest-lookup optimization.

#### Scenario: Schema fields present
- **WHEN** the Knowledge Ent schema is defined
- **THEN** it SHALL include `version` (int, default 1) and `is_latest` (bool, default true) fields alongside existing fields

#### Scenario: Composite unique index
- **WHEN** two knowledge entries exist with the same key but different versions
- **THEN** the database SHALL allow both rows
- **AND** attempting to insert a duplicate `(key, version)` pair SHALL fail with a unique constraint error

#### Scenario: Migration from existing data
- **WHEN** the database is migrated from the pre-versioning schema
- **THEN** all existing knowledge rows SHALL have `version=1` and `is_latest=true` (from column defaults)

### Requirement: Append-only SaveKnowledge
`SaveKnowledge` SHALL create a new version row instead of updating in place. When no existing entry exists for the key, it SHALL create version 1 with `is_latest=true`. When an existing latest entry exists, it SHALL atomically (within a transaction) set the old entry's `is_latest` to false and create a new entry with `version = old.version + 1` and `is_latest=true`. The new version SHALL carry forward `use_count` and `relevance_score` from the previous version.

#### Scenario: First save for a key
- **WHEN** `SaveKnowledge` is called with key "user-lang" that does not exist
- **THEN** a row SHALL be created with `version=1`, `is_latest=true`
- **AND** `ContentSavedEvent` SHALL have `IsNew=true`, `NeedsGraph=true`, `Version=1`

#### Scenario: Second save for an existing key
- **WHEN** `SaveKnowledge` is called with key "user-lang" that has version 1
- **THEN** the existing row SHALL have `is_latest` set to `false`
- **AND** a new row SHALL be created with `version=2`, `is_latest=true`
- **AND** the new row SHALL have the same `use_count` and `relevance_score` as version 1
- **AND** `ContentSavedEvent` SHALL have `IsNew=false`, `NeedsGraph=false`, `Version=2`

#### Scenario: Third save
- **WHEN** `SaveKnowledge` is called with key "user-lang" that has latest version 2
- **THEN** version 2 becomes `is_latest=false`, version 3 is created with `is_latest=true`
- **AND** the database SHALL contain 3 rows for key "user-lang"

### Requirement: is_latest singleton invariant
For any key, at most one row SHALL have `is_latest=true`. This invariant SHALL be enforced by the SaveKnowledge transaction (set old false → create new true). If the transaction fails, it SHALL roll back both operations.

#### Scenario: Transaction atomicity
- **WHEN** SaveKnowledge fails after setting old `is_latest=false` but before creating the new version
- **THEN** the transaction SHALL roll back, restoring the old row's `is_latest=true`

#### Scenario: Invariant maintained after multiple saves
- **WHEN** SaveKnowledge is called 5 times for the same key
- **THEN** exactly one row SHALL have `is_latest=true` (version 5)
- **AND** exactly 4 rows SHALL have `is_latest=false` (versions 1-4)

### Requirement: Concurrent save retry-on-conflict
When two concurrent SaveKnowledge calls target the same key, the second call SHALL detect a `(key, version)` unique constraint violation and retry once. On retry, it SHALL re-read the latest version and create the next version number.

#### Scenario: Concurrent save succeeds on retry
- **WHEN** two goroutines concurrently call `SaveKnowledge` with key "rule-x"
- **THEN** both calls SHALL succeed
- **AND** the key SHALL have two new versions (e.g., version N+1 and N+2)

#### Scenario: Retry limit
- **WHEN** retry also fails with a unique constraint violation
- **THEN** `SaveKnowledge` SHALL return an error (not retry infinitely)

### Requirement: GetKnowledgeHistory
The system SHALL provide `GetKnowledgeHistory(ctx, key)` that returns all versions of a knowledge entry ordered by version descending. Each entry SHALL include `Version` and `CreatedAt` fields. If no entries exist for the key, it SHALL return `ErrKnowledgeNotFound`.

#### Scenario: History with multiple versions
- **WHEN** `GetKnowledgeHistory` is called for a key with 3 versions
- **THEN** it SHALL return 3 `KnowledgeEntry` items ordered `[v3, v2, v1]`
- **AND** each entry SHALL have `Version` and `CreatedAt` populated

#### Scenario: History for nonexistent key
- **WHEN** `GetKnowledgeHistory` is called for a key that does not exist
- **THEN** it SHALL return `ErrKnowledgeNotFound`

### Requirement: KnowledgeEntry domain type extension
`KnowledgeEntry` SHALL include `Version int` and `CreatedAt time.Time` fields. These fields SHALL have zero-value defaults for backward compatibility with callers that construct entries without setting them.

#### Scenario: Existing callers unaffected
- **WHEN** a caller constructs `KnowledgeEntry{Key: "x", Category: ..., Content: "y"}`
- **THEN** `Version` SHALL be `0` and `CreatedAt` SHALL be zero-value `time.Time{}`
- **AND** `SaveKnowledge` SHALL treat `Version=0` as "auto-assign"

#### Scenario: GetKnowledge populates version
- **WHEN** `GetKnowledge` returns an entry
- **THEN** `Version` and `CreatedAt` SHALL be populated from the database row

### Requirement: ContentSavedEvent version field
`ContentSavedEvent` SHALL include a `Version int` field. For knowledge saves, `Version` SHALL be the version number of the created entry. For non-knowledge collections (learning, memory), `Version` SHALL be `0`. This field SHALL be a typed struct field, NOT passed via the Metadata map.

#### Scenario: Knowledge event includes version
- **WHEN** a knowledge entry version 3 is created
- **THEN** `ContentSavedEvent.Version` SHALL be `3`

#### Scenario: Learning event version is zero
- **WHEN** a learning entry is saved
- **THEN** `ContentSavedEvent.Version` SHALL be `0`

### Requirement: Agent tool contract updates
The `save_knowledge` tool description SHALL indicate that saving appends a new version when the key exists. The tool return value SHALL include a `version` field with the created version number. A new `get_knowledge_history` tool SHALL be provided that accepts a `key` parameter and returns all versions of the knowledge entry.

#### Scenario: save_knowledge returns version
- **WHEN** the agent calls `save_knowledge` with key "pref-lang" (already at version 2)
- **THEN** the return SHALL include `{"status":"saved","key":"pref-lang","version":3,"message":"Knowledge 'pref-lang' saved (version 3)"}`

#### Scenario: get_knowledge_history returns versions
- **WHEN** the agent calls `get_knowledge_history` with key "pref-lang" (3 versions)
- **THEN** the return SHALL include `{"key":"pref-lang","versions":[{version:3,...},{version:2,...},{version:1,...}]}`

#### Scenario: get_knowledge_history for missing key
- **WHEN** the agent calls `get_knowledge_history` with a nonexistent key
- **THEN** the return SHALL include an error message indicating the key was not found
