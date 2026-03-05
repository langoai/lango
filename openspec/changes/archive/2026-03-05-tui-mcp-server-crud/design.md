## Context

The TUI Settings editor uses a list+form CRUD pattern for managing map-based configurations (Providers, Auth Providers). MCP servers (`cfg.MCP.Servers`) follow the same map structure but lack a TUI CRUD interface—only CLI commands (`lango mcp add/remove`) are available.

The existing patterns are well-established: `ProvidersListModel` for list navigation, `NewProviderForm()` for item forms, `UpdateProviderFromForm()` for state persistence, and `StepProvidersList` for editor step wiring.

## Goals / Non-Goals

**Goals:**
- Replicate the Providers list+form CRUD pattern for MCP servers
- Support transport-conditional field visibility (stdio vs http/sse)
- Serialize map/slice fields (env, headers, args) as CSV in form text inputs
- Maintain consistent UX with existing list+form patterns

**Non-Goals:**
- Live MCP server connectivity testing from the form
- Drag-and-drop reordering of servers
- Inline editing without the form step

## Decisions

1. **Reuse existing list+form pattern** — Direct structural copy of `ProvidersListModel`/`NewProviderForm`/`UpdateProviderFromForm` applied to MCP servers. This ensures consistency and minimal learning curve.
   - Alternative: Generic list model with type parameters — rejected due to Go generics complexity and premature abstraction for only 3 list models.

2. **Transport-conditional fields via `VisibleWhen`** — stdio fields (command, args) and http/sse fields (url, headers) use `VisibleWhen` closures referencing the transport field's value. This leverages existing `Field.VisibleWhen` infrastructure.
   - Alternative: Separate forms per transport — rejected due to duplication of shared fields (enabled, env, timeout, safety).

3. **`KEY=VAL,KEY=VAL` serialization for maps** — Env and Headers maps serialize to/from comma-separated key=value pairs. A new `parseKeyValuePairs()` helper uses `=` as the delimiter (vs `:` in `parseCustomPatterns`).
   - Alternative: JSON text area — rejected as too complex for the TUI text input model.

4. **Separate menu entries** — "MCP Settings" (global config form) and "MCP Server List" (CRUD list). This avoids overloading a single menu item and follows the Providers pattern where global settings and per-item CRUD are separate concerns.

## Risks / Trade-offs

- **[Commas in values]** → The CSV serialization breaks if env values or headers contain commas. Mitigation: documented limitation; users with complex values should use CLI or config file directly.
- **[Form title collision detection]** → `isMCPServerForm()` checks `strings.Contains(title, "MCP Server")` which must not collide with the global "MCP Servers Configuration" form. Mitigation: global form uses "MCP Servers Configuration", server forms use "MCP Server:" or "MCP Server" prefix — the `isProviderForm` already excludes "OIDC" with the same pattern.
