## Purpose

The tool catalog provides a centralized registry for built-in tools grouped by named categories, with dispatcher tools (`builtin_list`, `builtin_invoke`, `builtin_health`) for dynamic discovery, invocation, and diagnostics at runtime.

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
The system SHALL provide `BuildDispatcher(catalog)` returning three tools: `builtin_list`, `builtin_invoke`, and `builtin_health`.

#### Scenario: builtin_list returns tool catalog
- **WHEN** `builtin_list` is invoked with no parameters
- **THEN** it SHALL return all categories (including disabled ones) and all tools with their schemas
- **AND** the total count of registered tools

#### Scenario: builtin_list filters by category
- **WHEN** `builtin_list` is invoked with `category: "exec"`
- **THEN** it SHALL return only tools in the "exec" category

#### Scenario: builtin_invoke executes a safe registered tool
- **WHEN** `builtin_invoke` is invoked with a tool_name whose SafetyLevel is less than Dangerous
- **THEN** it SHALL execute the tool's handler and return `{tool, result}`

#### Scenario: builtin_invoke blocks dangerous tools
- **WHEN** `builtin_invoke` is invoked with a tool_name whose SafetyLevel is Dangerous or higher
- **THEN** it SHALL return an error containing "requires approval" and "delegate to the appropriate sub-agent"
- **AND** it SHALL NOT execute the tool's handler

#### Scenario: builtin_invoke rejects unknown tool
- **WHEN** `builtin_invoke` is invoked with a tool_name not in the catalog
- **THEN** it SHALL return an error containing "not found in catalog"

### Requirement: Safety levels
`builtin_list` SHALL have SafetyLevelSafe. `builtin_invoke` SHALL have SafetyLevelDangerous. `builtin_health` SHALL have SafetyLevelSafe.

#### Scenario: Safety level assignment
- **WHEN** `BuildDispatcher()` creates the dispatcher tools
- **THEN** `builtin_list` safety level SHALL be Safe
- **AND** `builtin_invoke` safety level SHALL be Dangerous
- **AND** `builtin_health` safety level SHALL be Safe

### Requirement: Disabled category registration
The system SHALL register disabled categories in the tool catalog when a subsystem is not initialized, so that `builtin_list` and `builtin_health` can report their existence and required config keys.

#### Scenario: Smart account disabled registers disabled category
- **WHEN** `initSmartAccount()` returns nil (smart account disabled or payment missing)
- **THEN** a `smartaccount` category SHALL be registered with `Enabled: false`
- **AND** the category description SHALL include instructions for enabling

#### Scenario: Disabled category visible in builtin_list
- **WHEN** `builtin_list` is invoked
- **THEN** disabled categories SHALL appear in the `categories` list with `enabled: false`

### Requirement: ToolNamesForCategory query
The Catalog SHALL provide a `ToolNamesForCategory(category string) []string` method that returns tool names registered under the given category in insertion order.

#### Scenario: Query tool names for existing category
- **WHEN** `ToolNamesForCategory("cron")` is called and tools cron_add, cron_list, cron_remove are registered under "cron"
- **THEN** it SHALL return `["cron_add", "cron_list", "cron_remove"]`

#### Scenario: Query tool names for empty category
- **WHEN** `ToolNamesForCategory("nonexistent")` is called
- **THEN** it SHALL return nil

### Requirement: EnabledCategorySummary query
The Catalog SHALL provide an `EnabledCategorySummary() map[string][]string` method returning a map of enabled category names to their tool name lists.

#### Scenario: Summary with mixed categories
- **WHEN** `EnabledCategorySummary()` is called with enabled category "exec" (2 tools) and disabled category "cron"
- **THEN** the map SHALL contain key "exec" with 2 tool names
- **AND** the map SHALL NOT contain key "cron"

### Requirement: Dynamic tool catalog prompt section
The system SHALL inject a `SectionToolCatalog` prompt section (priority 410) into the agent system prompt listing active tool categories with up to 8 representative tool names each, and disabled categories with their config keys.

#### Scenario: Prompt includes active categories
- **WHEN** the system prompt is built with enabled categories "exec", "cron"
- **THEN** the prompt SHALL contain a "Available Tool Categories" section listing each category with description and tool names

#### Scenario: Prompt includes disabled category notice
- **WHEN** the system prompt is built with disabled category "smartaccount" (configKey: "smartAccount.enabled")
- **THEN** the prompt SHALL contain text mentioning "smartaccount" and "smartAccount.enabled" under disabled categories

### Requirement: Orchestrator routing entry tool names
The orchestrator routing table SHALL include tool name lists per sub-agent, rendering up to 10 tool names per agent in the instruction.

#### Scenario: Routing entry includes tool names
- **WHEN** the orchestrator instruction is built with an automator agent assigned cron_add, cron_list, cron_remove
- **THEN** the routing table entry for "automator" SHALL contain a "Tools" line listing those tool names
## Requirements
### Requirement: Comprehensive disabled category registration
Every tool subsystem SHALL register a disabled category with the tool catalog when it is not active, so that builtin_health diagnostics can report the full system state. The disabled category SHALL include the relevant configKey.

#### Scenario: Disabled subsystems appear in catalog
- **WHEN** a subsystem (browser, crypto, secrets, meta, graph, rag, memory, agent_memory, payment, p2p, librarian, economy, mcp, observability, contract, workspace) is disabled
- **THEN** a disabled category is registered with Name, Description containing "(disabled)", ConfigKey, and Enabled=false

#### Scenario: builtin_health reports disabled subsystems
- **WHEN** builtin_health runs diagnostics
- **THEN** all disabled subsystems appear in the disabled list of the tool registration summary

#### Scenario: P2P disabled with payment dependency
- **WHEN** p2p.enabled is true but payment is disabled
- **THEN** p2p disabled category description includes "(disabled — payment required)"

