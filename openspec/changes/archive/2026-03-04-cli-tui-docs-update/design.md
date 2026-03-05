## Context

The lango project has accumulated many internal features (P2P teams, ZKP, learning, agent registry, tool hooks, approval, librarian, agent memory, X402) that have no CLI/TUI exposure. Users cannot discover, inspect, or configure these features without reading source code. The existing CLI follows two established patterns: `bootLoader` for commands needing full bootstrap (DB, crypto, P2P), and `cfgLoader` for config-only commands.

## Goals / Non-Goals

**Goals:**
- Expose all major internal features through CLI subcommands
- Add TUI settings forms for hooks and agent memory configuration
- Add doctor health checks for new subsystems
- Update all documentation (feature docs, CLI reference, README) to reflect new commands
- Follow existing CLI patterns (bootLoader/cfgLoader, Cobra conventions, `--json` flag)

**Non-Goals:**
- Modifying internal feature behavior (pure CLI/TUI/docs surface)
- Adding interactive TUI workflows (forms only)
- Adding E2E or integration tests beyond unit + build verification
- Changing existing command signatures or breaking backward compatibility

## Decisions

### 1. bootLoader vs cfgLoader per command

**Decision**: Use `cfgLoader` for status commands that only read config, `bootLoader` for commands that need DB or runtime state.

| Command | Loader | Rationale |
|---------|--------|-----------|
| `learning status` | cfgLoader | Reads config only |
| `learning history` | bootLoader | Needs DB access for audit logs |
| `p2p team list` | bootLoader | Needs P2P config check (but NOT full P2P node) |
| `p2p zkp circuits` | neither | Static data, no loader needed |
| `graph add/export/import` | cfgLoader | Graph store init from config |
| `approval status` | bootLoader | Reads approval provider state |

**Alternative**: Always use bootLoader for simplicity. Rejected because many commands only need config, and full bootstrap is expensive (DB connection, crypto init).

### 2. Interface extensions for graph and agent memory

**Decision**: Add `AllTriples()` to `graph.Store` and `ListAgentNames()`/`ListAll()` to `agentmemory.Store` interfaces.

**Rationale**: Export and inspection commands need to enumerate all data. These are read-only methods that don't change store semantics. BoltDB implementations scan the SPO bucket (graph) or iterate memory map (agent memory).

**Alternative**: Use raw BoltDB access in CLI. Rejected because it bypasses the store abstraction.

### 3. TUI form organization

**Decision**: Add `forms_hooks.go` with `NewHooksForm()` and `NewAgentMemoryForm()`. Register under "Communication" and "AI & Knowledge" menu categories respectively.

**Rationale**: Follows existing form pattern (`forms_*.go` files, `tuicore.FormModel`, dispatched from `editor.go`). Menu categories match logical grouping.

### 4. Doctor checks as separate files

**Decision**: One file per check in `internal/cli/doctor/checks/` (tool_hooks.go, agent_registry.go, librarian.go, approval.go), registered in `AllChecks()`.

**Rationale**: Follows existing check pattern. Each check implements `Name()`, `Run()`, `Fix()` interface.

### 5. Documentation structure

**Decision**: New feature docs in `docs/features/` (agent-format.md, learning.md, zkp.md), security docs in `docs/security/` (approval-cli.md), CLI reference in `docs/cli/`.

**Rationale**: Matches existing directory structure. CLI docs reference command usage; feature docs explain concepts.

## Risks / Trade-offs

- **[Static P2P messages]** P2P team commands show informational messages rather than live data when P2P node is not running → Acceptable for v1; runtime integration can be added later
- **[Interface additions]** Adding methods to Store interfaces requires all implementations to be updated → Only one implementation per interface exists currently (BoltDB, mem_store), so risk is minimal
- **[Bootstrap cost]** Commands using bootLoader incur DB connection overhead → Mitigated by only using bootLoader where truly needed
- **[Large surface area]** 18 new subcommands across 11 groups increases maintenance → All commands follow identical patterns, reducing cognitive overhead
