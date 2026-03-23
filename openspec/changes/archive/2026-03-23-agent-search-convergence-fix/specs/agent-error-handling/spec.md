## MODIFIED Requirements

### Requirement: Error classification priority order

`classifyError()` SHALL check for `thought_signature` / `thoughtSignature` substrings BEFORE checking for `"tool"` / `"function call"` substrings. This ensures that Gemini API errors containing both keywords (e.g., `"Function call is missing a thought_signature in functionCall parts"`) are classified as `ErrModelError`, not `ErrToolError`.

#### Scenario: Gemini thought_signature error classified as model error

- **WHEN** `classifyError` receives an error with message containing both `"function call"` and `"thought_signature"`
- **THEN** it SHALL return `ErrModelError`, not `ErrToolError`

#### Scenario: Pure tool error still classified correctly

- **WHEN** `classifyError` receives an error with message containing `"tool"` but not `"thought_signature"`
- **THEN** it SHALL return `ErrToolError`

### Requirement: Streaming partial events omit thought metadata

Partial tool-call `LLMResponse` events yielded during streaming SHALL NOT carry `Thought` or `ThoughtSignature` fields. Only the final accumulated response (via `toolAccum.done()`) SHALL include the correct `Thought` and `ThoughtSignature` values.

#### Scenario: Partial tool-call event has no thought fields

- **WHEN** `ModelAdapter.GenerateContent` yields a partial event for a tool call with `Name` set
- **THEN** the `genai.Part` SHALL have `Thought=false` and `ThoughtSignature=nil`

#### Scenario: Final accumulated event preserves thought fields

- **WHEN** `ModelAdapter.GenerateContent` yields the final `TurnComplete=true` event
- **THEN** the accumulated `genai.Part` from `toolAccum.done()` SHALL preserve the original `Thought` and `ThoughtSignature` values from the stream
