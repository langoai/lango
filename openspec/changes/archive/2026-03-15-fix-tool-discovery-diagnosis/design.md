## Context

AI agents cannot find tools at runtime despite the tools being defined in code. Root causes: (1) config flags silently skip tool registration, (2) builtin_health shows category counts but not tool names, (3) the LLM system prompt contains no tool catalog awareness, (4) the multi-agent orchestrator's routing table lacks tool-level detail. The fix targets all four layers.

## Goals / Non-Goals

**Goals:**
- Make disabled tool categories visible in builtin_health with actionable enable commands
- Surface tool names per category in builtin_health output
- Inject active tool category summaries into the LLM system prompt
- Add tool name lists to orchestrator routing entries
- Log tool registration summary at app startup for debugging

**Non-Goals:**
- Changing the conditional wiring logic (tools remain config-gated)
- Dynamic tool loading or hot-reload
- Changing tool naming conventions

## Decisions

### 1. Register disabled categories in the Catalog
Disabled systems (cron, background, workflow) now register `Enabled: false` categories in the Catalog before the dispatcher is built. This makes them visible to builtin_health without adding tools.

**Alternative**: Post-hoc scan of config flags in builtin_health. Rejected because the Catalog is the authoritative source and should be self-describing.

### 2. Dynamic SectionToolCatalog prompt section
A new prompt section (priority 410, between ToolUsage at 400 and Automation at 450) renders active tool categories with up to 8 tool names each, plus a list of disabled categories. Built from `Catalog.EnabledCategorySummary()`.

**Alternative**: Static prompt listing all possible tools. Rejected because it would be stale when config changes.

### 3. ToolNames in orchestrator routing entries
`routingEntry` gains a `ToolNames []string` field, populated from the tools assigned to each sub-agent. The orchestrator instruction renders up to 10 tool names per agent.

**Alternative**: Only use capability descriptions. Rejected because the LLM benefits from seeing exact tool names for matching.

### 4. Startup diagnostic log
A single Info-level log line at the end of tool registration showing `total`, `enabled` (category(count) pairs), and `disabled` (category names).

## Risks / Trade-offs

- **Prompt token budget**: Adding tool names to the system prompt increases token usage. Mitigated by capping at 8 names per category and 10 per routing entry.
- **Stale disabled list**: If a new tool system is added without registering a disabled category, it won't appear in diagnostics. Mitigated by convention (same pattern as existing smartaccount disabled registration).
