## Why

AI agents fail to find registered tools at runtime (e.g., cron_remove, smart_account_info) because: (1) config flags silently disable tool categories without user-visible feedback, (2) builtin_health lacks tool-level detail, (3) the LLM system prompt has no awareness of which tool categories exist, and (4) the multi-agent orchestrator lacks tool-name-level routing information. Users see "tool not found" with no diagnostic path.

## What Changes

- Add diagnostic logging at app startup showing registered/disabled tool categories with tool counts
- Register disabled categories (cron, background, workflow) in the catalog so `builtin_health` can report them
- Enhance `builtin_health` to include tool name lists per enabled category and actionable `lango config set` hints for disabled categories
- Add a dynamic `SectionToolCatalog` prompt section that injects active tool categories and representative tool names into the LLM system prompt
- Add tool name lists to orchestrator routing entries so sub-agent delegation is informed by actual tool availability
- Add `ToolNamesForCategory` and `EnabledCategorySummary` methods to `Catalog`

## Capabilities

### New Capabilities

### Modified Capabilities
- `tool-health-diagnostics`: builtin_health now returns tool name lists per category and actionable enable/disable hints
- `tool-catalog`: Catalog gains ToolNamesForCategory and EnabledCategorySummary methods; disabled categories are registered for non-enabled systems

## Impact

- `internal/app/app.go`: Diagnostic log function, disabled category registration for cron/background/workflow
- `internal/toolcatalog/catalog.go`: New query methods (ToolNamesForCategory, EnabledCategorySummary)
- `internal/toolcatalog/dispatcher.go`: Enhanced builtin_health handler
- `internal/prompt/section.go`: New SectionToolCatalog ID
- `internal/app/wiring.go`: buildToolCatalogSection function, wired into initAgent prompt builder
- `internal/orchestration/tools.go`: ToolNames field in routingEntry, rendered in orchestrator instruction
- `internal/orchestration/orchestrator.go`: Pass tools to buildRoutingEntry
