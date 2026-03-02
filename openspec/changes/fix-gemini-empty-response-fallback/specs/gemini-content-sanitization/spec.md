## ADDED Requirements

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
