## Context

The ADK library (`base_flow.go:577`) stores all FunctionResponse events with `Content.Role = "user"`. Our `EventsAdapter.All()` uses the role to determine how to reconstruct genai parts — messages with role `"user"` are emitted as plain text, not as FunctionResponse parts. This causes tool responses to vanish from reconstructed session history.

When the LLM retries after a hallucination or error, the orphaned FunctionCall (with no matching tool response) triggers OpenAI's validation: `"No tool output found for function call"` (HTTP 400).

This is a separate bug from the previously fixed "empty tool call name" issue.

## Goals / Non-Goals

**Goals:**
- Ensure FunctionResponse events are stored with correct role (`"tool"`) regardless of ADK behavior
- Automatically correct legacy data at read-time without requiring database migration
- Provide defense-in-depth at the provider boundary for any remaining edge cases

**Non-Goals:**
- Patching the upstream ADK library
- Changing the session persistence schema
- Handling FunctionResponse events that lack ToolCall metadata entirely (already handled by existing legacy fallback)

## Decisions

### Defense-in-depth with 3 layers

**Decision**: Fix at write-time (Layer 1), read-time (Layer 2), and provider boundary (Layer 3).

**Rationale**: A single fix point would leave existing data broken (write-only) or silently correct data without preventing future corruption (read-only). The provider boundary layer catches any edge case where Layers 1+2 fail. Each layer is independently testable.

**Alternatives considered**:
- Write-time only: Would not fix existing data in production databases.
- Database migration: Risky, requires downtime, and doesn't prevent future ADK updates from re-introducing the issue.

### Role correction criteria

**Decision**: Only correct role when message has FunctionResponse ToolCalls (Output != "") AND no FunctionCall ToolCalls (Input != ""). Mixed messages are left unchanged.

**Rationale**: A message with both FunctionCall and FunctionResponse data could be a valid edge case that shouldn't be silently modified. The condition ensures only pure FunctionResponse messages are corrected.

### Synthetic response injection scope

**Decision**: Only inject synthetic error responses when an orphaned FunctionCall is followed by a user message. Never touch pending calls at the end of history.

**Rationale**: Pending calls at the end of history represent in-flight tool execution — injecting a synthetic response would prevent the real response from being processed. The "followed by user message" condition ensures we only repair truly orphaned calls from past turns.

## Risks / Trade-offs

- **[Risk]** ADK upstream changes role assignment logic → **Mitigation**: Layer 1's condition is narrow (only corrects "user" → "tool" for FunctionResponse-only messages), so a fix upstream would simply make Layer 1 a no-op.
- **[Risk]** Synthetic response injection masks real data loss → **Mitigation**: WARN-level log on every injection; condition is very strict (only when followed by user message); content clearly states "interrupted".
- **[Trade-off]** Read-time correction adds per-message overhead → Acceptable: the check is O(n) over ToolCalls per message, which is typically 1-3 items.
