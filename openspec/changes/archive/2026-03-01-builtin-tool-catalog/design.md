## Context

The current multi-agent orchestration assigns tools to sub-agents via prefix-based partitioning (`PartitionTools`). The orchestrator itself has no direct tools — it delegates everything. When a sub-agent rejects a task or a tool has no matching prefix, the orchestrator cannot execute it. Additionally, disabled features' tools are invisible at runtime.

Separately, `SkillConfig.AllowImport` exists but is never checked in the `import_skill` handler.

## Goals / Non-Goals

**Goals:**
- Enforce AllowImport flag so administrators can disable skill imports
- Provide a catalog of all initialized built-in tools for runtime discovery
- Give the orchestrator direct tool access via dispatcher tools (builtin_list/builtin_invoke)
- Keep dispatcher tools orchestrator-exclusive to preserve role separation

**Non-Goals:**
- Dynamic tool loading/unloading at runtime (tools are registered once at startup)
- Per-user or per-session tool visibility (catalog is global)
- Replacing PartitionTools — sub-agents still get their domain tools directly

## Decisions

### 1. Tool Catalog as a separate package (`internal/toolcatalog/`)
**Decision**: New standalone package rather than embedding in `internal/app/`.
**Rationale**: Avoids import cycles — orchestration package can import toolcatalog without depending on app. Clean separation of concerns: registration (catalog) vs wiring (app).
**Alternatives**: Embedding in `internal/agent/` (rejected: agent package is domain types, not infrastructure); inline in app.go (rejected: grows an already large file).

### 2. Two dispatcher tools instead of one
**Decision**: `builtin_list` (Safe) + `builtin_invoke` (Dangerous) as separate tools.
**Rationale**: Separation of discovery from execution allows the LLM to browse available tools without triggering approval gates. `builtin_invoke` is Dangerous because it proxies arbitrary tool execution.
**Alternatives**: Single `builtin_dispatch` with action parameter (rejected: conflates safe listing with dangerous invocation, complicates approval policy).

### 3. Orchestrator-only dispatcher scope
**Decision**: Only the orchestrator receives builtin_list/builtin_invoke. Sub-agents use their directly-assigned tools.
**Rationale**: Preserves role separation, saves sub-agent context window, maintains consistent approval at orchestrator level. Sub-agents reject out-of-scope requests via `[REJECT]` protocol, orchestrator handles fallback.
**Alternatives**: Give every agent dispatchers (rejected: defeats purpose of role separation, bloats context).

### 4. Category-based registration at wiring time
**Decision**: Each `buildXxxTools()` result is registered under a named category immediately after creation, before approval wrapping.
**Rationale**: Registration before approval wrapping means catalog stores the raw tools. The dispatcher's `builtin_invoke` calls the raw handler — approval is applied at the tool level via the approval middleware already wrapping the dispatcher tool itself (since it's SafetyLevelDangerous).

## Risks / Trade-offs

- **[Double approval]** builtin_invoke is Dangerous and triggers approval, then the proxied tool might also have approval wrapping → Mitigation: catalog stores pre-approval tools, so only the dispatcher's approval gate fires.
- **[Catalog size]** With 50+ tools registered, builtin_list output could be long → Mitigation: category filter parameter lets the LLM narrow results.
- **[No hot-reload]** Tools registered at startup only; new features require restart → Acceptable: matches current architecture where all components initialize in `app.New()`.
