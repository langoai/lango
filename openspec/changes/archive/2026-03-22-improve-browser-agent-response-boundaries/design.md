## Context

The current browser stack exposes only low-level primitives (`browser_navigate`, `browser_action`, `browser_screenshot`). That works for deterministic flows, but it forces the model to spend many turns on search-result discovery, element inspection, and page extraction. At the same time, the response collection path currently aggregates text from the agent run too loosely: partial drafts and delegated agent chatter can be accumulated and surfaced when a run ends in timeout or turn-limit failure.

This change crosses several layers:
- ADK runtime turn budgeting
- Browser tool design and orchestration prompts
- User-facing error formatting in channels and gateway
- Session timeout annotations
- CLI/docs defaults

## Goals / Non-Goals

**Goals:**
- Make the default turn budget materially less fragile for real tool workflows.
- Give the browser agent higher-level primitives that reduce turn count for search and extraction tasks.
- Ensure user-visible responses stay bounded to final/safe output rather than raw internal drafts.
- Keep browser primitives backward-compatible enough that existing low-level flows still work.

**Non-Goals:**
- Introducing external search APIs or paid search providers
- Replacing go-rod with a different browser engine
- Adding a second LLM inside browser tools for semantic action planning
- Solving every possible prompt leak purely with regex filtering

## Decisions

### D1: Raise defaults to 50 single-agent / 75 multi-agent

The baseline default turn limit will move from 25 to 50. The implicit multi-agent default will move from 50 to 75.

Rationale:
- 25 is too small for real browser-assisted workflows.
- Multi-agent still needs extra headroom over the single-agent baseline because orchestration adds additional tool turns and retry surfaces.

Alternative considered:
- Raising only the multi-agent default. Rejected because single-agent tool chains also hit the same ceiling.

### D2: Add higher-level browser primitives instead of replacing low-level ones

The browser toolset will keep `browser_action` and `browser_screenshot`, but add higher-level tools for:
- browser-native web search
- actionable element observation
- structured extraction

`browser_navigate` will also return a richer snapshot rather than only title/url/snippet.

Rationale:
- Existing low-level primitives remain useful for deterministic flows and backward compatibility.
- Higher-level primitives cut down model tool loops for search-heavy tasks.
- This matches the pattern used by modern browser-agent systems: separate search/observe/extract from low-level click/type primitives.

Alternative considered:
- Replacing `browser_action` entirely with one generic agent-browser tool. Rejected because it would be harder to test deterministically and would collapse too many behaviors into one opaque surface.

### D3: Treat isolated sub-agent text as internal for user-visible output

When collecting or streaming text for the end user, text authored by isolated sub-agents will be ignored. Only root/non-isolated user-visible text will be collected.

Rationale:
- Isolated child-session traffic is implementation detail, not user-facing content.
- This reduces leakage of delegated agent chatter even before error handling kicks in.

Alternative considered:
- Leaving collection unchanged and relying only on prompt wording. Rejected because it is too brittle.

### D4: Preserve partial drafts for diagnostics, not for user delivery or session history

`AgentError.Partial` remains available internally for logs/diagnostics, but channels/gateway will not echo it back to the user and timeout annotations will not persist it into session history.

Rationale:
- Raw partial drafts are often incomplete or internal.
- Persisting them into history creates follow-on leakage risk in later turns.

Alternative considered:
- Sanitizing and still displaying/storing the partial. Rejected for now because the safer default is non-disclosure.

### D5: Strengthen prompt-level response boundary instructions

Prompt output principles will explicitly forbid role-labeled dumps such as system prompt, user prompt, assistant response, tool output, or orchestration traces in final user-visible replies.

Rationale:
- Prompt guidance is still the first line of defense.
- This complements, but does not replace, runtime filtering and collection boundaries.

## Risks / Trade-offs

- [Search engine DOM drift] → Mitigation: use structured extraction heuristics with fallback link extraction; keep low-level tools available.
- [Higher turn defaults increase cost] → Mitigation: only defaults change; explicit config still overrides.
- [Filtering isolated-agent text could hide a response if orchestration fails to synthesize] → Mitigation: keep empty-response fallback and add stronger navigator/orchestrator guidance.
- [Removing partial persistence reduces debugging detail in chat history] → Mitigation: keep partial on `AgentError` and logs, but not in user/session surfaces.

## Migration Plan

1. Raise defaults in runtime, config-facing status output, and docs.
2. Add browser search/observe/extract helpers and richer navigation snapshots.
3. Change collection/error paths to suppress raw partial delivery and raw timeout annotation persistence.
4. Update prompts, README, and docs.
5. Run build/tests, then sync/archive the OpenSpec change.

## Open Questions

- None for this iteration. Search remains browser-native and does not introduce external provider dependencies.
