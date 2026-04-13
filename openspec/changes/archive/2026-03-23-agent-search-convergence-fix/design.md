## Context

Navigator agent fails to converge when performing web searches — it repeats `browser_search` with slight query variations despite having results, consuming the full 5-minute timeout. Prompt-only guidance proved insufficient; runtime enforcement is needed. Separately, `classifyError` ordering bug causes `thought_signature` errors to be misclassified, and partial streaming events leak premature Thought metadata into session history.

## Goals / Non-Goals

**Goals:**
- Enforce a hard 2-search limit per request at the tool level (prompt-independent)
- Provide agent-visible convergence signals (`LimitReached`, `NextStep`) in `SearchResponse`
- Fix `classifyError` so `thought_signature` errors are correctly classified as `ErrModelError`
- Prevent partial streaming events from corrupting session history with premature Thought data

**Non-Goals:**
- Semantic deduplication of search queries (too complex, not needed with hard limit)
- Limiting `browser_navigate` or `browser_extract` calls (not the current loop source)
- Modifying the Gemini provider's existing thought-call drop defense (`gemini.go:102-106`)

## Decisions

### D1: Hard limit at tool level, not prompt-only

**Decision**: Enforce `MaxSearchesPerRequest=2` in `RequestState.RecordSearch()`, checked before executing the search. Return a structured `SearchResponse` with `LimitReached=true` instead of an error.

**Rationale**: Prompt guidance was already present ("reformulate at most once") but the Gemini model ignored it. A runtime limit is the only reliable enforcement mechanism. Returning a structured response (not an error) preserves agent autonomy — the agent can still navigate to result URLs or extract from the current page.

**Alternative**: Semantic similarity blocking (reject queries similar to previous ones). Rejected — too complex, false positives on legitimate reformulations, and the hard limit already caps attempts.

### D2: Pre-check before search execution

**Decision**: Call `RecordSearch` before `Navigate`/`extractSearchResults`, not after. If limit reached, return immediately without performing the search.

**Rationale**: Executing the 3rd search and then discarding results wastes browser resources and time. Pre-checking is more efficient and guarantees the search never executes.

### D3: classifyError ordering — thought_signature before tool/function

**Decision**: Move the `thought_signature`/`thoughtSignature` substring check above the `"tool"`/`"function call"` check in `classifyError()`.

**Rationale**: The actual Gemini error message `"Function call is missing a thought_signature in functionCall parts"` contains both `"function call"` (matches tool error) and `"thought_signature"` (should match model error). The more specific check must come first.

### D4: Strip Thought from partial streaming events

**Decision**: Set `Thought: false, ThoughtSignature: nil` on partial tool-call LLMResponse events yielded during streaming. Only the final `toolAccum.done()` event carries the correct values.

**Rationale**: Partial events may be yielded before `ThoughtSignature` arrives in a later streaming chunk. If the ADK runner stores a partial event with `Thought=true, ThoughtSignature=nil`, subsequent API replays are rejected. The accumulator already tracks the full values correctly.

## Risks / Trade-offs

- **[Risk] Hard limit too aggressive**: A 2-search limit might prevent legitimate multi-query workflows → **Mitigation**: The limit is per-request, not per-session. The agent can still navigate to result URLs and extract content. The structured response guides the agent to alternative actions.
- **[Risk] LimitReached response confuses model**: Model may not understand the structured stop response → **Mitigation**: `NextStep` field gives explicit instructions. The same field on normal responses ("Results found. Do NOT search again.") trains the model to read it.
- **[Risk] classifyError reorder affects other error types**: Moving thought_signature check could mask other errors → **Mitigation**: Only errors containing `thought_signature` or `thoughtSignature` substrings are affected. Pure tool errors without these substrings are unchanged.
