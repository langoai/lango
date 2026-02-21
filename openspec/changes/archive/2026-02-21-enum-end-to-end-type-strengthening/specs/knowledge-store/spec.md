## MODIFIED Requirements

### Requirement: Knowledge entry structure
The `knowledge.KnowledgeEntry` struct SHALL use `knowledge.KnowledgeCategory` for its `Category` field instead of plain `string`. Internal code that creates `KnowledgeEntry` values SHALL use typed category constants (`knowledge.CategoryFact`, `knowledge.CategoryPreference`, `knowledge.CategoryRule`, `knowledge.CategoryDefinition`, etc.). The `string()` cast SHALL only occur at system boundaries: Ent DB writes, metadata maps, and tool parameter parsing.

#### Scenario: Knowledge entry uses typed category
- **WHEN** a `KnowledgeEntry` is created in learning, librarian, or app packages
- **THEN** the `Category` field SHALL be assigned a `knowledge.KnowledgeCategory` value

#### Scenario: DB boundary cast on write
- **WHEN** a knowledge entry is persisted via Ent `SetCategory()`
- **THEN** the category SHALL be cast: `SetCategory(entknowledge.Category(string(entry.Category)))`

#### Scenario: DB boundary cast on read
- **WHEN** a knowledge entry is loaded from Ent
- **THEN** the category SHALL be cast: `Category: KnowledgeCategory(k.Category)`

#### Scenario: Tool parameter boundary
- **WHEN** the `save_knowledge` tool receives a category string from tool parameters
- **THEN** the string SHALL be cast at the boundary: `Category: knowledge.KnowledgeCategory(category)`

#### Scenario: Metadata map boundary
- **WHEN** a knowledge entry category is placed into a `map[string]string` metadata map
- **THEN** the category SHALL be cast: `"category": string(entry.Category)`
