## Context

The current prompt guidance is directionally correct: use higher-level browser tools first. The gap is that it does not define a degradation path when `browser_search` is absent in the live runtime. In practice that leads the agent to report the missing tool instead of continuing with the tools it still has.

## Goals / Non-Goals

**Goals:**
- Keep the high-level tool preference intact.
- Add an explicit fallback sequence that uses already-supported browser tools.
- Ensure the navigator keeps working when only a subset of browser tools is available at runtime.

**Non-Goals:**
- Adding new browser runtime behavior
- Changing tool registration or routing logic
- Introducing new search providers or external dependencies

## Decisions

### D1: Prefer higher-level tools, but define a strict downgrade path

The prompt will explicitly direct the navigator to:
1. use `browser_search` when available,
2. otherwise navigate directly to a search URL with `browser_navigate`,
3. then extract results with `browser_extract(mode="search_results")`,
4. and only then fall back to low-level `browser_action` / `eval`.

This keeps the preference order clear while avoiding dead-ends.

### D2: Put fallback guidance in both shared tool guidance and navigator identity

The fallback chain will live in both:
- `TOOL_USAGE.md` for general agent behavior
- navigator-specific prompts for sub-agent behavior

This avoids relying on only one prompt layer.

## Risks / Trade-offs

- [Prompt gets slightly longer] → Mitigation: keep the fallback chain short and procedural.
- [Model may still overfit on the preferred tool] → Mitigation: explicitly state "do not stop if equivalent lower-level browser tools are still available."

## Migration Plan

1. Update prompt files with fallback instructions.
2. Update related specs.
3. Run verification.
4. Sync/archive the OpenSpec change.
