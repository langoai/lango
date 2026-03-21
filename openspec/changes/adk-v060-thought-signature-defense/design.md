## Context

Gemini models use `Thought` and `ThoughtSignature` fields on FunctionCall parts to mark internal reasoning steps. When these thought-tagged calls are persisted in session history and later replayed:
1. Through the **OpenAI provider** — the protocol has no concept of thought calls, causing API errors
2. Through the **Gemini provider** — if `ThoughtSignature` was lost during persistence (`Thought=true, ThoughtSignature=nil`), Gemini rejects the request

The existing defense (`gemini.go:101-106`) drops corrupted thought calls (Thought=true + empty signature) in the Gemini provider. However, no defense exists in the OpenAI provider, and orphaned FunctionResponses left behind by dropped FunctionCalls can also cause errors.

ADK v0.6.0 is a maintenance upgrade (additive: `ModelVersion` field, HTTP header merge, config copy) with no breaking changes for lango's usage surface.

## Goals / Non-Goals

**Goals:**
- Upgrade ADK v0.5.0 → v0.6.0 for maintenance/compatibility
- Prevent thought-tagged FunctionCalls from reaching the OpenAI API
- Remove orphaned FunctionResponses when their FunctionCalls are dropped
- Classify thought_signature errors as model errors to avoid futile learning retries

**Non-Goals:**
- Fixing session persistence to preserve ThoughtSignature (separate concern)
- Adding thought call support to the OpenAI provider
- Modifying the Gemini SDK or ADK internals

## Decisions

### D1: Filter thought calls by `Thought` flag, not by name heuristic

**Decision**: Use `tc.Thought == true` to identify thought calls in the OpenAI provider.

**Rationale**: The `Thought` field is the canonical marker set by Gemini. Name-based heuristics would be fragile and miss future naming changes. The flag is already persisted in `provider.ToolCall`.

**Alternative**: Filter by tool name pattern (e.g., `think`, `thought_*`) — rejected as unreliable.

### D2: Paired deletion (FunctionCall + FunctionResponse)

**Decision**: When a thought FunctionCall is dropped, also drop its corresponding FunctionResponse by tracking IDs in a `droppedThoughtIDs` set.

**Rationale**: An orphaned tool response without its triggering call causes API errors on both OpenAI (`missing tool_call_id reference`) and Gemini (`unmatched FunctionResponse`).

### D3: Orphan cleanup as a separate pass after sanitizeContents()

**Decision**: Add `dropOrphanedFunctionResponses()` as a post-sanitization step in the Gemini provider, rather than embedding the logic in `sanitizeContents()`.

**Rationale**: `sanitizeContents()` handles structural issues (turn ordering, merging). Orphan detection is a semantic concern (which calls survived filtering). Separating them keeps each function focused.

### D4: thought_signature errors classified as ErrModelError

**Decision**: Detect `thought_signature` / `thoughtSignature` in error messages and classify as `ErrModelError` instead of the generic `ErrInternal`.

**Rationale**: Learning-based retry (`errorFixProvider`) is designed for recoverable tool errors. Thought signature errors are server-side rejections that no prompt fix can resolve. Classifying as model error skips the retry path (only `ErrToolError` triggers `GetFixForError`).

## Risks / Trade-offs

- **[Risk] Over-filtering**: Thought calls that should be preserved for Gemini-to-Gemini replay get stripped → **Mitigation**: Only the OpenAI provider filters; Gemini provider only strips when signature is missing (existing behavior). The new `dropOrphanedFunctionResponses` only removes responses whose calls are genuinely absent.
- **[Risk] ADK v0.6.0 subtle behavior change**: New `ModelVersion` field or HTTP header merge could change behavior → **Mitigation**: Changes are additive; full test suite passes. `StartInvokeAgentSpan` signature change verified as unused in lango.
