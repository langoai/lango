## MODIFIED Requirements

### Requirement: Knowledge CRUD Operations
The system SHALL provide persistent CRUD operations for knowledge entries identified by key. Knowledge entries SHALL be versioned: each save appends a new version instead of updating in place. All read operations SHALL default to the latest version (`is_latest=true`).

#### Scenario: Save new knowledge entry
- **WHEN** `SaveKnowledge` is called with a key that does not exist
- **THEN** the system SHALL create a new knowledge entry with `version=1`, `is_latest=true`, and the given key, category, content, tags, and source

#### Scenario: Save existing knowledge entry (append version)
- **WHEN** `SaveKnowledge` is called with a key that already exists (latest version N)
- **THEN** the system SHALL atomically set the existing latest row's `is_latest` to `false` and create a new entry with `version=N+1`, `is_latest=true`
- **AND** the new version SHALL carry forward `use_count` and `relevance_score` from the previous version

#### Scenario: Get knowledge by key
- **WHEN** `GetKnowledge` is called with an existing key
- **THEN** the system SHALL return the latest version (`is_latest=true`) of the knowledge entry with `Version` and `CreatedAt` populated
- **AND** if no latest entry exists for the key, SHALL return an error

#### Scenario: Delete knowledge by key
- **WHEN** `DeleteKnowledge` is called with an existing key
- **THEN** the system SHALL remove ALL versions of the entry from the store

#### Scenario: Increment knowledge use count
- **WHEN** `IncrementKnowledgeUseCount` is called with a valid key
- **THEN** the system SHALL increment the use count by 1 on the latest version only (`is_latest=true`)

### Requirement: Knowledge Search
The system SHALL support full-text search across knowledge entries. All search operations SHALL return only the latest version (`is_latest=true`) of each key. When an FTS5 index is available, `SearchKnowledge` SHALL use FTS5 MATCH with BM25 ranking as the primary search path. When FTS5 is unavailable, `SearchKnowledge` SHALL fall back to per-keyword `ContentContains`/`KeyContains` LIKE predicates combined with OR logic, ordered by `RelevanceScore` descending. The system SHALL NOT use a single concatenated query string as a LIKE pattern in either path.

#### Scenario: FTS5 search path
- **WHEN** `SearchKnowledge` is called with a query string and FTS5 index is available
- **THEN** the system SHALL return latest-version entries matching the FTS5 query, ordered by BM25 relevance
- **AND** results SHALL be limited to the specified limit (default 10)

#### Scenario: LIKE fallback search path
- **WHEN** `SearchKnowledge` is called with a query string and FTS5 index is NOT available
- **THEN** the system SHALL return latest-version entries where the content or key contains any of the individual keywords
- **AND** the LIKE path SHALL include `is_latest=true` as a predicate
- **AND** results SHALL be ordered by relevance score descending
- **AND** results SHALL be limited to the specified limit (default 10)

#### Scenario: Multi-keyword FTS5 search
- **WHEN** `SearchKnowledge` is called with query "deploy server config" and FTS5 is available
- **THEN** the FTS5 MATCH query SHALL search for all keywords and rank by BM25

#### Scenario: Multi-keyword LIKE fallback
- **WHEN** `SearchKnowledge` is called with query "deploy server config" and FTS5 is NOT available
- **THEN** the SQL query uses per-keyword LIKE predicates: `(content LIKE '%deploy%' OR key LIKE '%deploy%') OR (content LIKE '%server%' OR key LIKE '%server%') OR (content LIKE '%config%' OR key LIKE '%config%')`
- **AND** an `is_latest = true` predicate SHALL be included

#### Scenario: Search with category filter
- **WHEN** `SearchKnowledge` is called with a query and a category
- **THEN** the system SHALL return only latest-version entries matching both the query and the category (in both FTS5 and LIKE paths)

#### Scenario: FTS5 error graceful degradation
- **WHEN** `SearchKnowledge` via FTS5 encounters an error
- **THEN** the system SHALL log a warning and fall back to the LIKE path for that query

#### Scenario: Search returns only latest version
- **WHEN** a key has version 1 with content "old data" and version 2 with content "new data"
- **AND** `SearchKnowledge` is called with query "old"
- **THEN** the search SHALL NOT return the key (version 1 is not latest)

### Requirement: Ent Schema Definitions
The system SHALL define Ent ORM schemas for the 5 knowledge entities.

#### Scenario: Knowledge schema
- **WHEN** the database is migrated
- **THEN** a `Knowledge` table SHALL exist with fields: key, version (int, default 1), is_latest (bool, default true), category (enum: rule/definition/preference/fact/pattern/correction), content, tags (JSON), source, relevance_score, use_count, created_at, updated_at
- **AND** a composite unique index SHALL exist on `(key, version)`
- **AND** a non-unique index SHALL exist on `(key, is_latest)`
- **AND** the `key` field SHALL NOT have a single-column UNIQUE constraint

#### Scenario: Learning schema
- **WHEN** the database is migrated
- **THEN** a `Learning` table SHALL exist with fields: trigger, error_pattern, diagnosis, fix, category (enum: tool_error/provider_error/user_correction/timeout/permission/general), tags (JSON), confidence, occurrence_count, success_count, created_at, updated_at

#### Scenario: Skill schema
- **WHEN** the database is migrated
- **THEN** a `Skill` table SHALL exist with fields: name (unique), description, skill_type (enum: composite/script/template), definition (JSON), parameters (JSON), status (enum: draft/active/deprecated), created_by, requires_approval, use_count, success_count, last_used_at, created_at, updated_at

#### Scenario: AuditLog schema
- **WHEN** the database is migrated
- **THEN** an `AuditLog` table SHALL exist with fields: session_key, action (enum), actor, target, details (JSON), created_at

#### Scenario: ExternalRef schema
- **WHEN** the database is migrated
- **THEN** an `ExternalRef` table SHALL exist with fields: name (unique), ref_type (enum: documentation/api/wiki/repository/tool), location, summary, metadata (JSON), created_at, updated_at
