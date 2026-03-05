## Context

Gemini 3+ models use a "thinking" mechanism where FunctionCall parts carry `Thought bool` and `ThoughtSignature []byte` metadata. The API requires these fields echoed back on subsequent FunctionCall parts in conversation history. Our provider abstraction layer strips these fields during the round-trip: response parsing → provider.ToolCall → session persistence → history reconstruction → genai.Part. This causes HTTP 400 errors for all multi-turn tool-calling conversations.

## Goals / Non-Goals

**Goals:**
- Preserve `ThoughtSignature` and `Thought` through the entire data flow: streaming response → provider.ToolCall → session.ToolCall → ent schema → session history replay → genai.Part
- Maintain backward compatibility with existing sessions (no migration needed)
- Keep non-Gemini providers unaffected (zero-valued fields ignored)

**Non-Goals:**
- Interpreting or modifying ThoughtSignature content (opaque passthrough only)
- Adding thinking/reasoning UI indicators
- Changing the provider interface contract beyond additive fields

## Decisions

**Decision 1: Add fields to all 3 ToolCall layers**

Add `Thought bool` and `ThoughtSignature []byte` to `provider.ToolCall`, `session.ToolCall`, and `entschema.ToolCall`. This ensures the data survives the full round-trip without special-casing.

*Alternative*: Store ThoughtSignature only in-memory on genai.Part pointers. Rejected because session persistence (DB restart) would lose the data.

**Decision 2: Use `omitempty` JSON tags for backward compatibility**

Existing sessions in the database have no `thought`/`thoughtSignature` JSON keys. Using `omitempty` ensures deserialization produces zero values cleanly without migration.

*Alternative*: Database migration to add columns. Rejected because ToolCall is stored as a JSON blob inside the `tool_calls` column — no schema change needed.

**Decision 3: Upgrade ADK v0.4.0 → v0.5.0**

ADK v0.5.0 has no breaking API changes in the surfaces we use. The upgrade ensures compatibility with latest genai types.

## Risks / Trade-offs

- **[Risk] ThoughtSignature size in DB** → The field is typically small (<1KB). JSON blob storage handles it naturally. No concern at current scale.
- **[Risk] Parallel function calls** → Only the first FunctionCall part in a response carries ThoughtSignature. Per-part preservation in the ToolCall array handles this correctly.
- **[Trade-off] Fields added to non-Gemini providers** → Zero-valued fields add negligible serialization overhead. Accepted for simplicity over provider-specific branching.
