## Purpose

The tool health diagnostics capability provides an agent-facing diagnostic tool (`builtin_health`) that reports tool registration health status, enabling agents to self-diagnose why tools may be missing.

## ADDED Requirements

### Requirement: builtin_health diagnostic tool
The system SHALL provide a `builtin_health` tool in `BuildDispatcher()` that reports tool registration health status. It SHALL return all categories grouped by enabled/disabled state, with tool name lists for enabled categories and actionable `lango config set` hints for disabled categories.

#### Scenario: Health check with all enabled categories
- **WHEN** `builtin_health` is invoked and all registered categories are enabled
- **THEN** it SHALL return `enabled_categories` listing each category with name, description, tool_count, and a `tools` field containing the list of tool names
- **AND** `disabled_categories` SHALL be empty or nil

#### Scenario: Health check with disabled categories
- **WHEN** `builtin_health` is invoked and some categories are disabled
- **THEN** disabled categories SHALL appear in `disabled_categories` with name, description, and a `hint` field containing an actionable command like `lango config set <configKey> true`

#### Scenario: Health tool safety level
- **WHEN** `BuildDispatcher()` creates the dispatcher tools
- **THEN** `builtin_health` SHALL have SafetyLevelSafe

### Requirement: Disabled automation categories registered
The system SHALL register disabled categories for cron, background, and workflow systems when their respective config flags are false, so builtin_health can report them.

#### Scenario: Cron disabled registers disabled category
- **WHEN** `cron.enabled` is false
- **THEN** a `cron` category SHALL be registered with `Enabled: false` and `ConfigKey: "cron.enabled"`

#### Scenario: Background disabled registers disabled category
- **WHEN** `background.enabled` is false
- **THEN** a `background` category SHALL be registered with `Enabled: false` and `ConfigKey: "background.enabled"`

#### Scenario: Workflow disabled registers disabled category
- **WHEN** `workflow.enabled` is false
- **THEN** a `workflow` category SHALL be registered with `Enabled: false` and `ConfigKey: "workflow.enabled"`

### Requirement: Tool registration diagnostic log
The system SHALL log a summary of tool registration at app startup including total tool count, enabled categories with their tool counts, and disabled category names.

#### Scenario: Startup log with mixed categories
- **WHEN** the app initializes with some enabled and some disabled tool categories
- **THEN** an Info-level log SHALL be emitted with fields: total, enabled (formatted as "category(count)"), and disabled (comma-separated names)
