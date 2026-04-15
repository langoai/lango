## Why

Lango currently exposes every tool and skill to every session regardless of user intent. Users cannot scope the agent to a specific task ("code review", "research", "debug") without manually disabling tools in config, and the LLM sees the entire tool catalog plus every skill description in every system prompt — wasting tokens and inviting off-target tool choices. Token usage is visible per turn but users have no sense of cost. The result is a rigid, opaque UX that forces users to trust the agent to figure out scope on its own.

## What Changes

- **Session Modes (intent profiles)**: Built-in `code-review`, `research`, `debug` modes that automatically narrow tool catalog, skill discovery, and system prompt hints. Selected via `/mode <name>` slash command or `--mode` CLI flag.
- **Tool catalog becomes fully dynamic**: The global tool catalog section is removed from `basePrompt` (`wiring.go:317-328`) and regenerated per-turn inside `ContextAwareModelAdapter.GenerateContent()`. When a session has a mode, only mode-filtered tools appear in the system prompt. Eliminates the "prompt says all tools available but mode limits execution" inconsistency.
- **Middleware-level enforcement**: `WithModeAllowlist` middleware in the `toolchain` chain blocks handler execution when the tool is outside the current session mode — ADK receives all tools at boot (`wiring.go:569`), so enforcement must live next to individual handlers, not at dispatcher.
- **Capability discovery via progressive disclosure**: `list_skills` gains an optional `summary=true` parameter returning only `{name, description, when_to_use}` metadata. New `view_skill(name)` tool loads full skill definition on demand. Instruction/template skills transition to `ExposureDeferred`; script/fork skills remain directly invocable.
- **Cost visibility**: Model price table added (`internal/provider/pricing.go`). `TurnTokenUsageMsg` + eventbus token usage event extended with `EstimatedCostUSD`. New `/cost` slash command summarizes session cumulative cost.
- **Multi-channel consistency**: Mode changes emit `ModeChangedEvent` on the eventbus so TUI and messaging channels render consistently.

## Capabilities

### New Capabilities
- `session-modes`: Intent profiles that scope tools, skills, and prompt hints per session with dispatch-level allowlist enforcement
- `capability-discovery`: Metadata-first skill surfacing with on-demand `view_skill` loading
- `turn-cost-visibility`: Per-turn estimated cost display and session cumulative cost summary

### Modified Capabilities
- `interactive-tui-chat`: New `/mode` and `/cost` slash commands; `--mode` CLI flag for bare `lango` invocation
- `context-budget`: Tool catalog section moves from boot-time `basePrompt` to per-turn dynamic generation (spec note only — budget math unchanged)

## Impact

- **Code**: `internal/config/types.go` (SessionMode struct), `internal/session/` (Mode persistence), `internal/toolcatalog/catalog.go` (`ListVisibleToolsForMode`), `internal/toolchain/` (`WithModeAllowlist` middleware), `internal/adk/context_model.go` (`WithCatalog`, dynamic tool catalog section), `internal/app/wiring.go` (remove static tool catalog section, wire catalog into adapter), `internal/app/tools_meta.go` (`list_skills` param, `view_skill` tool), `internal/skill/registry.go` (deferred exposure transition), `internal/provider/pricing.go` (new), `internal/cli/chat/commands.go` (`/mode`, `/cost`), `cmd/lango/main.go` (`--mode` flag), eventbus (`ModeChangedEvent`, extended token event).
- **Interfaces**: `SessionMode` struct, `Catalog.ListVisibleToolsForMode`, `ContextAwareModelAdapter.WithCatalog`. No breaking changes — `list_skills` default remains `summary=false`.
- **Dependencies**: No new external dependencies.
- **Risk**: Moving the tool catalog out of `basePrompt` affects prompt caching (boot-time stable prefix gets shorter, per-turn dynamic suffix grows). Mitigated because knowledge/memory sections already use the same dynamic-append pattern. Mode enforcement at middleware level is additive to existing `WithPolicy`/`WithApproval` layers.
