## MODIFIED Requirements

### Requirement: Ent Schema Definitions
The Ent Knowledge schema's `category` enum field SHALL include all values defined in the `KnowledgeCategory` domain type: `rule`, `definition`, `preference`, `fact`, `pattern`, `correction`.

#### Scenario: Knowledge schema with extended categories
- **WHEN** the database is migrated
- **THEN** the `Knowledge` table's `category` enum SHALL accept values: `rule`, `definition`, `preference`, `fact`, `pattern`, `correction`

### Requirement: Knowledge entry structure
The `knowledge.KnowledgeEntry` struct SHALL use `knowledge.KnowledgeCategory` for its `Category` field. The `knowledge.LearningEntry` struct SHALL use `knowledge.LearningCategory` for its `Category` field instead of plain `string`. Internal code that creates `LearningEntry` values SHALL use typed category constants (`knowledge.LearningToolError`, `knowledge.LearningUserCorrection`, etc.). The `string()` cast SHALL only occur at system boundaries: Ent DB writes, metadata maps, and tool parameter parsing.

#### Scenario: Learning entry uses typed category
- **WHEN** a `LearningEntry` is created in learning, app, or knowledge packages
- **THEN** the `Category` field SHALL be assigned a `knowledge.LearningCategory` value

#### Scenario: Learning DB boundary cast on write
- **WHEN** a learning entry is persisted via Ent `SetCategory()`
- **THEN** the category SHALL be cast: `SetCategory(entlearning.Category(string(entry.Category)))`

#### Scenario: Learning DB boundary cast on read
- **WHEN** a learning entry is loaded from Ent
- **THEN** the category SHALL be cast: `Category: LearningCategory(l.Category)`

#### Scenario: Learning tool parameter boundary
- **WHEN** the `save_learning` tool receives a category string from tool parameters
- **THEN** the string SHALL be cast at the boundary: `Category: knowledge.LearningCategory(category)`

#### Scenario: Learning metadata map boundary
- **WHEN** a learning entry category is placed into a `map[string]string` metadata map
- **THEN** the category SHALL be cast: `"category": string(entry.Category)`
