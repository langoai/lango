## Context

When using OpenAI-compatible providers in streaming mode, tool call responses arrive as multiple delta chunks. The first chunk carries `Index`, `ID`, and `Name`; subsequent chunks carry only `Index` and partial `Arguments`. The current code stores each delta as a separate `FunctionCall` part, resulting in parts with empty `Name` fields being persisted in the session. When these are replayed on the next turn, OpenAI rejects them with HTTP 400.

Anthropic uses a different pattern: `content_block_start` carries `ID`+`Name`, followed by `input_json_delta` with only `Arguments` (no Index field).

## Goals / Non-Goals

**Goals:**
- Correctly assemble streaming tool call deltas into complete FunctionCall parts
- Support both OpenAI (Index-based) and Anthropic (ID/Name start + orphan delta) patterns
- Prevent empty-name tool calls from reaching the OpenAI API at multiple layers
- Preserve existing non-streaming behavior

**Non-Goals:**
- Changing the session storage format or migration of existing sessions
- Supporting other streaming patterns beyond OpenAI and Anthropic
- Modifying the Anthropic or Gemini provider implementations

## Decisions

### 1. Accumulator pattern over in-place mutation

Introduce a `toolCallAccumulator` type that collects deltas by index and emits complete parts on `done()`. This is cleaner than mutating a growing slice of parts in-place.

**Alternative**: Mutate last part in `toolParts` slice — rejected because it requires checking Name emptiness on every append and doesn't handle interleaved multi-call streams.

### 2. Fallback chain for entry resolution

The accumulator resolves which entry a delta belongs to via: `Index` (OpenAI) → `ID`/`Name` presence (Anthropic start, assigns synthetic index) → `lastIndex` (Anthropic delta). This single code path handles both providers without branching on provider type.

### 3. Defense-in-depth filtering

Even with correct accumulation, add empty-name guards at three layers:
- `convertMessages()` — skip FunctionCall parts with empty Name
- `convertTools()` — skip FunctionDeclarations with empty Name
- OpenAI `convertParams()` — filter tool definitions and tool calls with empty Name

This ensures resilience against any upstream bug that might produce empty names.

### 4. Partial yield only for chunks with Name

During streaming, yield partial tool call responses to the UI only when the delta carries a Name (i.e., the first chunk that identifies the tool). Subsequent arg-only deltas are accumulated silently. This prevents the ADK runner from storing incomplete FunctionCall parts.

## Risks / Trade-offs

- [Partial UI latency] Tool call UI notification is delayed until the first chunk with Name arrives → Acceptable since Name always comes in the first chunk for both OpenAI and Anthropic.
- [Memory for accumulator map] Accumulator holds entries in memory until `done()` → Negligible; tool calls per turn are typically < 10.
- [Orphan delta logging] Orphan deltas (no preceding start) are logged and dropped → Correct behavior; if this happens frequently it indicates a provider bug.
