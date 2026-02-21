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
- **THEN** the string SHALL be cast at the boundary: `Category: entknowledge.Category(category)`

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
The system SHALL provide persistent CRUD operations for knowledge entries identified by a unique key.

#### Scenario: Save new knowledge entry
- **WHEN** `SaveKnowledge` is called with a key that does not exist
- **THEN** the system SHALL create a new knowledge entry with the given key, category, content, tags, and source

#### Scenario: Update existing knowledge entry
- **WHEN** `SaveKnowledge` is called with a key that already exists
- **THEN** the system SHALL update the existing entry's category, content, tags, and source

#### Scenario: Get knowledge by key
- **WHEN** `GetKnowledge` is called with an existing key
- **THEN** the system SHALL return the knowledge entry
- **AND** if the key does not exist, SHALL return an error

#### Scenario: Delete knowledge by key
- **WHEN** `DeleteKnowledge` is called with an existing key
- **THEN** the system SHALL remove the entry from the store

#### Scenario: Increment knowledge use count
- **WHEN** `IncrementKnowledgeUseCount` is called with a valid key
- **THEN** the system SHALL increment the use count by 1

### Requirement: Knowledge Search
The system SHALL support keyword-based search across knowledge entries. `SearchKnowledge` SHALL split the query into individual keywords and create separate `ContentContains`/`KeyContains` LIKE predicates for each keyword, combined with OR logic. The system SHALL NOT use a single concatenated query string as a LIKE pattern.

#### Scenario: Search by query
- **WHEN** `SearchKnowledge` is called with a query string
- **THEN** the system SHALL return entries where the content or key contains any of the individual keywords
- **AND** results SHALL be ordered by relevance score descending
- **AND** results SHALL be limited to the specified limit (default 10)

#### Scenario: Multi-keyword search
- **WHEN** `SearchKnowledge` is called with query "deploy server config"
- **THEN** the SQL query uses per-keyword LIKE predicates: `(content LIKE '%deploy%' OR key LIKE '%deploy%') OR (content LIKE '%server%' OR key LIKE '%server%') OR (content LIKE '%config%' OR key LIKE '%config%')`

#### Scenario: Single keyword search
- **WHEN** `SearchKnowledge` is called with query "deploy"
- **THEN** the SQL query uses `content LIKE '%deploy%' OR key LIKE '%deploy%'`

#### Scenario: Search with category filter
- **WHEN** `SearchKnowledge` is called with a query and a category
- **THEN** the system SHALL return only entries matching both the query and the category

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
- **THEN** a `Knowledge` table SHALL exist with fields: key (unique), category (enum: rule/definition/preference/fact/pattern/correction), content, tags (JSON), source, relevance_score, use_count, created_at, updated_at

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
