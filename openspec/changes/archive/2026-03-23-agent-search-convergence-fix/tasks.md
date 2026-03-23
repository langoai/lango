## Tasks

### Track B: Search Convergence (service stabilization)

- [x] 1.1 Add `MaxSearchesPerRequest = 2` constant to `internal/tools/browser/request_state.go`
- [x] 1.2 Expand `RecordSearch()` return to `(count, queries, shouldWarn, limitReached)` with `limitReached = count > MaxSearchesPerRequest`
- [x] 1.3 Preserve `currentURL` when `RecordSearch` receives empty string
- [x] 1.4 Add `LimitReached`, `NextStep`, `Warning` fields to `SearchResponse` struct in `internal/tools/browser/high_level.go`
- [x] 1.5 Add pre-check in `Search()` to call `RecordSearch` before executing search; return structured stop response when `limitReached`
- [x] 1.6 Set `NextStep` advisory on normal search results based on `resultCount`
- [x] 1.7 Update `request_state_test.go` for new `RecordSearch` signature and limit behavior
- [x] 2.1 Replace soft language in `prompts/agents/navigator/IDENTITY.md` with imperative search workflow
- [x] 2.2 Synchronize `internal/agentregistry/defaults/navigator/AGENT.md` with IDENTITY.md search workflow
- [x] 2.3 Replace soft language in `prompts/TOOL_USAGE.md` browser section with imperative search guidance

### Track A: ADK/Gemini stabilization

- [x] 3.1 Reorder `classifyError()` in `internal/adk/errors.go` to check `thought_signature` before `tool`/`function call`
- [x] 3.2 Add Gemini API error format test case to `internal/adk/errors_test.go`
- [x] 3.3 Strip `Thought`/`ThoughtSignature` from partial streaming tool-call events in `internal/adk/model.go`
