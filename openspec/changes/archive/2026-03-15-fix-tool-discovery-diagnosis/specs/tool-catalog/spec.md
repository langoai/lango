## ADDED Requirements

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
