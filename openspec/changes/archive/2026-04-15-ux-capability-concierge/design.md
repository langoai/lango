## Context

Currently, all registered tools are injected into the ADK agent at boot (`wiring.go:569` `adk.NewAgent(..., adkTools, ...)`), and the tool catalog section is baked into `basePrompt` at boot via `builder.Add(buildToolCatalogSection(catalog))` (`wiring.go:317-328`). `ContextAwareModelAdapter.basePrompt` is a fixed string at construction (`context_model.go:83`), so per-turn modifications cannot change the tool advertisement seen by the LLM. Session mode narrowing therefore requires two coordinated moves: (1) take the tool catalog section out of `basePrompt` and regenerate it dynamically per turn, and (2) enforce the mode allowlist at the individual tool handler level (middleware), because ADK dispatches directly to handlers and the dispatcher is not on the hot path for all call styles.

Skills currently surface as full agent tools via `skillToTool()` (`registry.go:208`), baking full definition into every system prompt. `list_skills` returns complete metadata for every active skill. Token cost is shown per turn (`TurnTokenUsageMsg` at `chat.go:283`) but never summed or priced.

## Goals / Non-Goals

**Goals:**
- Selecting `/mode code-review` narrows both the tool advertisement and the enforced tool set
- `list_skills(summary=true)` returns metadata only; `view_skill` loads full definitions on demand
- Estimated USD cost per turn displayed in the chat view; `/cost` prints session cumulative cost
- Mode changes propagate to multi-channel consumers via eventbus

**Non-Goals:**
- Per-user custom mode definitions loaded from disk (extension pack scope, Phase 4)
- Actual cost tracking against provider billing APIs (estimated only, using static price table)
- Removing skills from ADK tool injection entirely — deferred skills still register, just don't appear in prompts
- Mode-based pricing tiers or quota enforcement

## Decisions

### D1: Tool catalog section moves from basePrompt to dynamic per-turn generation
**Decision**: Remove `builder.Add(buildToolCatalogSection(catalog))` from `wiring.go`. Store `*toolcatalog.Catalog` reference on `ContextAwareModelAdapter` via `WithCatalog(*Catalog)`. Generate tool catalog section inside `GenerateContent()` after Phase 2, before section truncation, appending to the per-turn prompt.
**Alternative**: Keep global catalog section in `basePrompt` and append a mode-specific override section afterwards.
**Rationale**: The alternative produces contradictory instructions ("all tools available" + "use only these tools") — LLMs handle this poorly. Moving the entire tool catalog to dynamic keeps the prompt internally consistent. Knowledge/memory sections already follow the same dynamic-append pattern, so this is architecturally consistent.

### D2: Mode is stored on Session, resolved per turn from ctx
**Decision**: Add `Mode string` to `session.Session`. At turn start, the turn runner reads the session's mode and sets it on the context via a new `session.WithMode(ctx, mode)` helper. `GenerateContent()` and tool middleware resolve mode from context, matching the existing `SessionKeyFromContext` pattern.
**Alternative**: Pass mode as a separate parameter through the runner → executor chain.
**Rationale**: Context propagation is how `SessionKey`, `TurnID`, `TurnApprovalState`, and `BrowserRequestState` are already threaded. No new wiring contract, consistent surface.

### D3: Enforcement at toolchain middleware, not dispatcher
**Decision**: New `WithModeAllowlist(modeResolver func(ctx) SessionMode)` middleware wraps individual tool handlers. Before calling the underlying handler, it checks whether the tool name appears in the current session's mode allowlist. If not, it returns a structured error ("tool not available in current mode: <tool> — active mode: <mode>").
**Alternative**: Filter tools at `Dispatcher` or at ADK agent construction.
**Rationale**: ADK receives the full tool set at boot and dispatches directly to registered handlers (`wiring.go:569`) — the dispatcher is not on the direct-call path. Middleware runs for every tool invocation regardless of call style. Matches existing `WithPolicy`/`WithApproval` pattern.

### D4: Built-in modes as config, user-customizable later
**Decision**: Ship three built-in modes as Go values in `internal/config/modes.go`. `Config` gets a `Modes map[string]SessionMode` that merges built-ins with user-defined entries.
**Alternative**: Require every mode to be declared in user config.
**Rationale**: Built-ins give first-run value without configuration. User overrides (by key) or extensions come later without breaking the built-in surface.

### D5: `list_skills` backward compatibility via parameter
**Decision**: Add `summary bool` parameter to `list_skills` (default `false`). Existing behavior preserved for `summary=false`. `summary=true` returns `{name, description, when_to_use}` only.
**Alternative**: Replace `list_skills` with a new `skills_summary` tool.
**Rationale**: Existing consumers (if any) continue to work. The LLM can opt into the token-efficient form once mode-aware prompt guidance suggests it. Avoids deprecation cycle.

### D6: `view_skill` is a new tool, not a parameter to `list_skills`
**Decision**: Add `view_skill(name string, path ...string)` tool. With only `name`, returns the full SKILL.md. With `name` + `path`, returns the referenced supporting file (relative to the skill directory).
**Alternative**: Extend `list_skills` with a `detail string` parameter.
**Rationale**: Semantic clarity — `list_skills` lists, `view_skill` reads. Matches hermes-agent pattern. `path` variant supports progressive disclosure for skills with reference files without requiring a second tool.

### D7: Cost is estimated from static price table, not provider billing
**Decision**: `internal/provider/pricing.go` holds a static map of model IDs to `{InputPerMillion, OutputPerMillion}` USD prices. `EstimatedCostUSD` is computed from `Usage` tokens × price. Unknown models return 0 (hidden in UI) rather than a fallback price.
**Alternative**: Query provider billing APIs for real-time cost.
**Rationale**: Billing APIs require authenticated per-provider integration, introduce latency, and tie cost display to provider availability. Static table is sufficient for "am I about to spend $10?" UX.

### D8: Mode change emits eventbus event, TUI and channels both subscribe
**Decision**: `ModeChangedEvent{SessionKey, OldMode, NewMode}` published when `/mode` or `--mode` changes the session's mode. TUI `chatView.appendSystem()` renders the change; channel adapters render it per their format.
**Alternative**: Direct call from the slash command into `chatView`.
**Rationale**: Multi-channel consistency — a mode change from a Telegram command reaches the TUI view and vice versa without duplicated wiring.

## Risks / Trade-offs

- **[Prompt caching degradation]** Moving tool catalog out of `basePrompt` shortens the stable prefix. → Mitigation: The tool catalog section is typically small (<1KB for 30 tools) compared to the full system prompt + knowledge/memory. Caching was already invalidated per-turn by dynamic sections, so the marginal impact is low.
- **[Mode enforcement gap if middleware is not applied]** If a tool is registered without going through the toolchain builder, the allowlist won't apply. → Mitigation: Document and audit that all tool registration paths (`app/modules.go`, skill registry) feed through the same middleware chain. Add a test that asserts every tool has the `WithModeAllowlist` wrapper.
- **[User confusion if unknown tool fires in mode]** LLM may reference a tool in its reasoning that is blocked by allowlist. → Mitigation: The error message is explicit ("tool X not available in mode Y") and the LLM can recover on the next turn. Prompt guidance in mode `system_hint` steers the LLM toward allowed tools.
- **[Stale pricing]** Model prices change over time. → Mitigation: Unknown/stale entries return 0, hiding cost display gracefully. Pricing table updates are a low-cost follow-up change.
- **[Skill discovery regression]** If `list_skills` consumers rely on full metadata, `summary=true` would surprise them. → Mitigation: Default remains `summary=false`. LLM opts in via mode system_hint.
