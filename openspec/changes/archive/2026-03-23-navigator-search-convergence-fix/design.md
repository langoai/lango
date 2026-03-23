## Context

The approval replay guard fixed duplicate prompts for identical `tool + params`, but the navigator still burns turns by repeatedly issuing `browser_search` with slightly different queries. That behavior is now the dominant failure mode.

The browser tool layer also under-specifies page semantics:
- search results do not clearly say how many useful results exist
- snapshots do not clearly distinguish result pages from generic pages
- article extraction does not explicitly indicate article mode or emptiness

## Goals / Non-Goals

**Goals:**
- Make search workflows converge using current page context before new search.
- Give the model explicit result-count and page-type signals.
- Add diagnostics for repeated browser search churn inside one request.

**Non-Goals:**
- Changing browser approval policy
- Blocking semantically different searches at runtime
- Reworking child-session summary or cross-turn memory architecture

## Decisions

### D1: Prompt-level bounded search, not runtime semantic blocking

We will keep runtime replay protection exact-match only. Search reformulation budget is enforced in prompt guidance, not by semantic runtime blocking.

### D2: Page outputs must advertise convergence state

`browser_search`, `browser_extract(search_results)`, and `browser_navigate` will expose:
- page type
- result count
- empty signal

This is enough for the model to decide whether it should stop, extract, or reformulate.

### D3: Request-local diagnostics only

We will log search churn when a request performs 3 or more searches, but we will not hard-block it. The diagnostics are for operator visibility and later iteration.

## Risks / Trade-offs

- [Prompt guidance may still be ignored] → Mitigation: richer output shape plus churn logs makes failures diagnosable.
- [Generic pages might be misclassified as search results] → Mitigation: use result-card selectors first, not arbitrary visible links.
- [No runtime semantic cap] → Mitigation: accepted by plan; exact-match replay guard remains in place.
