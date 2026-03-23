## Why

The navigator agent loops indefinitely on `browser_search` calls without converging to a response. Existing prompt guidance ("prefer extract before new search", "reformulate at most once") and output signals (`resultCount`, `empty`, `pageType`) are insufficient — the Gemini model ignores soft prompt language and repeats searches with slight query variations until hitting the 5-minute idle timeout (observed: 171KB partial result, zero user-visible output). Additionally, `classifyError` misclassifies `thought_signature` Gemini API errors as tool errors, triggering futile learning retries, and streaming partial events leak premature `Thought=true` with empty `ThoughtSignature` into session history, corrupting subsequent API replays.

## What Changes

- Add runtime hard limit on `browser_search` calls per request (`MaxSearchesPerRequest=2`); 3rd attempt returns a structured stop response with `LimitReached=true` and `NextStep` advisory instead of executing the search
- Enrich `SearchResponse` with `LimitReached`, `NextStep`, and `Warning` fields for agent-visible convergence signals
- Strengthen navigator prompts from soft language ("Prefer", "before considering") to imperative mandates ("ONCE", "Do NOT search again", "NEVER more than twice")
- Reorder `classifyError()` to check `thought_signature` before `tool`/`function call` keywords, fixing misclassification as `ErrToolError`
- Strip `Thought`/`ThoughtSignature` from partial streaming tool-call events to prevent session corruption

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tool-browser`: Add `MaxSearchesPerRequest` hard limit and `LimitReached`/`NextStep`/`Warning` fields to `SearchResponse`
- `agent-prompting`: Navigator search workflow changed from soft guidance to mandatory bounded protocol
- `agent-error-handling`: `classifyError` reordered so `thought_signature` errors are classified as `ErrModelError` before `tool`/`function call` keyword match

## Impact

- `internal/tools/browser/request_state.go` — new constant, expanded `RecordSearch` return signature
- `internal/tools/browser/high_level.go` — `SearchResponse` struct fields, `Search()` pre-check logic
- `internal/adk/errors.go` — `classifyError` check ordering
- `internal/adk/model.go` — partial streaming event sanitization
- `prompts/agents/navigator/IDENTITY.md` — imperative search workflow
- `prompts/TOOL_USAGE.md` — imperative browser search guidance
- `internal/agentregistry/defaults/navigator/AGENT.md` — synchronized with IDENTITY.md
