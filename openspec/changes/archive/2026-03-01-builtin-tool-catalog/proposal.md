## Why

Agents cannot dynamically discover or invoke built-in tools outside their assigned domain. In multi-agent mode, the orchestrator has no tools and must delegate everything to sub-agents, but when a sub-agent rejects a task or a tool falls outside any agent's prefix mapping, the request fails silently. Additionally, the `AllowImport` flag on `SkillConfig` is defined but never enforced in the `import_skill` handler, creating a security gap.

## What Changes

- Enforce `SkillConfig.AllowImport` flag in the `import_skill` tool handler — reject imports when disabled
- Introduce `internal/toolcatalog/` package with a thread-safe `Catalog` type that registers all initialized built-in tools by category
- Add `builtin_list` (discovery) and `builtin_invoke` (proxy execution) dispatcher tools via `BuildDispatcher()`
- Wire catalog registration into `app.New()` for all 14+ tool categories (exec, filesystem, browser, crypto, secrets, meta, graph, rag, memory, payment, p2p, librarian, cron, background, workflow)
- Add `UniversalTools` field to `orchestration.Config` so the orchestrator agent receives dispatcher tools directly
- Skip `builtin_` prefixed tools in `PartitionTools` so they remain orchestrator-exclusive
- Update orchestrator instruction prompt to reflect direct tool access capability
- Update `blockLangoExec` catch-all message to hint at `builtin_list`

## Capabilities

### New Capabilities
- `tool-catalog`: Thread-safe registry for built-in tools with category grouping, discovery, and dynamic dispatch

### Modified Capabilities
- `meta-tools`: AllowImport guard enforcement on import_skill handler
- `multi-agent-orchestration`: UniversalTools support for orchestrator, builtin_ prefix exclusion from sub-agent partitioning
- `tool-exec`: blockLangoExec message updated with builtin_list hint

## Impact

- **New package**: `internal/toolcatalog/` (catalog.go, dispatcher.go, tests)
- **Modified files**: `internal/app/app.go`, `internal/app/types.go`, `internal/app/wiring.go`, `internal/app/tools.go`, `internal/app/tools_meta.go`, `internal/orchestration/orchestrator.go`, `internal/orchestration/tools.go`, `internal/orchestration/orchestrator_test.go`
- **No breaking changes**: All existing tools continue to work; catalog is additive
- **No new dependencies**: Uses only stdlib (`sync`, `sort`, `context`, `fmt`)
