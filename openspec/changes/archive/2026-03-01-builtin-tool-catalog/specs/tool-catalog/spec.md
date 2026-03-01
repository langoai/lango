## ADDED Requirements

### Requirement: Tool Catalog registry
The system SHALL provide a thread-safe `Catalog` type in `internal/toolcatalog/` that registers built-in tools grouped by named categories.

#### Scenario: Register and retrieve a tool
- **WHEN** a tool is registered under category "exec" via `Register("exec", tools)`
- **THEN** `Get(toolName)` SHALL return the tool entry with its category

#### Scenario: List categories
- **WHEN** multiple categories are registered via `RegisterCategory()`
- **THEN** `ListCategories()` SHALL return all categories sorted by name

#### Scenario: List tools by category
- **WHEN** tools are registered under multiple categories
- **THEN** `ListTools("exec")` SHALL return only tools in the "exec" category
- **AND** `ListTools("")` SHALL return all tools across all categories

#### Scenario: Tool count
- **WHEN** tools are registered
- **THEN** `ToolCount()` SHALL return the total number of unique tools
- **AND** re-registering the same tool SHALL NOT increase the count

### Requirement: Dispatcher tools
The system SHALL provide `BuildDispatcher(catalog)` returning two tools: `builtin_list` and `builtin_invoke`.

#### Scenario: builtin_list returns tool catalog
- **WHEN** `builtin_list` is invoked with no parameters
- **THEN** it SHALL return all categories and all tools with their schemas
- **AND** the total count of registered tools

#### Scenario: builtin_list filters by category
- **WHEN** `builtin_list` is invoked with `category: "exec"`
- **THEN** it SHALL return only tools in the "exec" category

#### Scenario: builtin_invoke executes a registered tool
- **WHEN** `builtin_invoke` is invoked with `tool_name: "exec_shell"` and valid params
- **THEN** it SHALL execute the tool's handler and return `{tool, result}`

#### Scenario: builtin_invoke rejects unknown tool
- **WHEN** `builtin_invoke` is invoked with a tool_name not in the catalog
- **THEN** it SHALL return an error containing "not found in catalog"

### Requirement: Safety levels
`builtin_list` SHALL have SafetyLevelSafe. `builtin_invoke` SHALL have SafetyLevelDangerous.

#### Scenario: Safety level assignment
- **WHEN** `BuildDispatcher()` creates the dispatcher tools
- **THEN** `builtin_list` safety level SHALL be Safe
- **AND** `builtin_invoke` safety level SHALL be Dangerous
