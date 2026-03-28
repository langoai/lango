## ADDED Requirements

### Requirement: Knowledge entry structure
The `knowledge.KnowledgeEntry` struct SHALL use `entknowledge.Category` (Ent-generated type) for its `Category` field. The `knowledge.LearningEntry` struct SHALL use `entlearning.Category` (Ent-generated type) for its `Category` field. No duplicate domain enum types SHALL exist — the Ent-generated types are the single source of truth. The `string()` cast SHALL only occur at system boundaries: metadata maps and tool parameter parsing.

#### Scenario: Knowledge entry uses Ent category type directly
- **WHEN** a `KnowledgeEntry` is created in learning, librarian, or app packages
- **THEN** the `Category` field SHALL be assigned an `entknowledge.Category` value (e.g., `entknowledge.CategoryFact`)

#### Scenario: No DB boundary cast needed
- **WHEN** a knowledge entry is persisted via Ent `SetCategory()`
- **THEN** the category SHALL be passed directly: `SetCategory(entry.Category)` with no intermediate cast

#### Scenario: No read boundary cast needed
- **WHEN** a knowledge entry is loaded from Ent
- **THEN** the category SHALL be assigned directly: `Category: k.Category`

#### Scenario: Tool parameter boundary
- **WHEN** the `save_knowledge` tool receives a category string from tool parameters
- **THEN** the string SHALL be validated via `entknowledge.CategoryValidator()` at the boundary before use

### Requirement: Category Mapping
The system SHALL map LLM analysis type strings to valid `entknowledge.Category` enum values. The `mapCategory()` and `mapKnowledgeCategory()` functions SHALL return `(Category, error)` and SHALL return an error for any unrecognized type string instead of silently defaulting. Valid types SHALL include: `preference`, `fact`, `rule`, `definition`, `pattern`, `correction`.

#### Scenario: Valid type mapping
- **WHEN** a recognized type string (preference, fact, rule, definition, pattern, correction) is passed to `mapCategory()` or `mapKnowledgeCategory()`
- **THEN** the corresponding `entknowledge.Category` value SHALL be returned with a nil error

#### Scenario: Unrecognized type rejection
- **WHEN** an unrecognized type string is passed to `mapCategory()` or `mapKnowledgeCategory()`
- **THEN** an empty category and a non-nil error containing `"unrecognized knowledge type"` SHALL be returned

#### Scenario: Case sensitivity
- **WHEN** a type string with incorrect casing (e.g., `"FACT"`, `"Preference"`) is passed
- **THEN** the function SHALL return an error (types are case-sensitive)

#### Scenario: Metadata map boundary
- **WHEN** a knowledge entry category is placed into a `map[string]string` metadata map
- **THEN** the category SHALL be cast: `"category": string(entry.Category)`

#### Scenario: Learning entry uses Ent category type directly
- **WHEN** a `LearningEntry` is created in learning, app, or knowledge packages
- **THEN** the `Category` field SHALL be assigned an `entlearning.Category` value (e.g., `entlearning.CategoryToolError`)

#### Scenario: No Learning DB boundary cast needed
- **WHEN** a learning entry is persisted or loaded via Ent
- **THEN** the category SHALL be passed/assigned directly with no intermediate cast

### Requirement: LearningStats ByCategory field type
The `LearningStats.ByCategory` field SHALL use `map[entlearning.Category]int` instead of `map[string]int`. The map SHALL be populated using the `Category` field directly without string casting.

#### Scenario: ByCategory uses enum type as key
- **WHEN** `GetLearningStats` is called and learning entries exist
- **THEN** `ByCategory` map keys SHALL be of type `entlearning.Category`, not `string`

#### Scenario: JSON serialization compatibility
- **WHEN** `LearningStats` is serialized to JSON
- **THEN** the `by_category` field SHALL produce identical JSON output as the previous `map[string]int` representation

### Requirement: Knowledge CRUD Operations
The system SHALL provide persistent CRUD operations for knowledge entries identified by key. Knowledge entries SHALL be versioned: each save appends a new version instead of updating in place. All read operations SHALL default to the latest version (`is_latest=true`). When the latest version has the same `(category, content)` as the new entry, SaveKnowledge SHALL be a no-op (no new version created). Changes to `source`, `tags`, or temporal hints alone do NOT justify a new version.

#### Scenario: Save new knowledge entry
- **WHEN** `SaveKnowledge` is called with a key that does not exist
- **THEN** the system SHALL create a new knowledge entry with `version=1`, `is_latest=true`, and the given key, category, content, tags, and source

#### Scenario: Save existing knowledge entry (append version)
- **WHEN** `SaveKnowledge` is called with a key that already exists (latest version N) and the content or category differs
- **THEN** the system SHALL atomically set the existing latest row's `is_latest` to `false` and create a new entry with `version=N+1`, `is_latest=true`
- **AND** the new version SHALL carry forward `use_count` and `relevance_score` from the previous version

#### Scenario: Content-dedup no-op
- **WHEN** `SaveKnowledge` is called with a key whose latest version has the same `(category, content)`
- **THEN** the system SHALL return nil without creating a new version

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

### Requirement: Knowledge Scored Search
The system SHALL provide `SearchKnowledgeScored(ctx, query, category, limit)` returning `[]ScoredKnowledgeEntry` with normalized Score (higher=better) and SearchSource ("fts5"/"like"). FTS5 path SHALL negate BM25 rank for normalization. LIKE path SHALL use RelevanceScore. Existing `SearchKnowledge` SHALL remain unchanged.

#### Scenario: FTS5 scored search
- **WHEN** `SearchKnowledgeScored` is called and FTS5 index is available
- **THEN** results SHALL include `Score = -rank` (BM25 negated) and `SearchSource = "fts5"`

#### Scenario: LIKE scored search
- **WHEN** `SearchKnowledgeScored` is called and FTS5 is unavailable
- **THEN** results SHALL include `Score = RelevanceScore` and `SearchSource = "like"`

#### Scenario: Scored search returns latest only
- **WHEN** `SearchKnowledgeScored` is called
- **THEN** only `is_latest=true` entries SHALL be returned

### Requirement: Learning Scored Search
The system SHALL provide `SearchLearningsScored(ctx, errorPattern, category, limit)` returning `[]ScoredLearningEntry` with Score and SearchSource.

#### Scenario: Learning scored search
- **WHEN** `SearchLearningsScored` is called
- **THEN** results SHALL include `Score = Confidence` and `SearchSource = "like"`

### Requirement: Knowledge relevance score mutations
The system SHALL provide `BoostRelevanceScore(ctx, key, delta, maxScore)` that atomically increases relevance_score for the latest version, clamped at maxScore via two-step update (add where safe, cap remainder). It SHALL provide `DecayAllRelevanceScores(ctx, delta, minScore)` that subtracts delta from all latest-version entries, floored at minScore via two-step update (subtract where safe, floor remainder). It SHALL provide `ResetAllRelevanceScores(ctx)` that sets all latest-version scores to 1.0.

#### Scenario: Boost with cap (two-step clamping)
- **WHEN** BoostRelevanceScore is called and current score + delta would exceed maxScore
- **THEN** score SHALL be set to maxScore (not exceed it)

#### Scenario: Boost within range
- **WHEN** BoostRelevanceScore is called and current score + delta <= maxScore
- **THEN** score SHALL increase by delta

#### Scenario: Decay with floor (two-step clamping)
- **WHEN** DecayAllRelevanceScores is called and current score - delta would go below minScore
- **THEN** score SHALL be set to minScore (not go below it)

#### Scenario: Decay within range
- **WHEN** DecayAllRelevanceScores is called and current score - delta >= minScore
- **THEN** score SHALL decrease by delta

#### Scenario: Reset all
- **WHEN** ResetAllRelevanceScores is called
- **THEN** all latest-version entries SHALL have relevance_score set to 1.0

### Requirement: Learning CRUD Operations
The system SHALL provide persistent CRUD operations for learning entries.

#### Scenario: Save new learning
- **WHEN** `SaveLearning` is called with trigger, error pattern, diagnosis, fix, and category
- **THEN** the system SHALL create a new learning entry

#### Scenario: Search learnings
- **WHEN** `SearchLearnings` is called with an error pattern query
- **THEN** the system SHALL split the query into individual keywords and create separate `ErrorPatternContains`/`TriggerContains` LIKE predicates for each keyword, combined with OR logic
- **AND** results SHALL be ordered by confidence descending

#### Scenario: Boost learning confidence
- **WHEN** `BoostLearningConfidence` is called with a learning ID and success delta
- **THEN** the system SHALL increment success count and recalculate confidence as `success / (success + occurrence)`
- **AND** confidence SHALL NOT drop below 0.1

### Requirement: Skill Persistence
The system SHALL provide persistent CRUD operations for skill entries.

#### Scenario: Save new skill
- **WHEN** `SaveSkill` is called with a skill entry
- **THEN** the system SHALL create the skill with default status `draft`

#### Scenario: Activate skill
- **WHEN** `ActivateSkill` is called with a skill name
- **THEN** the system SHALL set the skill status to `active`

#### Scenario: List active skills
- **WHEN** `ListActiveSkills` is called
- **THEN** the system SHALL return all skills with status `active`

#### Scenario: Increment skill usage
- **WHEN** `IncrementSkillUsage` is called with a skill name and success flag
- **THEN** the system SHALL increment use count by 1 and update last used timestamp
- **AND** if success is true, SHALL also increment success count

### Requirement: Audit Logging
The system SHALL maintain an append-only audit log of knowledge operations.

#### Scenario: Save audit log entry
- **WHEN** `SaveAuditLog` is called with an action, actor, target, and optional details
- **THEN** the system SHALL create an immutable audit log record

### Requirement: External Reference Management
The system SHALL support CRUD operations for external references (docs, APIs, wiki links).

#### Scenario: Save or update external reference
- **WHEN** `SaveExternalRef` is called with a name that does not exist
- **THEN** the system SHALL create the external reference
- **AND** if the name already exists, SHALL update the existing reference

#### Scenario: Search external references
- **WHEN** `SearchExternalRefs` is called with a query
- **THEN** the system SHALL split the query into individual keywords and create separate `NameContains`/`SummaryContains` LIKE predicates for each keyword, combined with OR logic

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

### Requirement: Knowledge configuration exposed in TUI
The Onboard TUI SHALL provide a dedicated Knowledge configuration form accessible from the main menu.

#### Scenario: Knowledge menu category
- **WHEN** user views the Configuration Menu in the onboard wizard
- **THEN** a "Knowledge" category SHALL appear with label "🧠 Knowledge" and description "Learning, Skills, Context limits"

#### Scenario: Knowledge form fields
- **WHEN** user selects the Knowledge category
- **THEN** the form SHALL display 4 fields:
  - knowledge_enabled (boolean toggle)
  - knowledge_max_context (integer input for MaxContextPerLayer)
  - knowledge_auto_approve (boolean toggle for AutoApproveSkills)
  - knowledge_max_skills_day (integer input for MaxSkillsPerDay)

#### Scenario: Knowledge config persistence
- **WHEN** user modifies Knowledge form fields and saves
- **THEN** values SHALL be written to the `knowledge` section of `lango.json`

### Requirement: Knowledge retrieval result truncation
The system SHALL provide a `TruncateResult(result *RetrievalResult, budgetTokens int) *RetrievalResult` function that reduces a `RetrievalResult` to fit within a token budget. Truncation SHALL operate at the item level — removing lower-priority items from each layer until the total estimated tokens fit within the budget. The function SHALL NOT modify assembled text; it operates on `RetrievalResult` before `AssemblePrompt()` is called.

#### Scenario: Result fits within budget
- **WHEN** `TruncateResult` is called with a result whose total tokens are within budget
- **THEN** the result SHALL be returned unchanged

#### Scenario: Result exceeds budget
- **WHEN** `TruncateResult` is called with a result exceeding the budget
- **THEN** items SHALL be removed from the end of each layer (lowest priority first) until the total fits
- **AND** the layer structure and headings SHALL remain intact

#### Scenario: Zero budget means unlimited
- **WHEN** `TruncateResult` is called with `budgetTokens == 0`
- **THEN** the result SHALL be returned unchanged (0 = unlimited, legacy mode)

#### Scenario: Budget too small for any items
- **WHEN** `TruncateResult` is called with a budget smaller than any single item
- **THEN** the result SHALL be empty (zero items) but not nil

### Requirement: ContextLayer String method
The `ContextLayer` type SHALL implement the `String()` method returning human-readable snake_case names for all 9 known layers. Unknown layer values SHALL return `"layer_N"` where N is the integer value.

#### Scenario: All known layers have string names
- **WHEN** `String()` is called on each of the 9 known `ContextLayer` values
- **THEN** each SHALL return a distinct non-empty snake_case string

#### Scenario: Unknown layer fallback
- **WHEN** `String()` is called on an unrecognized `ContextLayer` value
- **THEN** it SHALL return `"layer_N"` format without panicking

### Requirement: VectorStore build-tag isolation
The `sqlite-vec` backend SHALL be compiled only when the `vec` build tag is present. The default build (no `vec` tag) SHALL produce a binary without sqlite-vec linked. The `NewVectorStore` factory function SHALL be the canonical entry point for VectorStore creation in wiring code.

#### Scenario: Build with vec tag
- **WHEN** the binary is built with `-tags vec`
- **THEN** `NewVectorStore(db, dimensions)` SHALL return a `*SQLiteVecStore` implementing `VectorStore`

#### Scenario: Build without vec tag
- **WHEN** the binary is built without the `vec` tag
- **THEN** `NewVectorStore(db, dimensions)` SHALL return `nil, ErrVecNotCompiled`

#### Scenario: Wiring graceful degradation
- **WHEN** `NewVectorStore` returns an error during embedding initialization
- **THEN** the system SHALL log a warning and return nil `embeddingComponents` with a FeatureStatus suggesting "rebuild with -tags vec"
- **AND** RAG, EmbeddingBuffer, and ContextSearchAgent SHALL NOT be registered

### Requirement: Coordinator primary retrieval
The `RetrievalCoordinator` SHALL run as the primary retrieval path for factual layers (UserKnowledge, AgentLearnings, ExternalKnowledge) in Phase 1 of `GenerateContent`. The old `ContextRetriever` SHALL handle non-factual layers only (RuntimeContext, ToolRegistry, SkillPatterns, PendingInquiries).

#### Scenario: Phase 1 parallel retrieval
- **WHEN** `GenerateContent` is called with a non-empty user query and both retriever and coordinator are configured
- **THEN** both SHALL run as parallel goroutines in Phase 1
- **AND** retriever SHALL request only non-factual layers
- **AND** coordinator SHALL retrieve factual layers

#### Scenario: Result merge
- **WHEN** both retriever and coordinator return results
- **THEN** `mergeRetrievalResults` SHALL combine Items maps from both sources
- **AND** TotalItems SHALL equal the sum of both results' items

#### Scenario: ContextInjectedEvent coverage
- **WHEN** `ContextInjectedEvent` is published after context assembly
- **THEN** `Items` SHALL contain items from the merged result (both factual and non-factual layers)

### Requirement: RAG enabled flag enforcement
The system SHALL NOT create `RAGService` or register `ContextSearchAgent` when `embedding.rag.enabled` is false. The embedding buffer and provider SHALL still be initialized for async knowledge embedding regardless of the RAG flag.

#### Scenario: RAG disabled with embedding configured
- **WHEN** `embedding.provider` is configured but `embedding.rag.enabled` is false
- **THEN** `ragService` SHALL be nil in embeddingComponents
- **AND** `ContextSearchAgent` SHALL NOT be registered in the coordinator

### Requirement: Settings TUI explicit key preservation
The system SHALL mark all context-related config keys as explicitly set when saving from the settings TUI. This prevents `ResolveContextAutoEnable` from overriding user intent on subsequent bootstrap.

#### Scenario: Settings save preserves disabled flags
- **WHEN** user sets `knowledge.enabled=false` in settings TUI and saves
- **THEN** the saved explicitKeys SHALL include `knowledge.enabled`
- **AND** subsequent bootstrap SHALL NOT auto-enable knowledge
