## Why

Gemini's `thought_signature` feature causes runtime errors when thought-tagged FunctionCalls are replayed through providers that don't support them (OpenAI), or when the signature is lost during session persistence. ADK v0.6.0 brings compatibility improvements and the existing defense in `gemini.go` needs complementary filtering in the OpenAI provider and error classification layers.

## What Changes

- Upgrade ADK dependency from v0.5.0 to v0.6.0 (additive API changes, no breaking changes for lango)
- Filter Gemini thought tool calls (`Thought=true`) and their corresponding tool responses in the OpenAI provider's `convertParams()`, since OpenAI protocol cannot represent thought metadata
- Remove orphaned `FunctionResponse` parts in Gemini's content pipeline when the matching `FunctionCall` was dropped by thought filtering
- Classify `thought_signature` errors as model errors (not tool errors) to prevent futile learning-based retries

## Capabilities

### New Capabilities

- `thought-call-filtering`: Cross-provider defense that strips Gemini thought tool calls and their responses when routing through providers that cannot represent them

### Modified Capabilities

- `agent-error-handling`: Add thought_signature error classification to skip learning-based retry
- `gemini-content-sanitization`: Extend sanitization to drop orphaned FunctionResponses after thought call filtering

## Impact

- `go.mod` / `go.sum`: ADK v0.5.0 → v0.6.0, golang.org/x/sys minor bump
- `internal/provider/openai/openai.go`: thought call + response filtering in `convertParams()`
- `internal/provider/gemini/sanitize.go`: new `dropOrphanedFunctionResponses()` function
- `internal/provider/gemini/gemini.go`: call `dropOrphanedFunctionResponses()` after `sanitizeContents()`
- `internal/adk/errors.go`: thought_signature pattern in `classifyError()`
- Test files updated in all three packages
