## Purpose

The tool health diagnostics capability provides an agent-facing diagnostic tool (`builtin_health`) that reports tool registration health status, enabling agents to self-diagnose why tools may be missing.

## ADDED Requirements

### Requirement: builtin_health diagnostic tool
The system SHALL provide a `builtin_health` tool in `BuildDispatcher()` that reports tool registration health status. It SHALL return all categories grouped by enabled/disabled state, with config key hints for disabled categories.

#### Scenario: Health check with all enabled categories
- **WHEN** `builtin_health` is invoked and all registered categories are enabled
- **THEN** it SHALL return `enabled_categories` listing each category with name, description, and tool_count
- **AND** `disabled_categories` SHALL be empty or nil

#### Scenario: Health check with disabled categories
- **WHEN** `builtin_health` is invoked and some categories are disabled
- **THEN** disabled categories SHALL appear in `disabled_categories` with name, description, and a `hint` field containing the config key needed to enable them

#### Scenario: Health tool safety level
- **WHEN** `BuildDispatcher()` creates the dispatcher tools
- **THEN** `builtin_health` SHALL have SafetyLevelSafe
