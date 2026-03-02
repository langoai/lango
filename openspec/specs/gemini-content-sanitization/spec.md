## ADDED Requirements

### Requirement: Gemini content turn-order sanitization pipeline
The Gemini provider SHALL sanitize the content sequence before every API call to satisfy Gemini's strict turn-ordering rules. The sanitization pipeline SHALL execute 5 steps in order: (1) drop leading orphaned FunctionResponses, (2) merge consecutive same-role contents, (3) prepend synthetic user turn if sequence starts with model, (4) ensure FunctionCall/FunctionResponse pairing, (5) final merge pass.

#### Scenario: Consecutive same-role contents merged
- **WHEN** the content sequence contains consecutive entries with the same role (e.g., model, model)
- **THEN** the sanitizer SHALL merge their Parts into a single Content entry with that role

#### Scenario: Sequence starting with model turn
- **WHEN** the first content entry has role "model"
- **THEN** the sanitizer SHALL prepend a synthetic user Content with text "[continue]"

#### Scenario: Orphaned FunctionResponse at start of sequence
- **WHEN** the content sequence starts with user-role entries containing only FunctionResponse parts (no preceding FunctionCall)
- **THEN** the sanitizer SHALL drop those entries

#### Scenario: FunctionCall without matching FunctionResponse
- **WHEN** a model Content contains FunctionCall parts and the next Content is not a user FunctionResponse
- **THEN** the sanitizer SHALL insert a synthetic user Content with FunctionResponse parts (status "[no response available]") for each FunctionCall

#### Scenario: Valid FunctionCall/FunctionResponse pair preserved
- **WHEN** a model Content with FunctionCall is immediately followed by a user Content with matching FunctionResponse
- **THEN** the sanitizer SHALL pass the pair through unchanged

#### Scenario: Empty content sequence
- **WHEN** the content sequence is empty
- **THEN** the sanitizer SHALL return the empty sequence unchanged

### Requirement: Consecutive role merging in session events
The EventsAdapter.All() method SHALL merge consecutive same-role events as a defense-in-depth measure to prevent turn-order violations at the ADK/provider boundary. Parts from consecutive same-role events SHALL be concatenated into a single event.

#### Scenario: Consecutive assistant events merged
- **WHEN** the event history contains two consecutive events with role "model"
- **THEN** EventsAdapter.All() SHALL yield a single event with both events' Parts concatenated

#### Scenario: Alternating roles preserved
- **WHEN** the event history contains alternating user and model events
- **THEN** EventsAdapter.All() SHALL yield each event separately without merging

#### Scenario: Len() consistent with All()
- **WHEN** consecutive same-role events exist in history
- **THEN** EventsAdapter.Len() SHALL return the count of merged events (matching All() output), not the raw history count

### Requirement: Gemini message builder preserves ThoughtSignature
When constructing `genai.Content` for Gemini API requests, the message builder SHALL set `Thought` and `ThoughtSignature` on `genai.Part` from the corresponding `provider.ToolCall` fields.

#### Scenario: Reconstruct FunctionCall with ThoughtSignature for API request
- **WHEN** a `provider.Message` with ToolCalls containing `ThoughtSignature` is converted to `genai.Content`
- **THEN** the resulting `genai.Part` SHALL have `Thought` and `ThoughtSignature` fields set to match the ToolCall values

#### Scenario: Session history replay preserves ThoughtSignature
- **WHEN** session history is converted to ADK events via `EventsAdapter`
- **THEN** FunctionCall `genai.Part` instances SHALL include `Thought` and `ThoughtSignature` from the stored `session.ToolCall`

### Requirement: Gemini thought text emits observable event
The Gemini provider SHALL emit a `StreamEventThought` event for text parts with `Thought=true` instead of silently discarding them. The event SHALL carry `ThoughtLen` (byte length of the thought text) but SHALL NOT include the thought text content.

#### Scenario: Thought-only text part emits thought event
- **WHEN** a Gemini streaming response contains a text part with `Thought=true`
- **THEN** the provider SHALL yield a `StreamEvent` with `Type: StreamEventThought` and `ThoughtLen` equal to `len(part.Text)`

#### Scenario: Non-thought text part unchanged
- **WHEN** a Gemini streaming response contains a text part with `Thought=false`
- **THEN** the provider SHALL yield a `StreamEvent` with `Type: StreamEventPlainText` and `Text` set to the part text

#### Scenario: Mixed thought and visible text parts
- **WHEN** a Gemini response contains both thought parts and visible text parts
- **THEN** the provider SHALL yield `StreamEventThought` for thought parts and `StreamEventPlainText` for visible text parts in order

### Requirement: ModelAdapter handles thought events
The `ModelAdapter` SHALL handle `StreamEventThought` as a no-op in both streaming and non-streaming paths. Thought events SHALL NOT contribute to accumulated text or tool call parts.

#### Scenario: Streaming mode thought event ignored
- **WHEN** a `StreamEventThought` is received in streaming mode
- **THEN** the `ModelAdapter` SHALL not yield any `LLMResponse` and SHALL not modify the accumulated text builder

#### Scenario: Non-streaming mode thought event ignored
- **WHEN** a `StreamEventThought` is received in non-streaming mode
- **THEN** the `ModelAdapter` SHALL not modify the text accumulator or tool parts

### Requirement: ModelAdapter propagates ThoughtSignature bidirectionally
The `ModelAdapter` SHALL propagate `Thought` and `ThoughtSignature` from `provider.ToolCall` to `genai.Part` in both streaming and non-streaming paths, and from `genai.Part` to `provider.ToolCall` in `convertMessages`.

#### Scenario: Streaming path preserves ThoughtSignature
- **WHEN** a streaming ToolCall event arrives with `ThoughtSignature`
- **THEN** the `genai.Part` yielded in `LLMResponse` SHALL include the `Thought` and `ThoughtSignature` fields

#### Scenario: convertMessages extracts ThoughtSignature
- **WHEN** `convertMessages` processes a `genai.Content` with FunctionCall parts carrying `ThoughtSignature`
- **THEN** the resulting `provider.ToolCall` SHALL include the `Thought` and `ThoughtSignature` values
