## Why

ADK translation layer (~79KB across 7 files in `internal/adk/`) has grown thick with scattered conversion logic and bug fixes spread across multiple boundaries. Four critical bug fixes (FunctionResponse role correction, orphaned tool call delta drop, streaming partial/final deduplication, thought_signature error classification) are embedded inline without clear seam boundaries. ADK v0.6.0 provides plugin callbacks, MCPToolset, and MemoryService that our runner config does not yet pass through. This refactoring consolidates the translation boundary, establishes regression guards, and spikes ADK native surface integration feasibility.

## What Changes

- Extract and consolidate FunctionCall/FunctionResponse conversion logic from two inline locations (`eventToMessage()` and `EventsAdapter.All()` inline block) into shared converter functions
- Add comprehensive golden test suite covering all session round-trip scenarios as regression safety net
- Split `context_model.go` (13.9KB) into three focused files: adapter entry point, retrieval orchestration, prompt assembly
- Restructure `toolCallAccumulator` in `model.go` from OpenAI/Anthropic branching to provider-agnostic state machine
- Spike ADK plugin callback integration by wiring `PluginConfig` into runner.Config
- Spike ADK MCPToolset parity against current `internal/mcp/` adapter with concrete adoption criteria
- Document ADK `Get()` contract deviation (auto-create/renew behavior) with regression tests

## Capabilities

### New Capabilities

- `adk-plugin-spike`: Feasibility assessment of ADK plugin callback integration — parity gap analysis against current toolchain middleware (SecurityFilterHook, learning observation, per-tool vs agent-level scope)
- `adk-mcp-spike`: Feasibility assessment of ADK MCPToolset adoption — parity gap analysis against current MCP adapter (naming contract, approval path, safety metadata, output truncation, event publication)

### Modified Capabilities

- `adk-architecture`: Session event conversion consolidated to single source of truth; runner.Config extended with optional PluginConfig pass-through
- `streaming-tool-call-assembly`: toolCallAccumulator restructured to provider-agnostic state machine while preserving orphaned delta drop behavior

## Impact

- **Code**: `internal/adk/` (all 7 files), `internal/adk/*_test.go` (new golden tests)
- **APIs**: No external API changes. Internal `SessionServiceAdapter` and `ModelAdapter` signatures preserved.
- **Dependencies**: No new dependencies. Uses existing ADK v0.6.0 `plugin`, `tool/mcptoolset`, `memory` packages for spike evaluation only.
- **Systems**: No deployment or infrastructure changes. No CLI/TUI/config changes.
