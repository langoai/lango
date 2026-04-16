## Context

`ToolCapability` (`agent/capability.go`) already has:
- `ReadOnly bool` — "tool performs no mutations"
- `Activity ActivityKind` — `read`, `write`, `execute`, `query`, `manage`

The `Catalog` (`toolcatalog/catalog.go`) holds all registered tools with their capabilities. The `KnowledgeSaveHook` needs the list at construction time in `app.go` builder.

Currently `BuildHookRegistry` in `app.go` uses `toolchain.DefaultSaveableTools` (hard-coded constant from Wave 1). The catalog is available at the same wiring point (`app.go:150+` — catalog is built before hooks).

## Goals / Non-Goals

**Goals:**
- `KnowledgeSaveable()` derived from existing capability fields (no new fields)
- `Catalog.SaveableToolNames()` returns the dynamic list
- `app.go` builder uses catalog-derived list
- `DefaultSaveableTools` retained as fallback for CLI mode (where catalog is unavailable)

**Non-Goals:**
- Adding new fields to `ToolCapability`
- Changing tool registration call sites (they already set `ReadOnly`/`Activity`)

## Decisions

### D1: KnowledgeSaveable() on ToolCapability

**Choice**: `func (c ToolCapability) KnowledgeSaveable() bool` returns true if `c.ReadOnly || c.Activity == ActivityRead || c.Activity == ActivityQuery`. Conservative — only explicitly read/query tools.

### D2: Catalog.SaveableToolNames()

**Choice**: Iterates all registered tools, filters by `KnowledgeSaveable()`, returns sorted `[]string`. Thread-safe (uses existing RLock).

### D3: Dual source — catalog in runtime, fallback in CLI

**Choice**: In `BuildHookRegistry`:
- When `catalog != nil`, use `catalog.SaveableToolNames()`
- When `catalog == nil` (CLI snapshot mode), fall back to `DefaultSaveableTools`

This avoids breaking `lango agent hooks` which calls `BuildHookRegistry(cfg, nil, nil)` without a catalog.

### D4: Update BuildHookRegistry signature

**Choice**: Add optional `catalog *toolcatalog.Catalog` parameter. The private `buildHookRegistry` passes `app.ToolCatalog`. The CLI caller passes `nil`.
